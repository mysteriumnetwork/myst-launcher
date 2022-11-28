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
	"os"

	"github.com/mysteriumnetwork/myst-launcher/controller"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

func main() {
	utils.AllocConsole(false)
	defer func() {
		fmt.Println("Press Enter to exit ..")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}()

	controller.CheckLauncherUpdatesCli()
}
