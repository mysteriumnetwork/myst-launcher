// +build darwin

package app

import (
	"fmt"
	"log"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

// returns exit model: true means exit
func (s *AppState) tryInstall() bool {
	s.model.SwitchState(model.UIStateInstallInProgress)

	features, err := utils.QueryFeatures()
	if err != nil {
		log.Println("Failed to query feature:", err)
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	if len(features) > 0 {
		log.Println("Virtualization is not supported !")
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	s.model.UpdateProperties(model.UIProps{"CheckVTx": true})

	hasDocker, err := utils.HasDocker()
	if err != nil {
		log.Println("Failed to check Docker:", err)
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}
	if hasDocker {
		s.model.UpdateProperties(model.UIProps{"CheckDocker": true})
	} else {
		log.Println("Please install Docker Desktop")
		s.model.SwitchState(model.UIStateInstallError)
		return true
	}

	s.model.SwitchState(model.UIStateInstallFinished)
	ok := s.ui.WaitDialogueComplete()
	if !ok {
		return true
	}
	s.model.SwitchState(model.UIStateInitial)
	return false
}
