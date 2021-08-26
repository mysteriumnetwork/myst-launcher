package app

import (
	"sync"

	"github.com/asaskevich/EventBus"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/myst"
)

type AppState struct {
	// flags
	InTray        bool
	InstallStage2 bool
	Config        model.Config

	Bus       EventBus.Bus
	WaitGroup sync.WaitGroup
	Action    chan string

	ImgVer myst.ImageVersionInfo

	// state
	didInstallation bool
}

func NewApp() *AppState {
	s := &AppState{}
	s.Action = make(chan string, 1)
	s.Bus = EventBus.New()
	s.ImgVer.ImageName = myst.GetImageName()

	return s
}

func (s *AppState) Write(b []byte) (int, error) {
	// copy to avoid data corruption
	// see https://stackoverflow.com/a/20688698/4413696
	bCopy := make([]byte, len(b))
	copy(bCopy, b)

	s.Bus.Publish("log", bCopy)
	return len(bCopy), nil
}

func (s *AppState) Publish(topic string, args ...interface{}) {
	s.Bus.Publish(topic, args...)
}

func (s *AppState) Subscribe(topic string, fn interface{}) error {
	return s.Bus.Subscribe(topic, fn)
}

func (s *AppState) Unsubscribe(topic string, fn interface{}) error {
	return s.Bus.Unsubscribe(topic, fn)
}

func (s *AppState) TriggerAction(action string) {
	s.Action <- action
}

func (s *AppState) GetInTray() bool {
	return s.InTray
}
func (s *AppState) GetConfig() *model.Config {
	return &s.Config
}

func (s *AppState) GetImageName() string {
	return s.ImgVer.ImageName
}
