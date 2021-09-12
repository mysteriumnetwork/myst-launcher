// +build !windows

package app

import (
	"log"

	"github.com/mysteriumnetwork/myst-launcher/gui"
)

// returns exit model: true means exit
func (s *AppState) tryInstall() bool {
	log.Println("tryInstall >>>>>")
	s.model.SwitchState(gui.ModalStateInstallNeeded)



	//s.model.SwitchState(gui.ModalStateInstallFinished)
	ok := s.model.WaitDialogueComplete()
	if !ok {
		return true
	}
	s.model.SwitchState(gui.ModalStateInitial)

	return false
}
