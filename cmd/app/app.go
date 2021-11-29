/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"log"
	"os"
	"runtime"

	"github.com/mysteriumnetwork/myst-launcher/app"
	"github.com/mysteriumnetwork/myst-launcher/const"
	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	ipc_ "github.com/mysteriumnetwork/myst-launcher/ipc"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const (
	gitHubOrg  = "mysteriumnetwork"
	gitHubRepo = "myst-launcher"
)

func main() {
	defer utils.PanicHandler("main")
	runtime.LockOSThread()
	utils.Win32Initialize()

	ap := app.NewApp()
	ipc := ipc_.NewHandler()

	if len(os.Args) > 1 {
		ap.InTray = os.Args[1] == _const.FlagTray
		ap.InstallStage2 = os.Args[1] == _const.FlagInstallStage2

		switch os.Args[1] {
		case _const.FlagInstall:
			utils.InstallExe()
			return

		case _const.FlagUninstall:
			ipc.SendStopApp()
			utils.UninstallExe()
			return
		}
	}

	mod := model.NewUIModel()
	mod.SetApp(ap)
	mod.DuplicateLogToConsole = true

	prodVersion, _ := utils.GetProductVersion()
	mod.SetProductVersion(prodVersion)

	gui_win32.InitGDIPlus()
	ui := gui_win32.NewGui(mod)
	if updateLauncherFromNewBinary(ui, ipc) {
		return
	}

	ui.CreateNotifyIcon(mod)
	ui.CreateMainWindow()

	ap.SetModel(mod)
	ap.SetUI(ui)
	ap.WaitGroup.Add(1)

	go ap.SuperviseDockerNode()
	go ap.CheckLauncherUpdates(gitHubOrg, gitHubRepo)

	ipc.Listen(ui)

	log.SetOutput(ap)

	// Run the message loop
	ui.Run()
	gui_win32.ShutdownGDIPlus()

	// send stop action to SuperviseDockerNode
	ap.TriggerAction("stop")
	ap.Shutdown()
}

// return: bool exit
func updateLauncherFromNewBinary(ui *gui_win32.Gui, p *ipc_.Handler) bool {
	if utils.LauncherUpgradeAvailable() {
		update := ui.YesNoModal("Mysterium launcher upgrade", "You are running a newer version of launcher.\r\nUpgrade launcher installation ?")
		if model.IDYES == update {
			if !p.OwnsPipe() {
				p.SendStopApp()
				p.OpenPipe()
			}
			utils.UpdateExe()
			return false
		}
	}

	if !p.OwnsPipe() {
		p.SendPopupApp()
		return true
	}
	return false
}
