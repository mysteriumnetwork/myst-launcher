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

	"github.com/lxn/walk"
)

const (
	flagTray          = "-tray"
	flagInstall       = "-install-binary"
	flagInstallStage2 = "-install-stage2"
)

func main() {
	if len(os.Args) > 1 {
		model.inTray = os.Args[1] == flagTray
		model.installStage2 = os.Args[1] == flagInstallStage2

		if os.Args[1] == flagInstall {
			installExe()
			return
		}
	}

	if !isAnotherInstanceRunning() {
		return
	}
	log.SetOutput(&model)
	model.icon, _ = walk.NewIconFromResourceId(2)
	createDialogue()

	go func() {
		superviseDockerNode()
	}()
	createNotifyIcon()
}
