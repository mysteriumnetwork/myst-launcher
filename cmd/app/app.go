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
	"time"

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

func main() {
	ap := app.NewApp()
	ipc := ipc_.NewHandler()

	cmd := ""
	debugMode := false
	for _, v := range os.Args {
		switch v {
		case _const.FlagInstall,
			_const.FlagUninstall,
			_const.FlagInstallFirewall,
			_const.FlagStop:
			cmd = v
		case _const.FlagDebug:
			debugMode = true
		}
	}
	if debugMode {
		utils.AllocConsole(false)
		defer func() {
			fmt.Println("Press 'Enter' to continue...")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}()
	}

	if cmd != "" {
		switch cmd {
		case _const.FlagInstall:
			utils.InstallExe()
			return

		case _const.FlagUninstall:
			ipc.SendStopApp()
			utils.UninstallExe()

			// for older versions e.g <=1.0.30
			// it should be shutdown forcefully with a related docker container

			native.KillPreviousLauncher()
			if err := docker.UninstallMystContainer(); err != nil {
				fmt.Println("UninstallMystContainer failed:", err)
				return
			}

		case _const.FlagStop:
			ipc.SendStopApp()
			time.Sleep(1 * time.Second) // wait for main process to finish, this is important for MSI to finish
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
	err := utils.EnableAutorun(mod.Config.AutoStart)
	fmt.Println("utils.EnableAutorun", err)

	prodVersion, _ := utils.GetProductVersion()
	mod.SetProductVersion(prodVersion)
	if debugMode {
		log.Println("Product version:", prodVersion)
	}

	gui_win32.InitGDIPlus()
	ui := gui_win32.NewGui(mod)

	// exit if another instance is running already
	if updates.PopupFirstInstance(ui, ipc) {
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

}
