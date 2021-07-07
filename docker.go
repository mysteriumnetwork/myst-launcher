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
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sys/windows"

	"github.com/lxn/walk"
)

const (
	docker = "docker_"
)

func checkSystemsAndTry() {
	mod.Invalidate()
	dckr := os.Getenv("ProgramFiles") + "\\Docker\\Docker\\resources\\bin\\" + docker

_begin:
	for i := 0; i < 10000; i++ {

		//time.Sleep(5 * time.Second)
		ex := runProc(mod.lv, dckr, []string{"ps"})
		switch ex {
		case 0:
			mod.lbDocker.SetText("Running [OK]")
			//mod.lbContainer.SetText("Starting...")

			ex := runProc(mod.lv, dckr, strings.Split("container start myst", " "))
			switch ex {
			case 0:
				mod.lbContainer.SetText("Running [OK]")
				mod.btnCmd2.SetEnabled(true)

			default:
				log.Printf("Failed to start cmd: %v", ex)
				mod.lbContainer.SetText("Installing")

				ex := runProc(mod.lv, dckr, strings.Split("run --cap-add NET_ADMIN -d -p 4449:4449 --name myst -v myst-data:/var/lib/mysterium-node mysteriumnetwork/myst:latest service --agreed-terms-and-conditions", " "))
				if ex == 0 {
					mod.lbDocker.SetText("Running [OK]")
					goto _begin
				}
			}

		case 1:
			mod.lbDocker.SetText("Starting..")
			mod.lbContainer.SetText("-")

			dd := os.Getenv("ProgramFiles") + "\\Docker\\Docker\\Docker Desktop.exe"
			cmd := exec.Command(dd)
			if err := cmd.Start(); err != nil {
				//log.Printf("Failed to start cmd: %v", err)
				//return
			}
			break
			//fmt.Println(dd)
			//ex := runProc(lv, dd, nil)
			//fmt.Println(ex)
			//goto _begin

		default:
			mod.SetState(INSTALL_NEED)
			mod.WaitDialogueComplete()

			list := []struct{ url, name string }{
				{"https://desktop.docker.com/win/stable/amd64/Docker%20Desktop%20Installer.exe", "DockerDesktopInstaller.exe"},
				{"https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi", "wsl_update_x64.msi"},
			}
			for _, v := range list {
				if _, err := os.Stat(os.Getenv("TMP") + v.name); err != nil {
					err := DownloadFile(os.Getenv("TMP")+"\\"+v.name, v.url, mod.PrintProgress)
					if err != nil {
						log.Println("Download failed")
					}
				}
			}

			log.Println("msiexec.exe /I wsl_update_x64.msi /quiet")
			err := runMeElevated("msiexec.exe", "/I wsl_update_x64.msi /quiet", os.Getenv("TMP"))
			if err != nil {
				log.Println(err)
			}
			ex := runProc(mod.lv, os.Getenv("TMP")+"\\DockerDesktopInstaller.exe", []string{"install", "--quiet"})
			if ex != 0 {
				log.Println("DockerDesktopInstaller failed", ex)
				goto _begin
			}

			if !checkExe() {
				installExe()
			}
			if !CurrentGroupMembership("docker-users_") {
				// request to logout

				walk.MsgBox(mod.mw, "Installation", "Log of from the current session to finish the installation.", walk.MsgBoxTopMost|walk.MsgBoxYesNo|walk.MsgBoxIconExclamation)
				windows.ExitWindowsEx(windows.EWX_LOGOFF, 0)
				return
			}

			mod.SetState(INSTALL_FIN)
			mod.WaitDialogueComplete()
			mod.SetState(0)
			goto _begin
		}
		time.Sleep(10000 * time.Millisecond)
	}
}
