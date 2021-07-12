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

func superviseDockerNode() {
	model.refreshState()
	dockerCmd := os.Getenv("ProgramFiles") + "\\Docker\\Docker\\resources\\bin\\" + docker

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
			var err error
			model.SwitchState(installNeeded)
			model.WaitDialogueComplete()
			model.SwitchState(installInProgress)

			if !CheckWindowsVersion() {
				model.lbInstallationState2.SetText("Reason:\r\nYou must run Windows 10 version 2004 or above.")
				model.SwitchState(installError)

				if !isWindowsUpdateEnabled() {
					exe := "reg"
					cmdArgs := "add \"HKLM\\SOFTWARE\\Policies\\Microsoft\\Windows\\WindowsUpdate\" /v DisableWUfBSafeguards /t REG_DWORD /d 1 /f"
					err := _ShellExecuteAndWait(0, "runas", exe, cmdArgs, "", syscall.SW_HIDE)
					if err != nil {
						model.lbInstallationState2.SetText("Reason:\r\nCommand failed: failed to enable Windows Updates")
						model.SwitchState(installError)
						model.WaitDialogueComplete()
						model.ExitApp()
						return
					}
				}
				ret := walk.MsgBox(model.mw, "Installation", "Please update to Windows 10 version 2004 or above. \r\nClick OK to open Update settings", walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
				if ret == win.IDOK {
					cmdRun("rundll32", "url.dll,FileProtocolHandler", "ms-settings:windowsupdate-action")
				}
				model.WaitDialogueComplete()
				model.ExitApp()
				return
			}
			if !hasVTx() {
				walk.MsgBox(model.mw, "Installation", "Please Enable virtualization in BIOS", walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)

				model.WaitDialogueComplete()
				model.ExitApp()
				return
			}

			if !model.installStage2 {
				log.Println("enable-feature..")
				var err error
				{
					exe := "dism.exe"
					cmdArgs := "/online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart"
					err = _ShellExecuteAndWait(0, "runas", exe, cmdArgs, "", syscall.SW_HIDE)
					if err != nil {
						model.lbInstallationState2.SetText("Reason:\r\nCommand failed: failed to enable Microsoft-Windows-Subsystem-Linux")
						model.SwitchState(installError)
						model.WaitDialogueComplete()
						model.ExitApp()
						return
					}
				}
				if !checkExe() {
					fullExe, _ := os.Executable()
					cmdArgs := flagInstall
					_ShellExecuteAndWait(0, "runas", fullExe, cmdArgs, "", syscall.SW_NORMAL)
					CreateAutostartShortcut(flagInstallStage2)
				}

				if true {
					ret := walk.MsgBox(model.mw, "Installation", "Reboot is needed to finish installation of WSL\r\nClick OK to reboot", walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
					if ret == win.IDOK {
						_ShellExecuteAndWait(0, "runas", "shutdown", "-r", "", syscall.SW_NORMAL)
					}
					model.WaitDialogueComplete()
					model.ExitApp()
					return
				}
			}

			CreateAutostartShortcut("")
			log.Println("downloading..")
			list := []struct{ url, name string }{
				{"https://desktop.docker.com/win/stable/amd64/Docker%20Desktop%20Installer.exe", "DockerDesktopInstaller.exe"},
				{"https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi", "wsl_update_x64.msi"},
			}
			for fi, v := range list {
				if _, err := os.Stat(os.Getenv("TMP") + "\\" + v.name); err != nil {

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
			model.lbInstallationState2.SetText("")

			log.Println("msiexec")
			exe := "msiexec.exe"
			cmdArgs := "/i " + os.Getenv("TMP") + "\\wsl_update_x64.msi /quiet"
			err = _ShellExecuteAndWait(0, "runas", exe, cmdArgs, os.Getenv("TMP"), syscall.SW_NORMAL)
			if err != nil {
				model.lbInstallationState2.SetText("Reason:\r\nCommand failed: msiexec.exe /i wsl_update_x64.msi")
				model.SwitchState(installError)
				model.WaitDialogueComplete()
				model.ExitApp()
				return
			}

			log.Println("DockerDesktopInstaller")
			model.lbInstallationState2.SetText("DockerDesktopInstaller.exe")
			ex := cmdRun(os.Getenv("TMP")+"\\DockerDesktopInstaller.exe", "install", "--quiet")
			if ex != 0 {
				model.lbInstallationState2.SetText("Reason:\r\nDockerDesktopInstaller failed")
				model.SwitchState(installError)
				model.WaitDialogueComplete()
				model.ExitApp()
				return
			}

			if !CurrentGroupMembership(group) {
				// request a logout //

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
