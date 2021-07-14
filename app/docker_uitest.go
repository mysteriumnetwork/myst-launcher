package app

import (
	"fmt"
	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"log"
	"os"
	"syscall"
)

func uiTest() {
	gui.SModel.SwitchState(gui.ModalStateInstallNeeded)
	gui.SModel.WaitDialogueComplete()

	gui.SModel.SwitchState(gui.ModalStateInstallInProgress)
	gui.SModel.WaitDialogueComplete()

	gui.SModel.CheckWindowsVersion = true
	gui.SModel.TriggerUpdate()

	log.Println(fmt.Sprintf("Downloading 1 of 2: abc"))
	gui.SModel.SwitchState(gui.ModalStateInstallInProgress)
	gui.SModel.WaitDialogueComplete()

	log.Println(fmt.Sprintf("Downloading 2 of 2: abc"))
	gui.SModel.SwitchState(gui.ModalStateInstallInProgress)
	gui.SModel.WaitDialogueComplete()

	log.Println("Reason:\r\nCommand failed: failed to enable Microsoft-Windows-Subsystem-Linux")
	gui.SModel.SwitchState(gui.ModalStateInstallError)
	gui.SModel.WaitDialogueComplete()

	log.Println("Installation successfully finished!")
	gui.SModel.SwitchState(gui.ModalStateInstallFinished)
	gui.SModel.WaitDialogueComplete()
	gui.SModel.SwitchState(gui.ModalStateInitial)
	gui.SModel.SwitchState(gui.ModalStateInstallInProgress)

	log.Println("msiexec")
	exe := "msiexec.exe"
	cmdArgs := "/i " + os.Getenv("TMP") + "\\wsl_update_x64.msi /quiet"
	err := native.ShellExecuteAndWait(0, "runas", exe, cmdArgs, os.Getenv("TMP"), syscall.SW_NORMAL)
	if err != nil {
		log.Println("Error:\r\nCommand failed: msiexec.exe /i wsl_update_x64.msi /quiet")

		gui.SModel.SwitchState(gui.ModalStateInstallError)
		gui.SModel.WaitDialogueComplete()
	}
	gui.SModel.SwitchState(gui.ModalStateInstallFinished)
	gui.SModel.WaitDialogueComplete()
	gui.SModel.SwitchState(gui.ModalStateInitial)
	return
}
