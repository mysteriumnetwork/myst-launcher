package model

type Config struct {
	AutoStart bool `json:"auto_start"`
	Enabled   bool `json:"enabled"`
}

type AppInterface interface {
	ReadConfig()
	SaveConfig()

	Publish(topic string, args ...interface{})
	Subscribe(topic string, fn interface{}) error
	TriggerAction(action string)

	GetInTray() bool
	GetConfig() *Config
	GetImageName() string
}
