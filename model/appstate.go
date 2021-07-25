package model

import (
	"sync"

	"github.com/asaskevich/EventBus"
)

type Config struct {
	AutoStart bool `json:"auto_start"`
	Enabled   bool `json:"enabled"`
}

type AppState struct {
	// flags
	InTray        bool
	InstallStage2 bool

	Config    Config
	WaitGroup sync.WaitGroup
	Bus       EventBus.Bus
	// UIAction chan string

	// common
	StateDocker    runnableState
	StateContainer runnableState

	//VersionLatest   string
	//VersionCurrent  string
	//VersionUpToDate bool

	ImageName string
}

var State AppState

func init() {
	State.Bus = EventBus.New()
	//State.ImageName = myst.GetImageName()
}

func (s *AppState) IsRunning() bool {
	return s.StateContainer == RunnableStateRunning
}

func (s *AppState) SetStateDocker(r runnableState) {
	s.StateDocker = r
	s.Bus.Publish("model-change")
}

func (s *AppState) SetStateContainer(r runnableState) {
	s.StateContainer = r
	s.Bus.Publish("model-change")
}
