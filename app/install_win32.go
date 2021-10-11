//go:build windows
// +build windows

package app

import (
	"fmt"
	"log"
	"os"
	"syscall"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"github.com/mysteriumnetwork/myst-launcher/utils"

	"github.com/lxn/win"
	"github.com/winlabs/gowin32"
	"golang.org/x/sys/windows"
)

const group = "docker-users"

// returns exit model: true means exit
func (s *AppState) tryInstall() bool {
	var err error

	s.model.SwitchState(model.UIStateInstallNeeded)
	if !s.InstallStage2 {
		ok := s.ui.WaitDialogueComplete()
		if !ok {
			return true
		}
	}
	s.model.SwitchState(model.UIStateInstallInProgress)

	log.Println("Checking Windows version")
	if !utils.IsWindowsVersionCompatible() {
		s.model.SwitchState(model.UIStateInstallError)

		log.Println("You must run Windows 10 version 2004 or above.")

		s.ui.ConfirmModal("Installation", "Please update to Windows 10 version 2004 or above.")
		s.model.SwitchState(model.UIStateInstallError)

		return true
	}
	s.model.UpdateProperties(model.UIProps{"CheckWindowsVersion": true})

	if !s.InstallStage2 {
		log.Println("Install executable")
		if err := utils.CheckAndInstallExe(); err != nil {
			log.Println("Failed to install executable")
			
			s.model.SwitchState(model.UIStateInstallError)
			return true
		}
		utils.CreateAutostartShortcut(_const.FlagInstallStage2)

		s.model.UpdateProperties(model.UIProps{"InstallExecutable": true})
	}

	// Don't check VT-x / EPT as it's just enough to check VMPlatform WSL and vmcompute

	features, err := utils.QueryFeatures()
	if err != nil {
		log.Println("Failed to query feature:", err)
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	err = utils.InstallFeatures(features, nil)
	if err != nil {
		s.model.SwitchState(model.UIStateInstallError)
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
		s.model.UpdateProperties(model.UIProps{"InstallExecutable": true, "RebootAfterWSLEnable": true})
	}

	// Instead of chechking VT-x check vmcompute service is running
	log.Println("Checking vmcompute (Hyper-V Host Compute Service)")
	isVMComputeRunning := utils.IsVMComputeRunning()
	if !isVMComputeRunning {
		log.Println("Vmcompute (Hyper-V Host Compute Service) is not running")

		s.ui.ConfirmModal("Installation", "Vmcompute (Hyper-V Host Compute Service) is not running.\r\n\r\n"+
			"Please enable virtualization in a system BIOS: VT-x and EPT options for Intel, SVM for AMD")
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	s.model.UpdateProperties(model.UIProps{"CheckVTx": true})

	utils.CreateAutostartShortcut(_const.FlagTray)
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

		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	s.model.UpdateProperties(model.UIProps{"DownloadFiles": true})

	log.Println("Installing wsl_update_x64.msi")
	gowin32.SetInstallerInternalUI(gowin32.InstallUILevelProgressOnly) // UI Level for a prompt
	err = gowin32.InstallProduct(os.Getenv("TMP")+"\\wsl_update_x64.msi", "ACTION=INSTALL")
	if err != nil {
		log.Println("Command failed: msiexec.exe /i wsl_update_x64.msi /quiet")

		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	s.model.UpdateProperties(model.UIProps{"InstallWSLUpdate": true})

	log.Println("Installing docker desktop (wait ~5 minutes)")
	exe := os.Getenv("TMP") + "\\DockerDesktopInstaller.exe"
	err = native.ShellExecuteAndWait(0, "runas", exe, "install --quiet", os.Getenv("TMP"), syscall.SW_NORMAL)
	if err != nil {
		log.Println("DockerDesktopInstaller failed:", err)

		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	if err := myst.StartDockerDesktop(); err != nil {
		log.Println("Failed starting docker:", err)

		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	s.model.UpdateProperties(model.UIProps{"InstallDocker": true})

	log.Println("Checking current group membership")
	if !utils.CurrentGroupMembership(group) {
		// request a logout //
		log.Println("Log of from the current session to finish the installation.")

		ret := s.ui.ConfirmModal("Installation", "Log of from the current session to finish the installation.")
		if ret == win.IDYES {
			windows.ExitWindowsEx(windows.EWX_LOGOFF, 0)
		}

		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	s.model.UpdateProperties(model.UIProps{"CheckGroupMembership": true})
	s.model.Config.AutoStart = true
	s.model.Config.Save()

	log.Println("Installation succeeded")
	s.didInstallation = true

	s.model.SwitchState(model.UIStateInstallFinished)
	ok := s.ui.WaitDialogueComplete()
	if !ok {
		return true
	}
	s.model.SwitchState(model.UIStateInitial)
	return false
}
