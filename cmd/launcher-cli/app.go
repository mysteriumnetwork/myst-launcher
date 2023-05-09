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
	"os/signal"
	"syscall"

	"github.com/mysteriumnetwork/myst-launcher/app"
	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/controller/native"
	"github.com/mysteriumnetwork/myst-launcher/model"
)

func main() {

	cmd := ""
	for i, v := range os.Args {
		if i == 0 {
			continue
		}
		switch v {
		case _const.FlagInstallFirewall:
			cmd = v
		default:
			log.Println("Unknown arg")
			return
		}
	}

	defer func() {
		fmt.Println("Press Enter to exit ..")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}()

	if cmd == _const.FlagInstallFirewall {
		native.CheckAndInstallFirewallRules()
		return
	}

	ap := app.NewApp()

	mod := model.NewUIModel()
	mod.SetApp(ap)
	mod.DuplicateLogToConsole = true

	ap.SetModel(mod)
	log.SetOutput(ap)
	ap.StartAppController()

	// wait for Ctrl-C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	ap.StopAppController()
}
