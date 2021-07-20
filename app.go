/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

//go:generate goversioninfo -icon=ico/icon_512x512.ico -manifest=launcher.exe.manifest

package main

import (
	"log"
	"os"

	"github.com/mysteriumnetwork/myst-launcher/app"
	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

func main() {
	if len(os.Args) > 1 {
		gui.UI.InTray = os.Args[1] == app.FlagTray
		gui.UI.InstallStage2 = os.Args[1] == app.FlagInstallStage2

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

	log.SetOutput(&gui.UI)
	gui.CreateNotifyIcon()
	gui.CreateDialogue()

	gui.UI.WaitGroup.Add(1)
	go app.SuperviseDockerNode()

	utils.CreatePipeAndListen(&gui.UI)

	// Run the message loop
	gui.UI.Run()
}
