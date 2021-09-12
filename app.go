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

	"github.com/mysteriumnetwork/myst-launcher/app"
	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/gui"
	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

func main() {
	ap := app.NewApp()

	if len(os.Args) > 1 {
		ap.InTray = os.Args[1] == _const.FlagTray
		ap.InstallStage2 = os.Args[1] == _const.FlagInstallStage2

		switch os.Args[1] {
		case _const.FlagInstall:
			utils.InstallExe()
			return

		case _const.FlagUninstall:
			app.StopApp()
			utils.UninstallExe()
			return
		}
	}
	// Upgrade binary on start
	//if utils.LauncherUpgradeAvailable() {
	//	fmt.Println("LauncherUpgradeAvailable !")
	//}

	if app.IsAlreadyRunning() {
		return
	}
	log.SetOutput(ap)

	mod := gui.NewUIModel()
	mod.SetApp(ap)

	ui := gui_win32.NewGui(mod)
	ui.CreateNotifyIcon(mod)
	ui.CreateMainWindow()

	ap.SetModel(mod)
	ap.SetUI(ui)
	ap.WaitGroup.Add(1)

	go ap.SuperviseDockerNode()
	app.CreatePipeAndListen(ui)

	// Run the message loop
	ui.Run()

	// send stop action to SuperviseDockerNode
	ap.TriggerAction("stop")

	// wait for SuperviseDockerNode to finish its work
	ap.WaitGroup.Wait()
}
