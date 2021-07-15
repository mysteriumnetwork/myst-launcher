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

	"github.com/lxn/walk"
)

func main() {
	if len(os.Args) > 1 {
		gui.UI.InTray = os.Args[1] == app.FlagTray
		gui.UI.InstallStage2 = os.Args[1] == app.FlagInstallStage2

		if os.Args[1] == app.FlagInstall {
			app.InstallExe()
			return
		}
	}

	if utils.IsAlreadyRunning() {
		return
	}
	utils.CreatePipeAndListen(&gui.UI)

	log.SetOutput(&gui.UI)
	gui.UI.Icon, _ = walk.NewIconFromResourceId(2)
	gui.CreateDialogue()

	go func() {
		app.SuperviseDockerNode()
	}()
	app.CreateNotifyIcon()
}
