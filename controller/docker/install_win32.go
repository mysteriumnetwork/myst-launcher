//go:build windows
// +build windows

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package docker

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/gonutz/w32"
	"github.com/lxn/win"
	"github.com/winlabs/gowin32"
	"golang.org/x/sys/windows"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"github.com/mysteriumnetwork/myst-launcher/platform"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const dockerUsersGroup = "docker-users"

// returns: will exit
func (s *Controller) tryInstallDocker() bool {
	mgr := s.mgr.(*platform.Manager)
	mdl := s.a.GetModel()
	ui := s.a.GetUI()
	// action := s.a.GetAction()

	mdl.ResetProperties()
	mdl.SwitchState(model.UIStateInstallNeeded)

	if mdl.Config.InitialState != model.InitialStateStage1 && mdl.Config.InitialState != model.InitialStateStage2 {
		ok := ui.WaitDialogueComplete()
		if !ok {
			return true
		}
	}

	if !w32.SHIsUserAnAdmin() {
		if mdl.Config.InitialState != model.InitialStateStage1 && mdl.Config.InitialState != model.InitialStateStage2 {
			mdl.Config.InitialState = model.InitialStateStage1
			mdl.Config.Save()
		}
		utils.RunasWithArgsNoWait("")
		return true
	}

	mdl.SwitchState(model.UIStateInstallInProgress)

	executor := StepExec{s.a.GetModel(), nil}
	executor.AddStep("CheckWindowsVersion", func() bool {
		log.Println("Checking Windows version")
		if !utils.IsWindowsVersionCompatible() {
			log.Println("You must run Windows 10 version 2004 or above.")
			ui.ConfirmModal("Installation", "Please update to Windows 10 version 2004 or above.")
			return false
		}
		return true
	})
	executor.AddStep("InstallExecutable", func() bool {
		if mdl.Config.InitialState != model.InitialStateStage2 {
			log.Println("Install executable")
			if err := utils.CheckAndInstallExe(); err != nil {
				log.Println("Failed to install executable")
				return false
			}

			mdl.Config.InitialState = model.InitialStateStage2
			mdl.Config.Save()

			utils.CreateAutostartShortcut("")
			utils.CreateDesktopShortcut("")
			utils.CreateStartMenuShortcut("")
		}
		return true
	})

	executor.AddStep("CheckVTx", func() bool {
		// Don't check VT-x / EPT as it's just enough to check VMPlatform WSL and vmcompute
		ok, err := s.mgr.Features()
		if err != nil {
			log.Println(err)
			return false
		}
		if !ok {
			mgr.EnableHyperVPlatform()

			ret := ui.YesNoModal("Installation", "Reboot is required to enable Windows optional feature\r\n"+"Click Yes to reboot now")
			if ret == win.IDYES {
				native.ShellExecuteNowait(0, "", "shutdown", "-r", "", syscall.SW_NORMAL)
			}
			return false
		}

		// proceeding install after reboot
		mdl.UpdateProperties(model.UIProps{"RebootAfterWSLEnable": model.StepFinished})

		log.Println("Checking vmcompute (Hyper-V Host Compute Service)")
		ok, err = mgr.IsVMcomputeRunning()
		if err != nil {
			log.Println(err)
			return false
		}
		// force service to start
		if !ok {
			ok, _ = mgr.StartVmcomputeIfNotRunning()
		}
		if !ok {
			log.Println("Vmcompute (Hyper-V Host Compute Service) is not running")
			ui.ConfirmModal("Installation", "Vmcompute (Hyper-V Host Compute Service) is not running.\r\n\r\n"+
				"Please enable virtualization in a system BIOS: VT-x and EPT options for Intel, SVM for AMD")

			return false
		}

		return true
	})
	executor.AddStep("DownloadFiles", func() bool {
		mdl.UpdateProperties(model.UIProps{"DownloadFiles": model.StepInProgress})
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
			if err := download(); err == nil {
				break
			}
			log.Println("Download failed")
			ret := ui.YesNoModal("Download failed", "Retry download?")
			if ret == win.IDYES {
				continue
			}
			return false
		}
		return true
	})
	executor.AddStep("InstallWSLUpdate", func() bool {
		mdl.UpdateProperties(model.UIProps{"InstallWSLUpdate": model.StepInProgress})
		log.Println("Installing wsl_update_x64.msi")

		gowin32.SetInstallerInternalUI(gowin32.InstallUILevelProgressOnly) // UI Level for a prompt
		wslIsUpdated, err := utils.IsWSLUpdated()
		if err != nil {
			log.Println("IsWSLUpdated err>", err)
			return false
		}

		if !wslIsUpdated {
			err = gowin32.InstallProduct(utils.GetTmpDir()+"\\wsl_update_x64.msi", "ACTION=INSTALL")
			if err != nil {
				log.Println("InstallProduct err>", err)
				return false
			}
		} else {
			log.Println("WSL is already updated!")
		}
		return true
	})

	executor.AddStep("InstallDocker", func() bool {
		log.Println("Installing docker desktop (wait ~5 minutes)")

		exe := utils.GetTmpDir() + "\\DockerDesktopInstaller.exe"
		err := native.ShellExecuteAndWait(0, "runas", exe, "install --quiet", utils.GetTmpDir(), syscall.SW_NORMAL)
		if err != nil {
			log.Println("DockerDesktopInstaller failed:", err)
			return false
		}
		if err := StartDockerDesktop(); err != nil {
			log.Println("Failed starting docker:", err)
			return false
		}
		return true
	})
	executor.AddStep("InstallWSLUpdate", func() bool {
		mdl.UpdateProperties(model.UIProps{"InstallWSLUpdate": model.StepInProgress})
		log.Println("Installing wsl_update_x64.msi")

		gowin32.SetInstallerInternalUI(gowin32.InstallUILevelProgressOnly) // UI Level for a prompt
		wslIsUpdated, err := utils.IsWSLUpdated()
		if err != nil {
			log.Println("IsWSLUpdated err>", err)
			return false
		}

		if !wslIsUpdated {
			err = gowin32.InstallProduct(utils.GetTmpDir()+"\\wsl_update_x64.msi", "ACTION=INSTALL")
			if err != nil {
				log.Println("InstallProduct err>", err)
				return false
			}
		} else {
			log.Println("WSL is already updated!")
		}
		return true
	})

	executor.AddStep("CheckGroupMembership", func() bool {
		mdl.Config.InitialState = model.InitialStateFirstRunAfterInstall
		mdl.Config.AutoStart = true
		mdl.Config.Save()

		log.Println("Checking current docker-users group membership")
		if !utils.CurrentGroupMembership(dockerUsersGroup) {

			log.Println("Sign out from the current session to finish the installation.")
			ret := ui.ConfirmModal("Installation", "Click yes to sign out from the current session to finish the installation.")
			if ret == win.IDOK {
				windows.ExitWindowsEx(windows.EWX_LOGOFF, 0)
				return true
			}
			log.Println("Remember to sign out from the current session")
			return true
		}
		return true
	})

	if !executor.Run() {
		mdl.SwitchState(model.UIStateInstallError)
		log.Println("Installation have failed")
		return true
	}

	// TODO: unelevate rights

	utils.DiscoverDockerPathAndPatchEnv(true)
	mdl.SwitchState(model.UIStateInstallFinished)
	log.Println("Installation succeeded")

	ok := ui.WaitDialogueComplete()
	if !ok {
		return true
	}
	return false
}
