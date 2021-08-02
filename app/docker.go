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

func (s *AppState) SuperviseDockerNode() {
	runtime.LockOSThread()
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	defer s.WaitGroup.Done()

	mystManager, err := myst.NewManagerWithDefaults()
	if err != nil {
		panic(err) // TODO handle gracefully
	}

	t1 := time.Tick(15 * time.Second)
	tryStartCount := 0
	didDockerInstall := false

	s.ReadConfig()
	gui.UI.Update()

	for {
		tryStartOrInstallDocker := func() bool {
			// In case of suspend/resume some APIs may return unexpected error, so we need to retry it
			checkVMSettings, needSetup, err := false, false, error(nil)
			err = Retry(3, time.Second, func() error {
				checkVMSettings, err = isUnderVm()
				if err != nil {
					return err
				}
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
				log.Println("error", err)
				gui.UI.ErrorModal("Application error", err.Error())
				return true
			}
			if checkVMSettings && !s.Config.CheckVMSettingsConfirm {
				ret := gui.UI.YesNoModal("Requirements checker", "VM has been detected. \r\n\r\nPlease ensure that VT-x / EPT / IOMMU \r\nare enabled for this VM.\r\nRefer to VM settings.\r\n\r\nContinue ?")
				if ret == win.IDNO {
					gui.UI.ExitApp()
					return true
				}
				s.Config.CheckVMSettingsConfirm = true
				s.SaveConfig()
			}
			if needSetup {
				didDockerInstall = true
				return s.tryInstall()
			}

			canPingDocker := mystManager.CanPingDocker()
			if !canPingDocker {
				tryStartCount++

				// try starting docker for 10 times, else try install
				if !tryStartDocker() || tryStartCount == 10 {
					tryStartCount = 0
					didDockerInstall = true
					return s.tryInstall()
				}
			}
			return false
		}
		wantExit := tryStartOrInstallDocker()
		if wantExit {
			gui.UI.SetWantExit()
			return
		}
		gui.UI.SetStateDocker(gui.RunnableStateRunning)

		// docker is running now
		// check for image updates before starting container, offer upgrade interactively
		// start container
		func() {
			imageDigest := mystManager.GetCurrentImageDigest()

			if gui.UI.CurrentImgDigest != imageDigest || gui.UI.VersionCurrent == "" || s.Config.NeedToCheckUpgrade() {
				gui.UI.CurrentImgDigest = imageDigest

				ok := myst.CheckVersionAndUpgrades(imageDigest, &s.Config)
				if ok {
					s.SaveConfig()
				}
			}
			if s.Config.AutoUpgrade && !gui.UI.VersionUpToDate {
				s.upgrade(mystManager)
			}

			if s.Config.Enabled {
				gui.UI.SetStateContainer(gui.RunnableStateUnknown)
				containerAlreadyRunning, err := mystManager.Start()
				if err != nil {
					return
				}
				gui.UI.SetStateContainer(gui.RunnableStateRunning)
				gui.UI.Update()

				if !containerAlreadyRunning && didDockerInstall {
					didDockerInstall = false
					gui.UI.LastNotificationID = gui.NotificationContainerStarted
					gui.UI.ShowNotificationInstalled()
				}
			}
		}()

		select {
		case act := <-s.Action:
			switch act {
			case "check":
				ok := myst.CheckVersionAndUpgrades(mystManager.GetCurrentImageDigest(), &s.Config)
				if !ok {
					break
				}
				s.SaveConfig()
				gui.UI.Update()

			case "upgrade":
				s.upgrade(mystManager)

			case "enable":
				s.Config.Enabled = true
				s.SaveConfig()
				gui.UI.SetStateContainer(gui.RunnableStateRunning)
				mystManager.Start()

			case "disable":
				s.Config.Enabled = false
				s.SaveConfig()
				gui.UI.SetStateContainer(gui.RunnableStateUnknown)
				mystManager.Stop()

			case "stop":
				fmt.Println("stop")
				return
			}

		// wait for ticker event if no action
		case <-t1:
		}
	}
}

func tryStartDocker() bool {
	gui.UI.SetStateContainer(gui.RunnableStateUnknown)
	gui.UI.SetStateDocker(gui.RunnableStateStarting)

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

// returns exit model: true means exit
func (s *AppState) tryInstall() bool {
	var err error

	gui.UI.SwitchState(gui.ModalStateInstallNeeded)
	if !s.InstallStage2 {
		ok := gui.UI.WaitDialogueComplete()
		if !ok {
			return true
		}
	}
	gui.UI.SwitchState(gui.ModalStateInstallInProgress)

	log.Println("Checking Windows version")
	if !IsWindowsVersionCompatible() {
		gui.UI.SwitchState(gui.ModalStateInstallError)

		log.Println("You must run Windows 10 version 2004 or above.")
		gui.UI.ConfirmModal("Installation", "Please update to Windows 10 version 2004 or above.")

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	gui.UI.CheckWindowsVersion = true
	gui.UI.Update()

	log.Println("Checking VT-x / EPT")
	if !hasVTx() {
		log.Println("Please Enable virtualization in BIOS")
		gui.UI.ConfirmModal("Installation", "Please Enable virtualization in BIOS / Hypervisor: VT-x and EPT (Intel), SVM (AMD)")

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	gui.UI.CheckVTx = true
	gui.UI.Update()

	if !s.InstallStage2 {
		_, wslEnabled, err := QueryWindowsFeature(FeatureWSL)
		if err != nil {
			log.Println("Failed to get model: Microsoft-Windows-Subsystem-Linux")
			gui.UI.SwitchState(gui.ModalStateInstallError)
			return true
		}
		hyperVExists, hyperVEnabled, err := QueryWindowsFeature(FeatureHyperV)
		if err != nil {
			log.Println("Failed to get model: Microsoft-Hyper-V")
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
	CreateAutostartShortcut(FlagTray)
	CreateDesktopShortcut("")
	CreateStartMenuShortcut("")

	download := func() error {
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
					return err
				}
			}
		}
		return nil
	}

	for {
		if err = download(); err == nil {
			break
		}
		log.Println("Download failed")
		ret := gui.UI.YesNoModal("Download failed", "Retry download?")
		if ret == win.IDYES {
			continue
		}

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
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

	s.ReadConfig()
	s.Config.AutoStart = true
	s.SaveConfig()
	log.Println("Installation succeeded")

	gui.UI.SwitchState(gui.ModalStateInstallFinished)
	ok := gui.UI.WaitDialogueComplete()
	if !ok {
		return true
	}

	gui.UI.SwitchState(gui.ModalStateInitial)
	return false
}

func (s *AppState) upgrade(mystManager *myst.Manager) {
	gui.UI.SetStateContainer(gui.RunnableStateUnknown)
	mystManager.Stop()
	gui.UI.SetStateContainer(gui.RunnableStateInstalling)
	mystManager.Update()

	gui.UI.CurrentImgDigest = mystManager.GetCurrentImageDigest()
	ok := myst.CheckVersionAndUpgrades(gui.UI.CurrentImgDigest, &s.Config)
	if ok {
		s.SaveConfig()
	}
}
