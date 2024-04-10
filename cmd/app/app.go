/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gonutz/w32"

	"github.com/mysteriumnetwork/myst-launcher/app"
	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/controller"
	"github.com/mysteriumnetwork/myst-launcher/controller/docker"
	"github.com/mysteriumnetwork/myst-launcher/controller/native"
	"github.com/mysteriumnetwork/myst-launcher/controller/shutdown"
	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	ipc_ "github.com/mysteriumnetwork/myst-launcher/ipc"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

func pressAnyKey(do bool) {
	if !do {
		return
	}
	fmt.Println("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func main() {
	installFirewall := flag.Bool(_const.FlagInstallFirewall, false, "setup firewall rules")
	nodeArgsFlags := flag.String(_const.FlagNodeArgs, "", "pass args to node")
	installFlag := flag.Bool(_const.FlagInstall, false, "install")
	uninstallFlag := flag.Bool(_const.FlagUninstall, false, "uninstall")
	stopFlag := flag.Bool(_const.FlagStop, false, "stop all instances of app")
	debugMode := flag.Bool(_const.FlagDebug, false, "debug mode")
	flagAutorun := flag.Bool(_const.FlagAutorun, false, "app is started by means of autorun")
	flag.Parse()

	if *debugMode {
		utils.AllocConsole(false)
	}
	fmt.Println(*debugMode, *installFirewall, os.Args)

	ap := app.NewApp()

	if *installFirewall {
		log.Println("Setting firewall rules")
		native.CheckAndInstallFirewallRules()
		pressAnyKey(*debugMode)
		return
	}
	if *installFlag {
		utils.InstallExe()
		pressAnyKey(*debugMode)
		return
	}

	ipc := ipc_.NewHandler()
	defer func() {
		ipc.Close()
		pressAnyKey(*debugMode)
	}()

	if *uninstallFlag {
		ipc.SendStopApp()
		utils.UninstallExe()

		// for older versions e.g <=1.0.30
		// it should be shutdown forcefully with a related docker container

		native.KillPreviousLauncher()
		if err := docker.UninstallMystContainer(); err != nil {
			fmt.Println("UninstallMystContainer failed:", err)
		}
		return
	}
	if *stopFlag {
		ipc.SendStopApp()
		// wait for main process to finish, this is important for MSI to finish
		// 10 sec -- for docker container to stop
		// TODO: modify IPC proto to get rid of this sleep
		time.Sleep(10 * time.Second)
		return
	}

	mod := model.NewUIModel()
	if *flagAutorun && !mod.Config.AutoStart {
		return
	}
	mod.SetApp(ap)
	mod.DuplicateLogToConsole = true
	mod.FlagAutorun = *flagAutorun
	mod.NodeFlags = *nodeArgsFlags

	// in case of installation restart elevated
	if mod.Config.InitialState == model.InitialStateStage1 || mod.Config.InitialState == model.InitialStateStage2 {
		if !w32.SHIsUserAnAdmin() {
			utils.RunasWithArgsNoWait("")
			return
		}
	}

	prodVersion, _ := utils.GetProductVersion()
	mod.SetProductVersion(prodVersion)
	if *debugMode {
		log.Println("Product version:", prodVersion)
	}

	gui_win32.InitGDIPlus()
	ui := gui_win32.NewGui(mod)

	// skip update if IsUserAnAdmin
	if !w32.SHIsUserAnAdmin() && controller.UpdateLauncherFromNewBinary(ui, ipc) {
		return
	}

	// exit if another instance is running already
	if controller.PopupFirstInstance(ui, ipc) {
		return
	}

	ap.SetModel(mod)
	log.SetOutput(ap)

	// init ui in separate thread b.c of Coinitialize modes conflict when WMI is invoked in the same goroutine
	ui.CreateNotifyIcon(mod)
	ui.CreateMainWindow()
	ap.SetUI(ui)

	mod.Sh = shutdown.NewShutdownController(mod.UIBus)

	go controller.CheckLauncherUpdates(mod)
	ap.StartAppController()
	ipc.Listen(ui)

	// Run the message loop
	ui.Run()
	gui_win32.ShutdownGDIPlus()
	ap.StopAppController()
}
