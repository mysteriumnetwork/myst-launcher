/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

//go:generate goversioninfo -icon=ico/icon_512x512.ico -manifest=launcher.exe.manifest
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/lxn/walk"
)

const (
	flagTray    = "-tray"
	flagInstall = "-install-binary"
)

func main() {
	if len(os.Args) > 1 {
		mod.inTray = os.Args[1] == flagTray

		if os.Args[1] == flagInstall {
			fmt.Println(flagInstall, checkExe())
			installExe()
			return
		}
	}

	if !isAnotherInstanceRunning() {
		return
	}
	mod.icon, _ = walk.NewIconFromResourceId(2)
	createDialogue()

	productName := windowsProductName()
	if productSupported(productName) {
		go func() {
			superviseDockerNode()
		}()
	} else {
		sadMsg := fmt.Sprintf(`Supported windows products are: %s.Your windows product: %s`, strings.Join(supportedProductName, ", "), productName)
		mod.lv.PostAppendText(sadMsg)
	}

	createNotifyIcon()
}
