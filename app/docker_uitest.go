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
	gui.UI.SwitchState(gui.ModalStateInstallNeeded)
	gui.UI.WaitDialogueComplete()

	gui.UI.SwitchState(gui.ModalStateInstallInProgress)
	gui.UI.WaitDialogueComplete()

	gui.UI.CheckWindowsVersion = true
	gui.UI.Update()

	log.Println(fmt.Sprintf("Downloading 1 of 2: abc"))
	gui.UI.SwitchState(gui.ModalStateInstallInProgress)
	gui.UI.WaitDialogueComplete()

	log.Println(fmt.Sprintf("Downloading 2 of 2: abc"))
	gui.UI.SwitchState(gui.ModalStateInstallInProgress)
	gui.UI.WaitDialogueComplete()

	log.Println("Reason:\r\nCommand failed: failed to enable Microsoft-Windows-Subsystem-Linux")
	gui.UI.SwitchState(gui.ModalStateInstallError)
	gui.UI.WaitDialogueComplete()

	log.Println("Installation successfully finished!")
	gui.UI.SwitchState(gui.ModalStateInstallFinished)
	gui.UI.WaitDialogueComplete()
	gui.UI.SwitchState(gui.ModalStateInitial)
	gui.UI.SwitchState(gui.ModalStateInstallInProgress)

	log.Println("msiexec")
	exe := "msiexec.exe"
	cmdArgs := "/i " + os.Getenv("TMP") + "\\wsl_update_x64.msi /quiet"
	err := native.ShellExecuteAndWait(0, "runas", exe, cmdArgs, os.Getenv("TMP"), syscall.SW_NORMAL)
	if err != nil {
		log.Println("Error:\r\nCommand failed: msiexec.exe /i wsl_update_x64.msi /quiet")

		gui.UI.SwitchState(gui.ModalStateInstallError)
		gui.UI.WaitDialogueComplete()
	}
	gui.UI.SwitchState(gui.ModalStateInstallFinished)
	gui.UI.WaitDialogueComplete()
	gui.UI.SwitchState(gui.ModalStateInitial)
	return
}
