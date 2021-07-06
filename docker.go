package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	docker = "docker"
)

func checkSystemsAndTry() {
	env := env()
	mod.Invalidate()

_begin:
	for i := 0; i < 10000; i++ {
		ex := runProc(mod.lv, docker, []string{"ps"})
		switch ex {
		case 0:
			mod.lbDocker.SetText("Running [OK]")
			//mod.lbContainer.SetText("Starting...")

			ex := runProc(mod.lv, docker, strings.Split("container start myst", " "))
			switch ex {
			case 0:
				mod.lbContainer.SetText("Running [OK]")
				mod.btnCmd2.SetEnabled(true)

			default:
				log.Printf("Failed to start cmd: %v", ex)
				mod.lbContainer.SetText("Installing")

				ex := runProc(mod.lv, docker, strings.Split("run --cap-add NET_ADMIN -d -p 4449:4449 --name myst -v myst-data:/var/lib/mysterium-node mysteriumnetwork/myst:latest service --agreed-terms-and-conditions", " "))
				if ex == 0 {
					mod.lbDocker.SetText("Running [OK]")
					goto _begin
				}
			}

		case 1:
			mod.lbDocker.SetText("Starting..")
			mod.lbContainer.SetText("-")

			dd := env["ProgramFiles"] + "\\Docker\\Docker\\Docker Desktop.exe"
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
			// if first time ?

			if _, err := os.Stat(env["TMP"] + "\\wsl_update_x64.msi"); err != nil {
				err := DownloadFile(env["TMP"]+"\\wsl_update_x64.msi", "https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi", mod.PrintProgress)
				if err != nil {
					log.Println("Download failed")
				}
			}
			log.Println("msiexec.exe /I wsl_update_x64.msi /quiet")
			err := runMeElevated("msiexec.exe", "/I wsl_update_x64.msi /quiet", env["TMP"])
			if err != nil {
				log.Println(err)
			}

			if _, err := os.Stat(env["TMP"] + "\\DockerDesktopInstaller.exe"); err != nil {
				err := DownloadFile(env["TMP"]+"\\DockerDesktopInstaller.exe", "https://desktop.docker.com/win/stable/amd64/Docker%20Desktop%20Installer.exe", mod.PrintProgress)
				if err != nil {
					log.Println("Download failed")
				}
			}
			ex := runProc(mod.lv, env["TMP"]+"\\DockerDesktopInstaller.exe", []string{"install", "--quiet"})
			if ex != 0 {
				log.Println("DockerDesktopInstaller failed", ex)
				goto _begin
			}

			mod.SetState(INSTALL_FIN)
			mod.WaitDialogueComplete()
			goto _begin
		}
		time.Sleep(10000 * time.Millisecond)
	}
}
