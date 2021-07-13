package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
)

func uiTest() {
	model.SwitchState(installNeeded)
	model.WaitDialogueComplete()

	model.SwitchState(installInProgress)
	model.WaitDialogueComplete()

	model.checkWindowsVersion = true
	model.TriggerUpdate()

	model.installationStatus = fmt.Sprintf("Downloading 1 of 2: abc")
	model.TriggerUpdate()
	model.SwitchState(installInProgress)
	model.WaitDialogueComplete()

	model.installationStatus = fmt.Sprintf("Downloading 2 of 2: abc")
	model.TriggerUpdate()
	model.SwitchState(installInProgress)
	model.WaitDialogueComplete()

	model.installationStatus = "Reason:\r\nCommand failed: failed to enable Microsoft-Windows-Subsystem-Linux"
	model.SwitchState(installError)
	model.WaitDialogueComplete()

	model.installationStatus = "Installation successfully finished!"
	model.SwitchState(installFinished)
	model.WaitDialogueComplete()
	model.SwitchState(initial)

	model.SwitchState(installInProgress)

	log.Println("msiexec")
	exe := "msiexec.exe"
	cmdArgs := "/i " + os.Getenv("TMP") + "\\wsl_update_x64.msi /quiet"
	err := _ShellExecuteAndWait(0, "runas", exe, cmdArgs, os.Getenv("TMP"), syscall.SW_NORMAL)
	if err != nil {
		model.installationStatus = "Error:\r\nCommand failed: msiexec.exe /i wsl_update_x64.msi /quiet"
		model.TriggerUpdate()
		model.SwitchState(installError)

		model.WaitDialogueComplete()
		model.ExitApp()
		return
	}
	model.SwitchState(installFinished)
	model.WaitDialogueComplete()
	model.SwitchState(initial)
	return
}
