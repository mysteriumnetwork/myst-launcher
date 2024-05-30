/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package app

import (
	"fmt"
	"os"

	"github.com/mysteriumnetwork/myst-launcher/controller"
	"github.com/mysteriumnetwork/myst-launcher/model"
)

type AppState struct {
	model   *model.UIModel //gui.Model
	ui      model.Gui_
	ctrApp  model.Controller
	logfile *os.File
}

func NewApp(debug bool) *AppState {
	s := &AppState{}

	if debug {
		file, err := os.OpenFile("myst_launcher_log_.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		s.logfile = file
	}
	return s
}

func (s *AppState) StopAppController() {
	s.ctrApp.Shutdown()
}

func (s *AppState) StartAppController() {
	s.ctrApp = controller.NewController(s.model, s.ui, s)
	s.ctrApp.Start()
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
		if s.logfile != nil {
			s.logfile.Write(bCopy)
		}
	}

	s.model.Publish("log", bCopy)
	return len(bCopy), nil
}

func (s *AppState) TriggerAction(action string) {
	s.ctrApp.TriggerAction(action)
}

func (s *AppState) GetInTray() bool {
	return s.model.Config.InitialState == model.InitialStateNormalRun && s.model.FlagAutorun
}
