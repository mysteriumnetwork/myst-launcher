// +build windows

package app

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"github.com/mysteriumnetwork/myst-launcher/utils"

	"github.com/lxn/win"
	"github.com/winlabs/gowin32"
	"golang.org/x/sys/windows"
)

// returns exit model: true means exit
func (s *AppState) tryInstall() bool {
	var err error

	s.mod.SwitchState(gui.ModalStateInstallNeeded)
	if !s.InstallStage2 {
		ok := s.mod.WaitDialogueComplete()
		if !ok {
			return true
		}
	}
	s.mod.SwitchState(gui.ModalStateInstallInProgress)

	log.Println("Checking Windows version")
	if !utils.IsWindowsVersionCompatible() {
		s.mod.SwitchState(gui.ModalStateInstallError)

		log.Println("You must run Windows 10 version 2004 or above.")

		s.ui.ConfirmModal("Installation", "Please update to Windows 10 version 2004 or above.")
		s.mod.SwitchState(gui.ModalStateInstallError)

		return true
	}
	s.mod.UpdateProperties(gui.UIProps{"CheckWindowsVersion": true})

	if !s.InstallStage2 {
		log.Println("Install executable")
		fullExe, _ := os.Executable()
		cmdArgs := FlagInstall
		err = native.ShellExecuteAndWait(0, "runas", fullExe, cmdArgs, "", syscall.SW_NORMAL)
		if err != nil {
			log.Println("Failed to install executable")

			s.mod.SwitchState(gui.ModalStateInstallError)
			return true
		}
		utils.CreateAutostartShortcut(FlagInstallStage2)

		s.mod.UpdateProperties(gui.UIProps{"InstallExecutable": true})
	}

	// Don't check VT-x / EPT as it's just enough to check VMPlatform WSL and vmcompute

	features, err := utils.QueryFeatures()
	if err != nil {
		log.Println("Failed to query feature:", utils.FeatureWSL)
		s.mod.SwitchState(gui.ModalStateInstallError)
		return true
	}
	err = utils.InstallFeatures(features, func(feature int, name string) {

		pp := gui.UIProps{}
		switch feature {
		case utils.IDFeatureWSL:
			pp["EnableWSL"] = true
		case utils.IDFeatureHyperV:
			pp["EnableHyperV"] = true
		}
		s.mod.UpdateProperties(pp)
	})
	if err != nil {
		s.mod.SwitchState(gui.ModalStateInstallError)
		return true
	}

	if len(features) > 0 {
		ret := s.ui.YesNoModal("Installation", "Reboot is required to enable Windows optional feature\r\n"+"Click Yes to reboot now")
		if ret == win.IDYES {
			native.ShellExecuteNowait(0, "", "shutdown", "-r", "", syscall.SW_NORMAL)
		}
		return true
	}

	// proceeding install after reboot
	if s.InstallStage2 {
		s.mod.UpdateProperties(gui.UIProps{"InstallExecutable": true, "RebootAfterWSLEnable": true})
	}

	// Instead of chechking VT-x check vmcompute service is running
	log.Println("Checking vmcompute (Hyper-V Host Compute Service)")
	isVMComputeRunning := utils.IsVMComputeRunning()
	if !isVMComputeRunning {
		log.Println("Vmcompute (Hyper-V Host Compute Service) is not running")

		s.ui.ConfirmModal("Installation", "Vmcompute (Hyper-V Host Compute Service)")
		s.mod.SwitchState(gui.ModalStateInstallError)
		return true
	}
	s.mod.UpdateProperties(gui.UIProps{"CheckVTx": true})

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
		ret := s.ui.YesNoModal("Download failed", "Retry download?")
		if ret == win.IDYES {
			continue
		}

		s.mod.SwitchState(gui.ModalStateInstallError)
		return true
	}
	s.mod.UpdateProperties(gui.UIProps{"DownloadFiles": true})

	log.Println("Installing wsl_update_x64.msi")
	gowin32.SetInstallerInternalUI(gowin32.InstallUILevelProgressOnly) // UI Level for a prompt
	err = gowin32.InstallProduct(os.Getenv("TMP")+"\\wsl_update_x64.msi", "ACTION=INSTALL")
	if err != nil {
		log.Println("Command failed: msiexec.exe /i wsl_update_x64.msi /quiet")

		s.mod.SwitchState(gui.ModalStateInstallError)
		return true
	}
	s.mod.UpdateProperties(gui.UIProps{"InstallWSLUpdate": true})

	log.Println("Installing docker desktop (wait ~5 minutes)")
	exe := os.Getenv("TMP") + "\\DockerDesktopInstaller.exe"
	err = native.ShellExecuteAndWait(0, "runas", exe, "install --quiet", os.Getenv("TMP"), syscall.SW_NORMAL)
	if err != nil {
		log.Println("DockerDesktopInstaller failed:", err)

		s.mod.SwitchState(gui.ModalStateInstallError)
		return true
	}
	if err := myst.StartDockerDesktop(); err != nil {
		log.Println("Failed starting docker:", err)

		s.mod.SwitchState(gui.ModalStateInstallError)
		return true
	}
	s.mod.UpdateProperties(gui.UIProps{"InstallDocker": true})

	log.Println("Checking current group membership")
	if !utils.CurrentGroupMembership(group) {
		// request a logout //
		log.Println("Log of from the current session to finish the installation.")

		ret := s.ui.ConfirmModal("Installation", "Log of from the current session to finish the installation.")
		if ret == win.IDYES {
			windows.ExitWindowsEx(windows.EWX_LOGOFF, 0)
		}

		s.mod.SwitchState(gui.ModalStateInstallError)
		return true
	}
	s.mod.UpdateProperties(gui.UIProps{"CheckGroupMembership": true})

	s.Config.Read()
	s.Config.AutoStart = true
	s.Config.Save()

	log.Println("Installation succeeded")
	s.didInstallation = true

	s.mod.SwitchState(gui.ModalStateInstallFinished)
	ok := s.mod.WaitDialogueComplete()
	if !ok {
		return true
	}

	s.mod.SwitchState(gui.ModalStateInitial)
	return false
}
