/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/mysteriumnetwork/myst-launcher/app"
	_const "github.com/mysteriumnetwork/myst-launcher/const"
	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

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

	if app.IsAlreadyRunning() {
		return
	}

	mod := model.NewUIModel()
	mod.SetApp(ap)
	mod.GetConfig().DuplicateLogToConsole = true

	ui := gui_win32.NewGui(mod)
	ui.CreateNotifyIcon(mod)
	ui.CreateMainWindow()

	ap.SetModel(mod)
	ap.SetUI(ui)
	ap.WaitGroup.Add(1)

	go ap.SuperviseDockerNode()
	app.CreatePipeAndListen(ui)

	log.SetOutput(ap)

	// Run the message loop
	ui.Run()

	// send stop action to SuperviseDockerNode
	ap.TriggerAction("stop")

	// wait for SuperviseDockerNode to finish its work
	ap.WaitGroup.Wait()
}
