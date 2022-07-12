/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package model

type App interface {
	TriggerAction(action string)
	GetInTray() bool
}

type Controller interface {
	SetApp(a AppState)
	Start()
	GetCaps() int
	GetFinished() bool
}

const (
	ActionCheck      = "check"
	ActionUpgrade    = "upgrade"
	ActionRestart    = "restart"
	ActionEnable     = "enable"
	ActionDisable    = "disable"
	ActionStopRunner = "stop-runner"
	ActionStop       = "stop"
)

type AppState interface {
	StartAppController()
	StopAppController()

	GetModel() *UIModel
	// SetModel(ui *UIModel)
	GetAction() chan string
	GetUI() Gui_
	// SetUI(ui Gui_)

	// Write(b []byte) (int, error)
	// TriggerAction(action string)
	// GetInTray() bool
}
