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
	_const "github.com/mysteriumnetwork/myst-launcher/const"
	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
	"github.com/tryor/gdiplus"
	"github.com/tryor/winapi"
)

func main() {
	defer utils.PanicHandler("main")

	ap := app.NewApp()
	p := app.NewPipeHandler()

	if len(os.Args) > 1 {
		ap.InTray = os.Args[1] == _const.FlagTray
		ap.InstallStage2 = os.Args[1] == _const.FlagInstallStage2

		switch os.Args[1] {
		case _const.FlagInstall:
			utils.InstallExe()
			return

		case _const.FlagUninstall:
			p.SendStopApp()
			utils.UninstallExe()
			return
		}
	}

	mod := model.NewUIModel()
	mod.SetApp(ap)
	mod.Config.DuplicateLogToConsole = true
	mod.Config.ProductVersion, _ = utils.GetProductVersion()
	
	log.Println("Initializing GDI+ ....")
	var gpToken winapi.ULONG_PTR
	var input gdiplus.GdiplusStartupInput
	input.GdiplusVersion = 1
	_, err := gdiplus.Startup(&gpToken, &input, nil)
	if err != nil {
		panic(err)
	}
	
	ui := gui_win32.NewGui(mod)

	// update launcher binary
	if utils.LauncherUpgradeAvailable() {
		update := ui.YesNoModal("Mysterium launcher upgrade", "You are running a newer version of launcher.\r\nUpgrade launcher installation ?")
		if model.IDYES == update {
			if !p.OwnsPipe() {
				p.SendStopApp()
				p.OpenPipe()
			}
			utils.UpdateExe()
		} else {
			if !p.OwnsPipe() {
				p.SendPopupApp()
				return
			}
		}
		// continue execution
	} else {
		if !p.OwnsPipe() {
			p.SendPopupApp()
			return
		}
	}


	ui.CreateNotifyIcon(mod)
	ui.CreateMainWindow()

	ap.SetModel(mod)
	ap.SetUI(ui)
	ap.WaitGroup.Add(1)

	go ap.SuperviseDockerNode()
	p.Listen(ui)

	log.SetOutput(ap)

	// Run the message loop
	ui.Run()
	gdiplus.GdiplusShutdown(gpToken)

	// send stop action to SuperviseDockerNode
	ap.TriggerAction("stop")

	// wait for SuperviseDockerNode to finish its work
	ap.WaitGroup.Wait()
}
