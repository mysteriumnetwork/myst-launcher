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
	defer gui.UI.WaitGroup.Done()

	var err error
	mystManager, err := myst.NewManagerWithDefaults()
	if err != nil {
		panic(err) // TODO handle gracefully
	}

	t1 := time.Tick(15 * time.Second)
	tryStartCount := 0
	didInstall := false
	countStarted := 0

	for {
		tryStartOrInstall := func() bool {
			// In case of suspend/resume some APIs may return unexpected error, so we need to retry it
			needSetup, err := false, error(nil)
			Retry(3, time.Second, func() error {
				_, wslEnabled, err := QueryWindowsFeature(FeatureWSL)
				if err != nil {
					return err
				}
				if !wslEnabled {
					needSetup = true
					return nil
				}
				hyperVExists, hyperVEnabled, err := QueryWindowsFeature(FeatureHyperV)
				if err != nil {
					return err
				}
				if hyperVExists && !hyperVEnabled {
					needSetup = true
					return nil
				}
				return err
			})
			if err != nil {
				gui.UI.Bus.Publish("show-dlg", "error", err)
				return true
			}
			if needSetup {
				didInstall = true
				return tryInstall()
			}

			canPingDocker := mystManager.CanPingDocker()
			if canPingDocker {
				gui.UI.StateDocker = gui.RunnableStateRunning
				gui.UI.Update()

				if gui.UI.CFG.Enabled {
					gui.UI.StateContainer = gui.RunnableStateInstalling
					gui.UI.Update()

					alreadyRunning, err := mystManager.Start()
					if err != nil {
						return false
					}
					if !alreadyRunning {
						countStarted = 0
					}

					gui.UI.StateContainer = gui.RunnableStateRunning
					gui.UI.Update()
					if !alreadyRunning && didInstall {
						didInstall = false
						gui.UI.ShowNotification()
					}

					if countStarted == 0 {
						id := mystManager.GetCurrentImageDigest()
						myst.CheckUpdates(id)
					}
					countStarted++
				}

			} else {
				tryStartCount++
				started := tryStartDocker()

				// try starting docker for 10 times, else try install
				if !started || tryStartCount == 10 {
					tryStartCount = 0
					didInstall = true
					return tryInstall()
				}
			}
			return false
		}
		wantExit := tryStartOrInstall()
		if wantExit {
			gui.UI.SetWantExit()
			return
		}

		select {
		case act := <-gui.UI.UIAction:
			switch act {
			case "upgrade":
				id := mystManager.GetCurrentImageDigest()
				myst.CheckUpdates(id)

				if gui.UI.VersionUpToDate {
					gui.UI.Bus.Publish("show-dlg", "is-up-to-date", nil)
					return
				}
				gui.UI.StateContainer = gui.RunnableStateUnknown
				gui.UI.Update()
				mystManager.Stop()

				gui.UI.StateContainer = gui.RunnableStateInstalling
				gui.UI.Update()
				mystManager.Update()

				id = mystManager.GetCurrentImageDigest()
				myst.CheckUpdates(id)

			case "enable":
				gui.UI.StateContainer = gui.RunnableStateRunning
				gui.UI.Update()
				mystManager.Start()
				gui.UI.CFG.Enabled = true
				gui.UI.SaveConfig()

			case "disable":
				gui.UI.StateContainer = gui.RunnableStateUnknown
				gui.UI.Update()
				mystManager.Stop()
				gui.UI.CFG.Enabled = false
				gui.UI.SaveConfig()

			case "open-ui":

			case "stop":
				return
			}

		// wait for ticker event if no action
		case <-t1:
		}
	}
}

func tryStartDocker() bool {
	gui.UI.StateDocker = gui.RunnableStateStarting
	gui.UI.StateContainer = gui.RunnableStateUnknown
	gui.UI.Update()

	if isProcessRunning("Docker Desktop.exe") {
		return true
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

// returns exit state: true means exit
func tryInstall() bool {
	var err error

	gui.UI.SwitchState(gui.ModalStateInstallNeeded)
	if !gui.UI.InstallStage2 {
		ok := gui.UI.WaitDialogueComplete()
		if !ok {
			return true
		}
	}
	gui.UI.SwitchState(gui.ModalStateInstallInProgress)

	log.Println("Checking Windows version")
	gui.UI.Update()

	if !IsWindowsVersionCompatible() {
		gui.UI.SwitchState(gui.ModalStateInstallError)

		log.Println("You must run Windows 10 version 2004 or above.")
		gui.UI.ConfirmModal("Installation", "Please update to Windows 10 version 2004 or above.")

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	gui.UI.CheckWindowsVersion = true
	gui.UI.Update()

	log.Println("Checking VT-x")
	if !hasVTx() {
		log.Println("Please Enable virtualization in BIOS")
		gui.UI.ConfirmModal("Installation", "Please Enable virtualization in BIOS")

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	gui.UI.CheckVTx = true
	gui.UI.Update()

	if !gui.UI.InstallStage2 {
		_, wslEnabled, err := QueryWindowsFeature(FeatureWSL)
		if err != nil {
			log.Println("Failed to get state: Microsoft-Windows-Subsystem-Linux")
			gui.UI.SwitchState(gui.ModalStateInstallError)
			return true
		}
		hyperVExists, hyperVEnabled, err := QueryWindowsFeature(FeatureHyperV)
		if err != nil {
			log.Println("Failed to get state: Microsoft-Hyper-V")
			gui.UI.SwitchState(gui.ModalStateInstallError)
			return true
		}
		needEnableHyperV := hyperVExists && !hyperVEnabled

		if !wslEnabled {
			log.Println("Enable WSL..")
			exe := "dism.exe"
			cmdArgs := "/online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart"
			err = native.ShellExecuteAndWait(0, "runas", exe, cmdArgs, "", syscall.SW_HIDE)
			if err != nil {
				log.Println("Command failed: failed to enable Microsoft-Windows-Subsystem-Linux")

				gui.UI.SwitchState(gui.ModalStateInstallError)
				return true
			}
		}
		gui.UI.EnableWSL = true
		gui.UI.Update()

		if needEnableHyperV {
			log.Println("Enable Hyper-V..")
			exe := "dism.exe"
			cmdArgs := "/online /enable-feature /featurename:Microsoft-Hyper-V /all /norestart"
			err = native.ShellExecuteAndWait(0, "runas", exe, cmdArgs, "", syscall.SW_HIDE)
			if err != nil {
				log.Println("Command failed: failed to enable Microsoft-Hyper-V")

				gui.UI.SwitchState(gui.ModalStateInstallError)
				return true
			}
		}
		gui.UI.EnableHyperV = true
		gui.UI.Update()

		log.Println("Install executable")
		fullExe, _ := os.Executable()
		cmdArgs := FlagInstall
		err = native.ShellExecuteAndWait(0, "runas", fullExe, cmdArgs, "", syscall.SW_NORMAL)
		if err != nil {
			log.Println("Failed to install executable")

			gui.UI.SwitchState(gui.ModalStateInstallError)
			return true
		}
		CreateAutostartShortcut(FlagInstallStage2)
		gui.UI.InstallExecutable = true
		gui.UI.Update()

		if !wslEnabled || needEnableHyperV {
			ret := gui.UI.YesNoModal("Installation", "Reboot is required to enable Windows optional feature\r\n"+
				"Click Yes to reboot now")
			if ret == win.IDYES {
				native.ShellExecuteNowait(0, "", "shutdown", "-r", "", syscall.SW_NORMAL)
			}
			return true
		}
	} else {
		// proceeding install after reboot
		gui.UI.EnableWSL = true
		gui.UI.EnableHyperV = true
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

			err := DownloadFile(os.Getenv("TMP")+"\\"+v.name, v.url, func(progress int) {
				if progress%10 == 0 {
					log.Println(fmt.Sprintf("%s - %d%%", v.name, progress))
				}
			})
			if err != nil {
				log.Println("Download failed")

				gui.UI.SwitchState(gui.ModalStateInstallError)
				return true
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
		return true
	}
	gui.UI.InstallWSLUpdate = true
	gui.UI.Update()

	log.Println("Installing docker desktop (wait ~5 minutes)")
	exe = os.Getenv("TMP") + "\\DockerDesktopInstaller.exe"
	err = native.ShellExecuteAndWait(0, "runas", exe, "install --quiet", os.Getenv("TMP"), syscall.SW_NORMAL)
	if err != nil {
		log.Println("DockerDesktopInstaller failed:", err)

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	if err := startDocker(); err != nil {
		log.Println("Failed starting docker:", err)

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	gui.UI.InstallDocker = true
	gui.UI.Update()

	log.Println("Checking current group membership")
	if !CurrentGroupMembership(group) {
		// request a logout //
		log.Println("Log of from the current session to finish the installation.")

		ret := gui.UI.ConfirmModal("Installation", "Log of from the current session to finish the installation.")
		if ret == win.IDYES {
			windows.ExitWindowsEx(windows.EWX_LOGOFF, 0)
		}

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	gui.UI.CheckGroupMembership = true
	gui.UI.Update()

	gui.UI.ReadConfig()
	gui.UI.CFG.AutoStart = true
	gui.UI.SaveConfig()
	log.Println("Installation succeeded")

	gui.UI.SwitchState(gui.ModalStateInstallFinished)

	ok := gui.UI.WaitDialogueComplete()
	if !ok {
		return true
	}

	gui.UI.SwitchState(gui.ModalStateInitial)
	return false
}
