package app

import (
	"fmt"
	"sync"

	gui_win32 "github.com/mysteriumnetwork/myst-launcher/gui-win32"

	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/myst"
)

type AppState struct {
	// flags
	InTray        bool
	InstallStage2 bool

	Config    model.Config
	WaitGroup sync.WaitGroup // for graceful shutdown

	action chan string
	mod    gui.UIModelInterface
	ui     *gui_win32.Gui

	ImgVer myst.ImageVersionInfo

	// state
	didInstallation bool
}

func NewApp() *AppState {
	s := &AppState{}
	s.action = make(chan string, 1)
	s.ImgVer.ImageName = myst.GetImageName()
	return s
}

func (s *AppState) SetModel(ui gui.UIModelInterface) {
	s.mod = ui
}

func (s *AppState) SetUI(ui *gui_win32.Gui) {
	s.ui = ui
}

func (s *AppState) Write(b []byte) (int, error) {
	// copy to avoid data corruption
	// see https://stackoverflow.com/a/20688698/4413696
	bCopy := make([]byte, len(b))
	copy(bCopy, b)

	s.mod.Publish("log", bCopy)
	return len(bCopy), nil
}

func (s *AppState) TriggerAction(action string) {
	fmt.Println("TriggerAction", action)
	s.action <- action
}

func (s *AppState) GetInTray() bool {
	return s.InTray
}
func (s *AppState) GetConfig() *model.Config {
	return &s.Config
}
