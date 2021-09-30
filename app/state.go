package app

import (
	"fmt"
	"sync"

	model2 "github.com/mysteriumnetwork/myst-launcher/model"
)

type AppState struct {
	// flags
	InTray        bool
	InstallStage2 bool

	WaitGroup sync.WaitGroup // for graceful shutdown

	action chan string
	model  *model2.UIModel //gui.Model
	ui     model2.Gui_

	// state
	didInstallation bool
}

func NewApp() *AppState {
	s := &AppState{}
	s.action = make(chan string, 1)
	return s
}

func (s *AppState) SetModel(ui *model2.UIModel) {
	s.model = ui
}

func (s *AppState) SetUI(ui model2.Gui_) {
	s.ui = ui
}

func (s *AppState) Write(b []byte) (int, error) {
	// copy to avoid data corruption
	// see https://stackoverflow.com/a/20688698/4413696
	bCopy := make([]byte, len(b))
	copy(bCopy, b)

	if s.model.GetConfig().DuplicateLogToConsole {
		fmt.Print(string(bCopy))
	}

	s.model.Publish("log", bCopy)
	return len(bCopy), nil
}

func (s *AppState) TriggerAction(action string) {
	s.action <- action
}

func (s *AppState) GetInTray() bool {
	return s.InTray
}
