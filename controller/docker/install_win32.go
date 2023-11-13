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
	"os"
	"syscall"

	"github.com/gonutz/w32"
	"github.com/lxn/win"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"github.com/mysteriumnetwork/myst-launcher/utils"
	"github.com/winlabs/gowin32"
	"golang.org/x/sys/windows"
)

const dockerUsersGroup = "docker-users"

func (c *Docker_) TryInstallRuntime_() {

	if c.model.Config.InitialState.Not1Not2() {
		c.model.SwitchState(model.UIStateInstallNeeded)

	} else {
		// begin install immediately
		c.TryInstallRuntime()
	}
}

// returns: will exit
func (c *Docker_) TryInstallRuntime() bool {
	fmt.Println("TryInstallRuntime >")
	mdl := c.model
	ui := c.ui

	if !w32.SHIsUserAnAdmin() {
		if mdl.Config.InitialState.Not1Not2() {
			mdl.Config.InitialState = model.InitialStateStage1
			mdl.Config.Save()
		}
		utils.RunasWithArgsNoWait("")

		// force immediate exit
		c.model.UIBus.Publish("dlg-exit")
		return true
	}

	mdl.ResetProperties()
	mdl.SwitchState(model.UIStateInstallInProgress)

	executor := NewStepExecutor(mdl)
	executor.AddStep("CheckWindowsVersion", func() bool {
		c.lg.Println("Checking Windows version")

		if !utils.IsWindowsVersionCompatible() {
			c.lg.Println("You must run Windows 10 version 2004 or above.")
			ui.ConfirmModal("Installation", "Please update to Windows 10 version 2004 or above.")
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
				c.lg.Printf("Downloading %d of %d: %s", fi+1, len(list), v.name)

				if _, err := os.Stat(utils.GetTmpDir() + "\\" + v.name); err != nil {
					err := utils.DownloadFile(utils.GetTmpDir()+"\\"+v.name, v.url, func(progress int) {
						if progress%10 == 0 {
							c.lg.Printf("%s - %d%%\n", v.name, progress)
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
			c.lg.Println("Download failed")
			ret := ui.YesNoModal("Download failed", "Retry download?")
			if ret == win.IDYES {
				continue
			}
			return false
		}
		return true
	})

	InstallWSLUpdate := func() bool {
		mdl.UpdateProperties(model.UIProps{"InstallWSLUpdate": model.StepInProgress})
		c.lg.Println("Installing wsl_update_x64.msi")

		gowin32.SetInstallerInternalUI(gowin32.InstallUILevelProgressOnly) // UI Level for a prompt
		wslIsUpdated, err := utils.IsWSLUpdated()
		if err != nil {
			c.lg.Println("IsWSLUpdated err>", err)
			return false
		}

		if !wslIsUpdated {
			err = gowin32.InstallProduct(utils.GetTmpDir()+"\\wsl_update_x64.msi", "ACTION=INSTALL")
			if err != nil {
				c.lg.Println("InstallProduct err>", err)
				return false
			}
		} else {
			c.lg.Println("WSL is already updated!")
		}
		return true
	}

	executor.AddStep("InstallWSLUpdate", InstallWSLUpdate)
	executor.AddStep("InstallDocker", func() bool {
		c.lg.Println("Installing docker desktop (wait ~5 minutes)")

		exe := utils.GetTmpDir() + "\\DockerDesktopInstaller.exe"
		err := native.ShellExecuteAndWait(0, "runas", exe, "install --quiet", utils.GetTmpDir(), syscall.SW_NORMAL)
		if err != nil {
			c.lg.Println("DockerDesktopInstaller failed:", err)
			return false
		}
		if err := startDockerDesktop(); err != nil {
			c.lg.Println("Failed starting docker:", err)
			return false
		}
		return true
	})
	executor.AddStep("InstallWSLUpdate", InstallWSLUpdate)

	executor.AddStep("CheckGroupMembership", func() bool {
		mdl.Config.InitialState = model.InitialStateFirstRunAfterInstall
		mdl.Config.AutoStart = true
		mdl.Config.Save()

		c.lg.Println("Checking current docker-users group membership")
		if !utils.CurrentGroupMembership(dockerUsersGroup) {

			c.lg.Println("Sign out from the current session to finish the installation.")
			ret := ui.ConfirmModal("Installation", "Click yes to sign out from the current session to finish the installation.")
			if ret == win.IDOK {
				windows.ExitWindowsEx(windows.EWX_LOGOFF, 0)
				return true
			}
			c.lg.Println("Remember to sign out from the current session")
			return true
		}
		return true
	})

	if !executor.Run() {
		mdl.SwitchState(model.UIStateInstallError)
		c.lg.Println("Installation have stopped")
		return false
	}
	// TODO: unelevate rights
	utils.DiscoverDockerPathAndPatchEnv(true)
	mdl.SwitchState(model.UIStateInstallFinished)
	c.lg.Println("Installation succeeded")
	return true
}
