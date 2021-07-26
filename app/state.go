package app

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/asaskevich/EventBus"

	"github.com/mysteriumnetwork/myst-launcher/model"
)

type AppState struct {
	// flags
	InTray        bool
	InstallStage2 bool
	Config        model.Config

	Bus       EventBus.Bus
	WaitGroup sync.WaitGroup
	Action    chan string

	ImageName string
}

func NewApp() *AppState {
	s := &AppState{}
	s.Action = make(chan string, 1)
	s.Bus = EventBus.New()
	return s
}

func (s *AppState) ReadConfig() {
	f := os.Getenv("USERPROFILE") + "\\.myst_node_launcher"
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		// create default settings
		s.Config.AutoStart = true
		s.Config.Enabled = true
		s.SaveConfig()
		return
	}

	file, err := os.Open(f)
	if err != nil {
		return
	}

	// default value
	s.Config.Enabled = true
	json.NewDecoder(file).Decode(&s.Config)
}

func (s *AppState) SaveConfig() {
	f := os.Getenv("USERPROFILE") + "\\.myst_node_launcher"
	file, err := os.Create(f)
	if err != nil {
		log.Println(err)
		return
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", " ")
	enc.Encode(&s.Config)
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
	return s.ImageName
}
