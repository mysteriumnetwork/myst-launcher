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
	"github.com/mysteriumnetwork/myst-launcher/controller/native"
	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	ipc_ "github.com/mysteriumnetwork/myst-launcher/ipc"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/updates"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

var debugMode = ""

func main() {
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

		case _const.FlagInstallFirewall:
			native.CheckAndInstallFirewallRules()
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
	utils.EnableAutorun(mod.Config.AutoStart)


	prodVersion, _ := utils.GetProductVersion()
	mod.SetProductVersion(prodVersion)
	if debugMode != "" {
		log.Println("Product version:", prodVersion)
	}

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

	ap.StartAppController()
	ipc.Listen(ui)

	// Run the message loop
	ui.Run()
	gui_win32.ShutdownGDIPlus()
	ap.StopAppController()

	if debugMode != "" {
		fmt.Println("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}
