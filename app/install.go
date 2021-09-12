// +build windows

package app

import (
	"fmt"

	"github.com/mysteriumnetwork/myst-launcher/model"
)

// returns exit model: true means exit
func (s *AppState) tryInstall() bool {

	fmt.Println("tryInstall >>>>>")
	s.model.SwitchState(model.UIStateInstallNeeded)

	ok := s.ui.WaitDialogueComplete()
	if !ok {
		return true
	}

	//s.model.SwitchState(gui.UIStateInitial)

	fmt.Println("tryInstall >>>>> 2")
	s.model.SwitchState(model.UIStateInstallFinished)
	return true
}
