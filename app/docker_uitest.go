package app

import (
	"fmt"
	"log"
	"os"
	"syscall"
)

func uiTest() {
	SModel.SwitchState(installNeeded)
	SModel.WaitDialogueComplete()

	SModel.SwitchState(installInProgress)
	SModel.WaitDialogueComplete()

	SModel.checkWindowsVersion = true
	SModel.TriggerUpdate()

	log.Println(fmt.Sprintf("Downloading 1 of 2: abc"))
	SModel.SwitchState(installInProgress)
	SModel.WaitDialogueComplete()

	log.Println(fmt.Sprintf("Downloading 2 of 2: abc"))
	SModel.SwitchState(installInProgress)
	SModel.WaitDialogueComplete()

	log.Println("Reason:\r\nCommand failed: failed to enable Microsoft-Windows-Subsystem-Linux")
	SModel.SwitchState(installError)
	SModel.WaitDialogueComplete()

	log.Println("Installation successfully finished!")
	SModel.SwitchState(installFinished)
	SModel.WaitDialogueComplete()
	SModel.SwitchState(initial)
	SModel.SwitchState(installInProgress)

	log.Println("msiexec")
	exe := "msiexec.exe"
	cmdArgs := "/i " + os.Getenv("TMP") + "\\wsl_update_x64.msi /quiet"
	err := _ShellExecuteAndWait(0, "runas", exe, cmdArgs, os.Getenv("TMP"), syscall.SW_NORMAL)
	if err != nil {
		log.Println("Error:\r\nCommand failed: msiexec.exe /i wsl_update_x64.msi /quiet")

		SModel.SwitchState(installError)
		SModel.WaitDialogueComplete()
	}
	SModel.SwitchState(installFinished)
	SModel.WaitDialogueComplete()
	SModel.SwitchState(initial)
	return
}
