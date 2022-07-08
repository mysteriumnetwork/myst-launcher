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
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/gonutz/w32"

	"github.com/mysteriumnetwork/myst-launcher/app"
	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/controller/docker"
	"github.com/mysteriumnetwork/myst-launcher/controller/native"
	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	ipc_ "github.com/mysteriumnetwork/myst-launcher/ipc"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/updates"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

var debugMode = ""

func main() {
	defer utils.PanicHandler("main")
	go http.ListenAndServe("localhost:8080", nil)

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
	log.SetOutput(ap)

	ui.CreateNotifyIcon(mod)
	ui.CreateMainWindow()
	ap.SetUI(ui)

	setUIController := func() {
		// fmt.Println("set controller", mod.Config.Backend)
		var nc app.Controller

		switch mod.Config.Backend {
		case "native":
			nc = native.NewController()
		case "docker":
			nc = docker.NewController()
		}
		ap.SetAppController(nc)
		go nc.Start()
	}
	mod.Bus2.Subscribe("backend", setUIController)
	setUIController()

	ipc.Listen(ui)

	// Run the message loop
	ui.Run()
	gui_win32.ShutdownGDIPlus()
	ap.Stop()

	if debugMode != "" {
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}
