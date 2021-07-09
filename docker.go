/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

const (
	docker = "docker"
	group  = "docker-users"
)

func superviseDockerNode() {
	model.refreshState()
	dockerCmd := os.Getenv("ProgramFiles") + "\\Docker\\Docker\\resources\\bin\\" + docker

	productName := windowsProductName()
	if !productSupported(productName) {
		sadMsg := fmt.Sprintf(`Supported windows products are: %s.Your windows product: %s`, strings.Join(supportedProductName, ", "), productName)
		model.lbInstallationState2.SetText(sadMsg)
		model.SwitchState(installError)

		model.WaitDialogueComplete()
		model.ExitApp()
		return
	}

	for {
		ex := cmdRun(dockerCmd, "ps")
		switch ex {
		case 0:
			model.lbDocker.SetText("Running [OK]")
			ex := cmdRun(dockerCmd, "container", "start", "myst")

			switch ex {
			case 0:
				model.lbContainer.SetText("Running [OK]")
				model.btnOpenNodeUI.SetEnabled(true)

			default:
				log.Printf("Failed to start cmd: %v", ex)
				model.lbContainer.SetText("Installing")

				ex := cmdRun(dockerCmd, strings.Split("run --cap-add NET_ADMIN -d -p 4449:4449 --name myst -v myst-data:/var/lib/mysterium-node mysteriumnetwork/myst:latest service --agreed-terms-and-conditions", " ")...)
				if ex == 0 {
					model.lbDocker.SetText("Running [OK]")
					continue
				}
			}

		case 1:
			model.lbDocker.SetText("Starting..")
			model.lbContainer.SetText("-")

			if isProcessRunning("Docker Desktop.exe") {
				break
			}
			dd := os.Getenv("ProgramFiles") + "\\Docker\\Docker\\Docker Desktop.exe"
			cmd := exec.Command(dd, "-Autostart")
			if err := cmd.Start(); err != nil {
				log.Printf("Failed to start cmd: %v", err)
			}
			break

		default:
			model.SwitchState(installNeeded)
			model.WaitDialogueComplete()
			model.SwitchState(installInProgress)

			if !CheckWindowsVersion() {
				model.lbInstallationState2.SetText("Reason:\r\nYou must be running Windows 10 version 1607 (the Anniversary update) or above.")
				model.SwitchState(installError)
				model.WaitDialogueComplete()
				model.ExitApp()
				return
			}

			list := []struct{ url, name string }{
				{"https://desktop.docker.com/win/stable/amd64/Docker%20Desktop%20Installer.exe", "DockerDesktopInstaller.exe"},
				{"https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi", "wsl_update_x64.msi"},
			}
			for fi, v := range list {
				if _, err := os.Stat(os.Getenv("TMP") + v.name); err != nil {

					model.lbInstallationState2.SetText(fmt.Sprintf("%d of %d: %s", fi+1, len(list), v.name))
					model.PrintProgress(0)

					err := DownloadFile(os.Getenv("TMP")+"\\"+v.name, v.url, model.PrintProgress)
					if err != nil {
						model.lbInstallationState2.SetText("Reason:\r\nDownload failed")
						model.SwitchState(installError)
						model.WaitDialogueComplete()
						model.ExitApp()
						return
					}
				}
			}

			err := runMeElevated("msiexec.exe", "/I wsl_update_x64.msi /quiet", os.Getenv("TMP"))
			if err != nil {
				model.lbInstallationState2.SetText("Reason:\r\nCommand failed: msiexec.exe /I wsl_update_x64.msi")
				model.SwitchState(installError)
				model.WaitDialogueComplete()
				model.ExitApp()
				return
			}
			ex := cmdRun(os.Getenv("TMP")+"\\DockerDesktopInstaller.exe", "install", "--quiet")
			if ex != 0 {
				model.lbInstallationState2.SetText("Reason:\r\nDockerDesktopInstaller failed")
				model.SwitchState(installError)
				model.WaitDialogueComplete()
				model.ExitApp()
				return
			}

			if !checkExe() {
				installExe()
			}
			if !CurrentGroupMembership(group) {
				// request to logout //

				ret := walk.MsgBox(model.mw, "Installation", "Log of from the current session to finish the installation.", walk.MsgBoxTopMost|walk.MsgBoxYesNo|walk.MsgBoxIconExclamation)
				if ret == win.IDYES {
					windows.ExitWindowsEx(windows.EWX_LOGOFF, 0)
					return
				}
				model.SwitchState(installError)
				model.lbInstallationState2.SetText("Log of from the current session to finish the installation.")
				model.WaitDialogueComplete()
				model.ExitApp()
				return
			}

			model.SwitchState(installFinished)
			model.WaitDialogueComplete()
			model.SwitchState(initial)
			continue
		}

		time.Sleep(10 * time.Second)
	}
}
