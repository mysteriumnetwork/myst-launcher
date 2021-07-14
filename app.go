/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

//go:generate goversioninfo -icon=ico/icon_512x512.ico -manifest=launcher.exe.manifest
package main

import (
	"github.com/mysteriumnetwork/myst-launcher/app"
	"github.com/mysteriumnetwork/myst-launcher/utils"
	"log"
	"os"

	"github.com/lxn/walk"
)

func main() {
	if len(os.Args) > 1 {
		app.SModel.InTray = os.Args[1] == app.FlagTray
		app.SModel.InstallStage2 = os.Args[1] == app.FlagInstallStage2

		if os.Args[1] == app.FlagInstall {
				app.InstallExe()
			return
		}
	}

	if utils.IsAlreadyRunning() {
		return
	}
	utils.CreatePipeAndListen(&app.SModel)

	log.SetOutput(&app.SModel)
	app.SModel.Icon, _ = walk.NewIconFromResourceId(2)
	app.CreateDialogue()

	go func() {
		app.SuperviseDockerNode()
	}()
	app.CreateNotifyIcon()
}
