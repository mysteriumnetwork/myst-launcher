// +build windows

package app

import (
	"fmt"

	"github.com/mysteriumnetwork/myst-launcher/gui"
)

// returns exit model: true means exit
func (s *AppState) tryInstall() bool {

	fmt.Println("tryInstall >>>>>")
	s.model.SwitchState(gui.ModalStateInstallNeeded)

	ok := s.ui.WaitDialogueComplete()
	if !ok {
		return true
	}

	//s.model.SwitchState(gui.ModalStateInitial)

	fmt.Println("tryInstall >>>>> 2")
	s.model.SwitchState(gui.ModalStateInstallFinished)
	return true
}
