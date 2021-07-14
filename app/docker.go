/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

const (
	docker = "docker"
	group  = "docker-users"
)

func SuperviseDockerNode() {
	dockerCmd := os.Getenv("ProgramFiles") + "\\Docker\\Docker\\resources\\bin\\" + docker
	for {
		ex := cmdRun(dockerCmd, "ps")
		switch ex {
		case 0:
			SModel.stateDocker = stateRunning
			SModel.TriggerUpdate()

			ex := cmdRun(dockerCmd, "container", "start", "myst")
			switch ex {
			case 0:
				SModel.stateContainer = stateRunning
				SModel.TriggerUpdate()

			default:
				log.Printf("Failed to start cmd: %v", ex)
				SModel.stateContainer = stateInstalling
				SModel.TriggerUpdate()

				ex := cmdRun(dockerCmd, strings.Split("run --cap-add NET_ADMIN -d -p 4449:4449 --name myst -v myst-data:/var/lib/mysterium-node mysteriumnetwork/myst:latest service --agreed-terms-and-conditions", " ")...)
				if ex == 0 {
					SModel.stateContainer = stateRunning
					SModel.TriggerUpdate()

					continue
				}
			}

		case 1:
			SModel.stateDocker = stateStarting
			SModel.stateContainer = stateUnknown
			SModel.TriggerUpdate()

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
			var err error
			if !SModel.InstallStage2 {
				SModel.SwitchState(installNeeded)
				SModel.WaitDialogueComplete()
			}
			SModel.SwitchState(installInProgress)

			log.Println("Checking Windows version")
			SModel.TriggerUpdate()
			if !CheckWindowsVersion() {
				log.Println("You must run Windows 10 version 2004 or above.")
				SModel.SwitchState(installError)

				if !isWindowsUpdateEnabled() {
					exe := "reg"
					cmdArgs := "add \"HKLM\\SOFTWARE\\Policies\\Microsoft\\Windows\\WindowsUpdate\" /v DisableWUfBSafeguards /t REG_DWORD /d 1 /f"
					err := _ShellExecuteAndWait(0, "runas", exe, cmdArgs, "", syscall.SW_HIDE)
					if err != nil {
						log.Println("Command failed: failed to enable Windows Updates")
						SModel.SwitchState(installError)

						SModel.WaitDialogueComplete()
						SModel.ExitApp()
						return
					}
				}
				ret := walk.MsgBox(SModel.mw, "Installation", "Please signal to Windows 10 version 2004 or above. \r\nClick OK to open Update settings", walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
				if ret == win.IDOK {
					cmdRun("rundll32", "url.dll,FileProtocolHandler", "ms-settings:windowsupdate-action")
				}
				SModel.WaitDialogueComplete()
				SModel.ExitApp()
				return
			}
			SModel.checkWindowsVersion = true
			SModel.TriggerUpdate()

			log.Println("Checking VT-x")
			if !hasVTx() {
				walk.MsgBox(SModel.mw, "Installation", "Please Enable virtualization in BIOS", walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)

				SModel.WaitDialogueComplete()
				SModel.ExitApp()
				return
			}
			SModel.checkVTx = true
			SModel.TriggerUpdate()

			if !SModel.InstallStage2 {

				log.Println("Enable WSL..")
				exe := "dism.exe"
				cmdArgs := "/online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart"
				err = _ShellExecuteAndWait(0, "runas", exe, cmdArgs, "", syscall.SW_HIDE)
				if err != nil {
					log.Println("Command failed: failed to enable Microsoft-Windows-Subsystem-Linux")
					SModel.SwitchState(installError)

					SModel.WaitDialogueComplete()
					SModel.ExitApp()
					return
				}
				SModel.enableWSL = true
				SModel.TriggerUpdate()

				log.Println("Install executable")
				fullExe, _ := os.Executable()
				cmdArgs = FlagInstall
				err = _ShellExecuteAndWait(0, "runas", fullExe, cmdArgs, "", syscall.SW_NORMAL)
				if err != nil {
					log.Println("Failed to install executable")
					SModel.SwitchState(installError)

					SModel.WaitDialogueComplete()
					SModel.ExitApp()
					return
				}
				CreateAutostartShortcut(FlagInstallStage2)
				SModel.installExecutable = true
				SModel.TriggerUpdate()

				if true {
					ret := walk.MsgBox(SModel.mw, "Installation", "Reboot is needed to finish installation of WSL\r\nClick OK to reboot", walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
					if ret == win.IDOK {
						_ShellExecuteAndWait(0, "", "shutdown", "-r", "", syscall.SW_NORMAL)
					}
					SModel.WaitDialogueComplete()
					SModel.ExitApp()
					return
				}
			} else {
				// proceeding install after reboot
				SModel.enableWSL = true
				SModel.installExecutable = true
				SModel.rebootAfterWSLEnable = true
				SModel.TriggerUpdate()
			}

			CreateAutostartShortcut("")

			list := []struct{ url, name string }{
				{"https://desktop.docker.com/win/stable/amd64/Docker%20Desktop%20Installer.exe", "DockerDesktopInstaller.exe"},
				{"https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi", "wsl_update_x64.msi"},
			}
			for fi, v := range list {
				log.Println(fmt.Sprintf("Downloading %d of %d: %s", fi+1, len(list), v.name))
				if _, err := os.Stat(os.Getenv("TMP") + "\\" + v.name); err != nil {
					SModel.SetProgress(0)

					err := DownloadFile(os.Getenv("TMP")+"\\"+v.name, v.url, SModel.SetProgress)
					if err != nil {
						log.Println("Download failed")
						SModel.SwitchState(installError)

						SModel.WaitDialogueComplete()
						SModel.ExitApp()
						return
					}
				}
			}
			SModel.downloadFiles = true

			log.Println("Installing wsl_update_x64.msi")
			exe := "msiexec.exe"
			cmdArgs := "/i " + os.Getenv("TMP") + "\\wsl_update_x64.msi /quiet"
			err = _ShellExecuteAndWait(0, "runas", exe, cmdArgs, os.Getenv("TMP"), syscall.SW_NORMAL)
			if err != nil {
				log.Println("Command failed: msiexec.exe /i wsl_update_x64.msi /quiet")
				SModel.SwitchState(installError)

				SModel.WaitDialogueComplete()
				SModel.ExitApp()
				return
			}
			SModel.installWSLUpdate = true
			SModel.TriggerUpdate()

			log.Println("installing docker desktop")
			ex := cmdRun(os.Getenv("TMP")+"\\DockerDesktopInstaller.exe", "install", "--quiet")
			if ex != 0 {
				log.Println("DockerDesktopInstaller failed")
				SModel.SwitchState(installError)

				SModel.WaitDialogueComplete()
				SModel.ExitApp()
				return
			}
			SModel.installDocker = true
			SModel.TriggerUpdate()

			log.Println("Checking current group membership")
			if !CurrentGroupMembership(group) {
				// request a logout //

				ret := walk.MsgBox(SModel.mw, "Installation", "Log of from the current session to finish the installation.", walk.MsgBoxTopMost|walk.MsgBoxYesNo|walk.MsgBoxIconExclamation)
				if ret == win.IDYES {
					windows.ExitWindowsEx(windows.EWX_LOGOFF, 0)
					return
				}
				log.Println("Log of from the current session to finish the installation.")
				SModel.SwitchState(installError)

				SModel.WaitDialogueComplete()
				SModel.ExitApp()
				return
			}
			SModel.readConfig()
			SModel.cfg.AutoStart = true
			SModel.saveConfig()

			SModel.checkGroupMembership = true
			SModel.TriggerUpdate()

			SModel.SwitchState(installFinished)
			SModel.WaitDialogueComplete()
			SModel.SwitchState(initial)
			continue
		}

		time.Sleep(10 * time.Second)
	}
}
