package model

type AppInterface interface {
	//Publish(topic string, args ...interface{})
	//Subscribe(topic string, fn interface{}) error
	//Unsubscribe(topic string, fn interface{}) error

	TriggerAction(action string)
	GetInTray() bool

	//GetConfig() *Config
	//GetImageName() string
}
