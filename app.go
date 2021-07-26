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
	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

func main() {
	a := app.NewApp()

	if len(os.Args) > 1 {
		a.InTray = os.Args[1] == app.FlagTray
		a.InstallStage2 = os.Args[1] == app.FlagInstallStage2

		switch os.Args[1] {
		case app.FlagInstall:
			app.InstallExe()
			return

		case app.FlagUninstall:
			app.UninstallExe()
			return
		}
	}

	if utils.IsAlreadyRunning() {
		return
	}

	log.SetOutput(a)
	a.ReadConfig()
	a.ImageName = myst.GetImageName()

	gui.UI.SetApp(a)
	gui.CreateNotifyIcon()
	gui.CreateDialogue()

	a.WaitGroup.Add(1)
	go a.SuperviseDockerNode()

	utils.CreatePipeAndListen(&gui.UI)

	// Run the message loop
	gui.UI.Run()

	// send stop action to SuperviseDockerNode
	a.TriggerAction("stop")

	// wait for SuperviseDockerNode to finish its work
	a.WaitGroup.Wait()
}
