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
	"os/signal"
	"syscall"

	"github.com/mysteriumnetwork/myst-launcher/app"
	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/controller/native"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

func main() {
	installFirewall := flag.Bool(_const.FlagInstallFirewall, false, "setup firewall rules")
	nodeFlags := flag.String("node-flags", "", "pass flags to node")
	flag.Parse()

	defer func() {
		fmt.Println("Press Enter to exit ..")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}()

	if *installFirewall {
		log.Println("Setting firewall rules")
		native.CheckAndInstallFirewallRules()
		return
	}

	ap := app.NewApp(false)

	mod := model.NewUIModel()
	mod.SetApp(ap)
	mod.DuplicateLogToConsole = true
	mod.NodeFlags = *nodeFlags

	prodVersion, _ := utils.GetProductVersion()
	mod.SetProductVersion(prodVersion)
	log.Println("Launcher version:", prodVersion)

	ap.SetModel(mod)
	log.SetOutput(ap)
	ap.StartAppController()

	// wait for Ctrl-C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	ap.StopAppController()
}
