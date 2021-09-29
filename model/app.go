package model

type AppInterface interface {
	TriggerAction(action string)
	GetInTray() bool
}
