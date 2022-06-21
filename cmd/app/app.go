/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/gonutz/w32"

	"github.com/mysteriumnetwork/myst-launcher/app"
	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/controller"
	"github.com/mysteriumnetwork/myst-launcher/controller/docker"
	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	ipc_ "github.com/mysteriumnetwork/myst-launcher/ipc"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/updates"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const (
	gitHubOrg  = "mysteriumnetwork"
	gitHubRepo = "myst-launcher"
)

var debugMode = ""

func main() {
	defer utils.PanicHandler("main")

	ap := app.NewApp()
	ipc := ipc_.NewHandler()

	if len(os.Args) > 1 {
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

	// in case of installation restart elevated
	if mod.Config.InitialState == model.InitialStateStage1 || mod.Config.InitialState == model.InitialStateStage2 {
		if !w32.SHIsUserAnAdmin() {
			utils.RunasWithArgsNoWait("")
			return
		}
	}

	prodVersion, _ := utils.GetProductVersion()
	mod.SetProductVersion(prodVersion)

	gui_win32.InitGDIPlus()
	ui := gui_win32.NewGui(mod)

	// skip update if IsUserAnAdmin
	if !w32.SHIsUserAnAdmin() && updates.UpdateLauncherFromNewBinary(ui, ipc) {
		return
	}
	ap.SetModel(mod)

	ui.CreateNotifyIcon(mod)
	ui.CreateMainWindow()

	ap.SetUI(ui)

	dc := docker.NewController(ap)
	go dc.SuperviseDockerNode()
	go controller.CheckLauncherUpdates(gitHubOrg, gitHubRepo, ap.GetModel())

	ipc.Listen(ui)
	log.SetOutput(ap)

	// Run the message loop
	ui.Run()
	gui_win32.ShutdownGDIPlus()

	// send stop action to SuperviseDockerNode
	ap.TriggerAction("stop")
	dc.Shutdown()

	if debugMode != "" {
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}
