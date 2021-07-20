/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"

	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/native"
)

const group = "docker-users"

func SuperviseDockerNode() {
	runtime.LockOSThread()
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)

	mystManager, err := myst.NewManagerWithDefaults()
	if err != nil {
		panic(err) // TODO handle gracefully
	}

	t1 := time.Tick(10 * time.Second)

	for {
		isWLSEnabled, err := isWLSEnabled()
		if err != nil {
			gui.UI.Bus.Publish("show-dlg", "error", err)
			gui.UI.ExitApp()
		}
		if !isWLSEnabled {
			tryInstall(isWLSEnabled)
			continue
		}

		canPingDocker := mystManager.CanPingDocker()
		if canPingDocker {
			gui.UI.StateDocker = gui.RunnableStateRunning
			gui.UI.Update()

			err := mystManager.Start()
			if err != nil {
				gui.UI.Bus.Publish("show-dlg", "error", err)
				gui.UI.ExitApp()
			}
			gui.UI.StateContainer = gui.RunnableStateRunning
			gui.UI.Update()

			id := mystManager.GetCurrentImageDigest()
			myst.CheckUpdates(id)

		} else {
			tryInstall(isWLSEnabled)
		}

		select {
		case <-gui.UI.UpgradeClick:
			id := mystManager.GetCurrentImageDigest()
			myst.CheckUpdates(id)

			if gui.UI.VersionUpToDate {
				gui.UI.Bus.Publish("show-dlg", "is-up-to-date", nil)
				return
			}
			mystManager.Stop()
			mystManager.Update()

		case <-t1:
			break
		}
	}
}

func maybeDockerIsTurnedOff() bool {
	gui.UI.StateDocker = gui.RunnableStateStarting
	gui.UI.StateContainer = gui.RunnableStateUnknown
	gui.UI.Update()

	if isProcessRunning("Docker Desktop.exe") {
		return false
	}
	if err := startDocker(); err != nil {
		log.Printf("Failed to start cmd: %v", err)
		return false
	}
	return true
}

func startDocker() error {
	dd := os.Getenv("ProgramFiles") + "\\Docker\\Docker\\Docker Desktop.exe"
	cmd := exec.Command(dd, "-Autostart")
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start cmd: %v", err)
		return err
	}
	return nil
}

func tryInstall(isWLSEnabled bool) {
	var err error

	if maybeDockerIsTurnedOff() {
		return
	}

	if !gui.UI.InstallStage2 {
		gui.UI.SwitchState(gui.ModalStateInstallNeeded)
		gui.UI.WaitDialogueComplete()
	}
	gui.UI.SwitchState(gui.ModalStateInstallInProgress)

	log.Println("Checking Windows version")
	gui.UI.Update()

	if !IsWindowsVersionCompatible() {
		gui.UI.SwitchState(gui.ModalStateInstallError)

		log.Println("You must run Windows 10 version 2004 or above.")
		gui.UI.ConfirmModal("Installation", "Please update to Windows 10 version 2004 or above.")

		gui.UI.WaitDialogueComplete()
		gui.UI.ExitApp()
		return
	}
	gui.UI.CheckWindowsVersion = true
	gui.UI.Update()

	log.Println("Checking VT-x")
	if !hasVTx() {
		gui.UI.ConfirmModal("Installation", "Please Enable virtualization in BIOS")

		gui.UI.WaitDialogueComplete()
		gui.UI.ExitApp()
		return
	}
	gui.UI.CheckVTx = true
	gui.UI.Update()

	if !gui.UI.InstallStage2 {
		if !isWLSEnabled {
			log.Println("Enable WSL..")
			exe := "dism.exe"
			cmdArgs := "/online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart"
			err = native.ShellExecuteAndWait(0, "runas", exe, cmdArgs, "", syscall.SW_HIDE)
			if err != nil {
				log.Println("Command failed: failed to enable Microsoft-Windows-Subsystem-Linux")
				gui.UI.SwitchState(gui.ModalStateInstallError)

				gui.UI.WaitDialogueComplete()
				gui.UI.ExitApp()
				return
			}
		}
		gui.UI.EnableWSL = true
		gui.UI.Update()

		log.Println("Install executable")
		fullExe, _ := os.Executable()
		cmdArgs := FlagInstall
		err = native.ShellExecuteAndWait(0, "runas", fullExe, cmdArgs, "", syscall.SW_NORMAL)
		if err != nil {
			log.Println("Failed to install executable")
			gui.UI.SwitchState(gui.ModalStateInstallError)

			gui.UI.WaitDialogueComplete()
			gui.UI.ExitApp()
			return
		}
		CreateAutostartShortcut(FlagInstallStage2)
		gui.UI.InstallExecutable = true
		gui.UI.Update()

		if !isWLSEnabled {
			ret := gui.UI.ConfirmModal("Installation", "Reboot is needed to finish installation of WSL\r\nClick OK to reboot")
			if ret == win.IDOK {
				native.ShellExecuteAndWait(0, "", "shutdown", "-r", "", syscall.SW_NORMAL)
			}
			gui.UI.WaitDialogueComplete()
			gui.UI.ExitApp()
			return
		}
	} else {
		// proceeding install after reboot
		gui.UI.EnableWSL = true
		gui.UI.InstallExecutable = true
		gui.UI.RebootAfterWSLEnable = true
		gui.UI.Update()
	}
	CreateAutostartShortcut("")
	CreateDesktopShortcut("")
	CreateStartMenuShortcut("")

	list := []struct{ url, name string }{
		{"https://desktop.docker.com/win/stable/amd64/Docker%20Desktop%20Installer.exe", "DockerDesktopInstaller.exe"},
		{"https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi", "wsl_update_x64.msi"},
	}
	for fi, v := range list {
		log.Println(fmt.Sprintf("Downloading %d of %d: %s", fi+1, len(list), v.name))
		if _, err := os.Stat(os.Getenv("TMP") + "\\" + v.name); err != nil {

			err := DownloadFile(os.Getenv("TMP")+"\\"+v.name, v.url, func(ignored int) {

			})
			if err != nil {
				log.Println("Download failed")
				gui.UI.SwitchState(gui.ModalStateInstallError)

				gui.UI.WaitDialogueComplete()
				gui.UI.ExitApp()
				return
			}
		}
	}
	gui.UI.DownloadFiles = true

	log.Println("Installing wsl_update_x64.msi")
	exe := "msiexec.exe"
	cmdArgs := "/i " + os.Getenv("TMP") + "\\wsl_update_x64.msi /quiet"
	err = native.ShellExecuteAndWait(0, "runas", exe, cmdArgs, os.Getenv("TMP"), syscall.SW_NORMAL)
	if err != nil {
		log.Println("Command failed: msiexec.exe /i wsl_update_x64.msi /quiet")
		gui.UI.SwitchState(gui.ModalStateInstallError)

		gui.UI.WaitDialogueComplete()
		gui.UI.ExitApp()
		return
	}
	gui.UI.InstallWSLUpdate = true
	gui.UI.Update()

	log.Println("Installing docker desktop")
	exe = os.Getenv("TMP") + "\\DockerDesktopInstaller.exe"
	err = native.ShellExecuteAndWait(0, "runas", exe, "install --quiet", os.Getenv("TMP"), syscall.SW_NORMAL)
	if err != nil {
		log.Println("DockerDesktopInstaller failed:", err)
		gui.UI.SwitchState(gui.ModalStateInstallError)

		gui.UI.WaitDialogueComplete()
		gui.UI.ExitApp()
		return
	}
	if err := startDocker(); err != nil {
		log.Println("Failed starting docker:", err)
		gui.UI.SwitchState(gui.ModalStateInstallError)

		gui.UI.WaitDialogueComplete()
		gui.UI.ExitApp()
		return
	}
	gui.UI.InstallDocker = true
	gui.UI.Update()

	log.Println("Checking current group membership")
	if !CurrentGroupMembership(group) {
		// request a logout //

		ret := gui.UI.ConfirmModal("Installation", "Log of from the current session to finish the installation.")
		if ret == win.IDYES {
			windows.ExitWindowsEx(windows.EWX_LOGOFF, 0)
			return
		}
		log.Println("Log of from the current session to finish the installation.")
		gui.UI.SwitchState(gui.ModalStateInstallError)

		gui.UI.WaitDialogueComplete()
		gui.UI.ExitApp()
		return
	}
	gui.UI.CheckGroupMembership = true
	gui.UI.Update()

	gui.UI.ReadConfig()
	gui.UI.CFG.AutoStart = true
	gui.UI.SaveConfig()
	log.Println("Installation succeeded")

	gui.UI.SwitchState(gui.ModalStateInstallFinished)
	gui.UI.WaitDialogueComplete()
	gui.UI.SwitchState(gui.ModalStateInitial)
}
