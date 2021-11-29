//go:build windows
// +build windows

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package app

import (
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"log"
	"os"
	"syscall"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"github.com/mysteriumnetwork/myst-launcher/utils"

	"github.com/lxn/win"
	"github.com/winlabs/gowin32"
	"golang.org/x/sys/windows"
)

const group = "docker-users"

// returns exit model: true means exit
func (s *AppState) tryInstallDocker() bool {
	s.model.ResetProperties()

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
	s.model.UpdateProperties(model.UIProps{"CheckWindowsVersion": model.StepInProgress})
	if !utils.IsWindowsVersionCompatible() {
		s.model.SwitchState(model.UIStateInstallError)

		log.Println("You must run Windows 10 version 2004 or above.")

		s.ui.ConfirmModal("Installation", "Please update to Windows 10 version 2004 or above.")
		s.model.SwitchState(model.UIStateInstallError)
		s.model.UpdateProperties(model.UIProps{"CheckWindowsVersion": model.StepFailed})
		return true
	}
	s.model.UpdateProperties(model.UIProps{"CheckWindowsVersion": model.StepFinished})

	s.model.UpdateProperties(model.UIProps{"InstallExecutable": model.StepInProgress})
	if !s.InstallStage2 {
		log.Println("Install executable")
		if err := utils.CheckAndInstallExe(); err != nil {
			log.Println("Failed to install executable")

			s.model.SwitchState(model.UIStateInstallError)
			return true
		}
		utils.CreateAutostartShortcut(_const.FlagInstallStage2)
	}
	s.model.UpdateProperties(model.UIProps{"InstallExecutable": model.StepFinished})

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
	s.model.UpdateProperties(model.UIProps{"RebootAfterWSLEnable": model.StepFinished})

	// Instead of chechking VT-x check vmcompute service is running
	s.model.UpdateProperties(model.UIProps{"CheckVTx": model.StepInProgress})
	log.Println("Checking vmcompute (Hyper-V Host Compute Service)")
	isVMComputeRunning := utils.IsVMComputeRunning()
	if !isVMComputeRunning {
		log.Println("Vmcompute (Hyper-V Host Compute Service) is not running")

		s.ui.ConfirmModal("Installation", "Vmcompute (Hyper-V Host Compute Service) is not running.\r\n\r\n"+
			"Please enable virtualization in a system BIOS: VT-x and EPT options for Intel, SVM for AMD")
		s.model.SwitchState(model.UIStateInstallError)
		s.model.UpdateProperties(model.UIProps{"CheckVTx": model.StepFailed})
		return true
	}
	s.model.UpdateProperties(model.UIProps{"CheckVTx": model.StepFinished})

	utils.CreateAutostartShortcut(_const.FlagTray)
	utils.CreateDesktopShortcut("")
	utils.CreateStartMenuShortcut("")

	s.model.UpdateProperties(model.UIProps{"DownloadFiles": model.StepInProgress})
	download := func() error {
		list := []struct{ url, name string }{
			{"https://desktop.docker.com/win/stable/amd64/Docker%20Desktop%20Installer.exe", "DockerDesktopInstaller.exe"},
			{"https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi", "wsl_update_x64.msi"},
		}
		for fi, v := range list {
			log.Println(fmt.Sprintf("Downloading %d of %d: %s", fi+1, len(list), v.name))
			if _, err := os.Stat(utils.GetTmpDir() + "\\" + v.name); err != nil {

				err := utils.DownloadFile(utils.GetTmpDir()+"\\"+v.name, v.url, func(progress int) {
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
		s.model.UpdateProperties(model.UIProps{"DownloadFiles": model.StepFailed})
		return true
	}
	s.model.UpdateProperties(model.UIProps{"DownloadFiles": model.StepFinished})

	s.model.UpdateProperties(model.UIProps{"InstallWSLUpdate": model.StepInProgress})
	log.Println("Installing wsl_update_x64.msi")

	gowin32.SetInstallerInternalUI(gowin32.InstallUILevelProgressOnly) // UI Level for a prompt
	wslIsUpdated, err := IsWSLUpdated()
	if err != nil {
		log.Println("InstallProduct failed (wsl_update_x64.msi)", err)

		s.model.SwitchState(model.UIStateInstallError)
		s.model.UpdateProperties(model.UIProps{"InstallWSLUpdate": model.StepFailed})
		return true
	}
	if !wslIsUpdated {
		err = gowin32.InstallProduct(utils.GetTmpDir()+"\\wsl_update_x64.msi", "ACTION=INSTALL")
		if err != nil {
			log.Println("InstallProduct failed (wsl_update_x64.msi)", err)

			s.model.SwitchState(model.UIStateInstallError)
			s.model.UpdateProperties(model.UIProps{"InstallWSLUpdate": model.StepFailed})
			return true
		}
	} else {
		log.Println("WSL is already updated!")
	}
	s.model.UpdateProperties(model.UIProps{"InstallWSLUpdate": model.StepFinished})

	s.model.UpdateProperties(model.UIProps{"InstallDocker": model.StepInProgress})
	log.Println("Installing docker desktop (wait ~5 minutes)")
	exe := utils.GetTmpDir() + "\\DockerDesktopInstaller.exe"
	err = native.ShellExecuteAndWait(0, "runas", exe, "install --quiet", utils.GetTmpDir(), syscall.SW_NORMAL)
	if err != nil {
		log.Println("DockerDesktopInstaller failed:", err)

		s.model.SwitchState(model.UIStateInstallError)
		s.model.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	if err := StartDockerDesktop(); err != nil {
		log.Println("Failed starting docker:", err)

		s.model.SwitchState(model.UIStateInstallError)
		s.model.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	s.model.UpdateProperties(model.UIProps{"InstallDocker": model.StepFinished})

	s.model.UpdateProperties(model.UIProps{"CheckGroupMembership": model.StepInProgress})
	log.Println("Checking current group membership")
	if !utils.CurrentGroupMembership(group) {
		// request a logout //
		log.Println("Log of from the current session to finish the installation.")

		ret := s.ui.ConfirmModal("Installation", "Log of from the current session to finish the installation.")
		if ret == win.IDYES {
			windows.ExitWindowsEx(windows.EWX_LOGOFF, 0)
		}

		s.model.SwitchState(model.UIStateInstallError)
		s.model.UpdateProperties(model.UIProps{"CheckGroupMembership": model.StepFailed})
		return true
	}
	s.model.UpdateProperties(model.UIProps{"CheckGroupMembership": model.StepFinished})
	s.model.Config.AutoStart = true
	s.model.Config.Save()

	log.Println("Installation succeeded")
	s.didDockerInstallation = true

	s.model.SwitchState(model.UIStateInstallFinished)
	ok := s.ui.WaitDialogueComplete()
	if !ok {
		return true
	}
	s.model.SwitchState(model.UIStateInitial)
	return false
}

func IsWSLUpdated() (bool, error) {
	const WSLUpdateProductCode = "{36EF257E-21D5-44F7-8451-07923A8C465E}"
	state := gowin32.GetInstalledProductState(WSLUpdateProductCode)
	if state != gowin32.InstallStateDefault {
		return false, nil
	}

	installedVer, err := gowin32.GetInstalledProductProperty(WSLUpdateProductCode, gowin32.InstallPropertyVersionString)
	if err != nil {
		return false, wrap(err, errors.New("gowin32.GetInstalledProductProperty"))
	}

	pkg, err := gowin32.OpenInstallerPackage(utils.GetTmpDir() + "\\wsl_update_x64.msi")
	if err != nil {
		return false, wrap(err, errors.New("gowin32.OpenInstallerPackage"))
	}
	defer pkg.Close()

	fileVer, err := pkg.GetProductProperty("ProductVersion")
	if err != nil {
		return false, wrap(err, errors.New("gowin32.GetProductProperty"))
	}
	semverFileVer, err := semver.Parse(fileVer)
	if err != nil {
		return false, wrap(err, errors.New("semver.Parse"))
	}
	semverInstalledVer, err := semver.Parse(installedVer)
	if err != nil {
		return false, wrap(err, errors.New("semver.Parse"))
	}

	// semverInstalledVer >= semverFileVer
	return semverInstalledVer.Compare(semverFileVer) >= 0, nil
}

func wrap(external, internal error) error {
	return fmt.Errorf(external.Error()+": %w", internal)
}
