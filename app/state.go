/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package app

import (
	"fmt"

	"github.com/mysteriumnetwork/myst-launcher/controller/docker"
	"github.com/mysteriumnetwork/myst-launcher/controller/native"
	"github.com/mysteriumnetwork/myst-launcher/model"
)

type AppState struct {
	action chan string

	model *model.UIModel //gui.Model
	ui    model.Gui_

	ctrApp model.Controller
}

func NewApp() *AppState {
	s := &AppState{}

	s.action = make(chan string) // unbuffered, synchronous
	return s
}

func (s *AppState) StopAppController() {
	if !s.ctrApp.GetFinished() {
		s.TriggerAction(model.ActionStop)
	}
}

func (s *AppState) StartAppController() {
	setUIController := func() {
		var nc model.Controller

		switch s.model.Config.Backend {
		case "native":
			nc = native.NewController()
		case "docker":
			nc = docker.NewController()
		}
		s.setAppController(nc)
		go nc.Start()
	}
	setUIController()

	s.model.Bus2.Subscribe("backend", setUIController)
}

func (s *AppState) setAppController(c model.Controller) {
	if s.ctrApp != nil {
		s.action <- model.ActionStopRunner // wait prev. controller to finish
	}

	s.ctrApp = c
	s.model.Caps = c.GetCaps()
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

func (s *AppState) GetInTray() bool {
	return s.model.Config.InitialState == model.InitialStateNormalRun
}
