/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package app

import (
	"fmt"

	"github.com/mysteriumnetwork/myst-launcher/model"
)

type Controller interface {
	SetApp(a *AppState)
	Start()
}

type AppState struct {
	action chan string

	model  *model.UIModel //gui.Model
	ui     model.Gui_
	ctrApp Controller
}

func NewApp() *AppState {
	s := &AppState{}

	// s.action = make(chan string, 1)
	s.action = make(chan string) // unbuffered, synchronous
	return s
}

func (s *AppState) SetAppController(c Controller) {
	if s.ctrApp != nil {
		s.action <- "stop" // wait prev. controller to finish
	}
	s.ctrApp = c
	c.SetApp(s)
}

func (s *AppState) SetModel(ui *model.UIModel) {
	s.model = ui
}

func (s *AppState) GetModel() *model.UIModel {
	return s.model
}

func (s *AppState) GetUI() model.Gui_ {
	return s.ui
}

func (s *AppState) GetAction() chan string {
	return s.action
}

func (s *AppState) SetUI(ui model.Gui_) {
	s.ui = ui
}

func (s *AppState) Write(b []byte) (int, error) {
	// copy to avoid data corruption
	// see https://stackoverflow.com/a/20688698/4413696
	bCopy := make([]byte, len(b))
	copy(bCopy, b)

	if s.model.DuplicateLogToConsole {
		fmt.Print(string(bCopy))
	}

	s.model.Publish("log", bCopy)
	return len(bCopy), nil
}

func (s *AppState) TriggerAction(action string) {
	s.action <- action
}

func (s *AppState) Stop() {
	s.action <- "stop"
}

func (s *AppState) GetInTray() bool {
	return s.model.Config.InitialState == model.InitialStateNormalRun
}
