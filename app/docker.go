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
	"runtime"
	"syscall"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/lxn/win"
	"github.com/winlabs/gowin32"
	"golang.org/x/sys/windows"

	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const group = "docker-users"

func (s *AppState) SuperviseDockerNode() {
	runtime.LockOSThread()
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	defer s.WaitGroup.Done()

	if utils.LauncherUpgradeAvailable() {
		ret := gui.UI.YesNoModal("Launcher upgrade", "You are running a newer version of launcher.\r\n\r\nUpgrade launcher installation ?")
		if ret == win.IDYES {
			exePath, _ := os.Executable()
			err := native.ShellExecuteAndWait(0, "runas", exePath, FlagInstall, "", syscall.SW_NORMAL)
			if err != nil {
				log.Println("Failed to install exe", err)
			}
		}
	}

	mystManager, err := myst.NewManagerWithDefaults()
	if err != nil {
		panic(err) // TODO handle gracefully
	}
	docker := myst.NewDockerMonitor(mystManager)

	t1 := time.Tick(15 * time.Second)
	s.Config.Read()
	gui.UI.Update()

	for {
		tryStartOrInstallDocker := func() bool {
			if isRunning, _ := docker.IsRunning(); isRunning {
				gui.UI.SetStateDocker(gui.RunnableStateRunning)
				return false
			}

			// In case of suspend/resume some APIs may return unexpected error, so we need to retry it
			isUnderVM, needSetup, err := false, false, error(nil)
			err = utils.Retry(3, time.Second, func() error {
				isUnderVM, err = utils.SystemUnderVm()
				if err != nil {
					return err
				}
				features, err := utils.QueryFeatures()
				if err != nil {
					return err
				}
				if len(features) > 0 {
					needSetup = true
					return nil
				}
				return nil
			})
			if err != nil {
				log.Println("error", err)
				gui.UI.ErrorModal("Application error", err.Error())
				return true
			}

			if isUnderVM && !s.Config.CheckVMSettingsConfirm {
				ret := gui.UI.YesNoModal("Requirements checker", "VM has been detected. \r\n\r\nPlease ensure that VT-x / EPT / IOMMU \r\nare enabled for this VM.\r\nRefer to VM settings.\r\n\r\nContinue ?")
				if ret == win.IDNO {
					gui.UI.ExitApp()
					return true
				}
				s.Config.CheckVMSettingsConfirm = true
				s.Config.Save()
			}
			if needSetup {
				return s.tryInstall()
			}

			isRunning, couldNotStart := docker.IsRunning()
			if isRunning {
				gui.UI.SetStateDocker(gui.RunnableStateRunning)
				return false
			}
			gui.UI.SetStateDocker(gui.RunnableStateStarting)
			if couldNotStart {
				gui.UI.SetStateDocker(gui.RunnableStateUnknown)
				return s.tryInstall()
			}

			return false
		}
		wantExit := tryStartOrInstallDocker()
		if wantExit {
			gui.UI.SetWantExit()
			return
		}

		// docker is running now
		s.upgradeImageAndRun(mystManager)

		select {
		case act := <-s.Action:
			switch act {
			case "check":
				s.ImgVer.CurrentImgDigest = mystManager.GetCurrentImageDigest()
				ok := myst.CheckVersionAndUpgrades(&s.ImgVer, &s.Config)
				if !ok {
					break
				}
				s.Config.Save()
				gui.UI.Update()

			case "upgrade":
				s.upgrade(mystManager)

			case "enable":
				s.Config.Enabled = true
				s.Config.Save()
				gui.UI.SetStateContainer(gui.RunnableStateRunning)
				mystManager.Start(s.GetConfig())

			case "disable":
				s.Config.Enabled = false
				s.Config.Save()
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
	if !utils.IsWindowsVersionCompatible() {
		gui.UI.SwitchState(gui.ModalStateInstallError)

		log.Println("You must run Windows 10 version 2004 or above.")
		gui.UI.ConfirmModal("Installation", "Please update to Windows 10 version 2004 or above.")

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	gui.UI.CheckWindowsVersion = true
	gui.UI.Update()

	if !s.InstallStage2 {
		log.Println("Install executable")
		fullExe, _ := os.Executable()
		cmdArgs := FlagInstall
		err = native.ShellExecuteAndWait(0, "runas", fullExe, cmdArgs, "", syscall.SW_NORMAL)
		if err != nil {
			log.Println("Failed to install executable")

			gui.UI.SwitchState(gui.ModalStateInstallError)
			return true
		}
		utils.CreateAutostartShortcut(FlagInstallStage2)
		gui.UI.InstallExecutable = true
		gui.UI.Update()
	}

	// Don't check VT-x / EPT as it's just enough to check VMPlatform WSL and vmcompute

	features, err := utils.QueryFeatures()
	if err != nil {
		log.Println("Failed to query feature:", utils.FeatureWSL)
		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	err = utils.InstallFeatures(features, func(feature int, name string) {
		switch feature {
		case utils.IDFeatureWSL:
			gui.UI.EnableWSL = true
		case utils.IDFeatureHyperV:
			gui.UI.EnableHyperV = true
		}
		gui.UI.Update()
	})
	if err != nil {
		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}

	if len(features) > 0 {
		ret := gui.UI.YesNoModal("Installation", "Reboot is required to enable Windows optional feature\r\n"+"Click Yes to reboot now")
		if ret == win.IDYES {
			native.ShellExecuteNowait(0, "", "shutdown", "-r", "", syscall.SW_NORMAL)
		}
		return true
	}

	// proceeding install after reboot
	if s.InstallStage2 {
		gui.UI.InstallExecutable = true
		gui.UI.RebootAfterWSLEnable = true
		gui.UI.Update()
	}

	// Instead of chechking VT-x check vmcompute service is running
	log.Println("Checking vmcompute (Hyper-V Host Compute Service)")
	isVMComputeRunning := utils.IsVMComputeRunning()
	if !isVMComputeRunning {
		log.Println("Vmcompute (Hyper-V Host Compute Service) is not running")
		gui.UI.ConfirmModal("Installation", "Vmcompute (Hyper-V Host Compute Service)")

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	gui.UI.CheckVTx = true
	gui.UI.Update()

	utils.CreateAutostartShortcut(FlagTray)
	utils.CreateDesktopShortcut("")
	utils.CreateStartMenuShortcut("")

	download := func() error {
		list := []struct{ url, name string }{
			{"https://desktop.docker.com/win/stable/amd64/Docker%20Desktop%20Installer.exe", "DockerDesktopInstaller.exe"},
			{"https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi", "wsl_update_x64.msi"},
		}
		for fi, v := range list {
			log.Println(fmt.Sprintf("Downloading %d of %d: %s", fi+1, len(list), v.name))
			if _, err := os.Stat(os.Getenv("TMP") + "\\" + v.name); err != nil {

				err := utils.DownloadFile(os.Getenv("TMP")+"\\"+v.name, v.url, func(progress int) {
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
	gowin32.SetInstallerInternalUI(gowin32.InstallUILevelProgressOnly) // UI Level for a prompt
	err = gowin32.InstallProduct(os.Getenv("TMP")+"\\wsl_update_x64.msi", "ACTION=INSTALL")
	if err != nil {
		log.Println("Command failed: msiexec.exe /i wsl_update_x64.msi /quiet")

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	gui.UI.InstallWSLUpdate = true
	gui.UI.Update()

	log.Println("Installing docker desktop (wait ~5 minutes)")
	exe := os.Getenv("TMP") + "\\DockerDesktopInstaller.exe"
	err = native.ShellExecuteAndWait(0, "runas", exe, "install --quiet", os.Getenv("TMP"), syscall.SW_NORMAL)
	if err != nil {
		log.Println("DockerDesktopInstaller failed:", err)

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	if err := myst.StartDockerDesktop(); err != nil {
		log.Println("Failed starting docker:", err)

		gui.UI.SwitchState(gui.ModalStateInstallError)
		return true
	}
	gui.UI.InstallDocker = true
	gui.UI.Update()

	log.Println("Checking current group membership")
	if !utils.CurrentGroupMembership(group) {
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

	s.Config.Read()
	s.Config.AutoStart = true
	s.Config.Save()
	log.Println("Installation succeeded")
	s.didInstallation = true

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
	mystManager.Update(s.GetConfig())

	s.ImgVer.CurrentImgDigest = mystManager.GetCurrentImageDigest()
	ok := myst.CheckVersionAndUpgrades(&s.ImgVer, &s.Config)
	if ok {
		s.Config.Save()
	}
	gui.UI.Update()
}

// check for image updates before starting container, offer upgrade interactively
func (s *AppState) upgradeImageAndRun(mystManager *myst.Manager) {
	imageDigest := mystManager.GetCurrentImageDigest()

	if s.ImgVer.CurrentImgDigest != imageDigest || s.ImgVer.VersionCurrent == "" || s.Config.NeedToCheckUpgrade() {
		// docker has a new image (in result the of external command)

		s.ImgVer.CurrentImgDigest = imageDigest
		ok := myst.CheckVersionAndUpgrades(&s.ImgVer, &s.Config)
		if ok {
			s.Config.Save()
		}
		gui.UI.Update()
	}
	if s.Config.AutoUpgrade && s.ImgVer.HasUpdate {
		s.upgrade(mystManager)
	}

	if s.Config.Enabled {
		gui.UI.SetStateContainer(gui.RunnableStateUnknown)
		containerAlreadyRunning, err := mystManager.Start(s.GetConfig())
		if err != nil {
			return
		}

		gui.UI.SetStateContainer(gui.RunnableStateRunning)
		gui.UI.Update()

		if !containerAlreadyRunning && s.didInstallation {
			s.didInstallation = false

			gui.UI.LastNotificationID = gui.NotificationContainerStarted
			gui.UI.ShowNotificationInstalled()
		}
	}
}
