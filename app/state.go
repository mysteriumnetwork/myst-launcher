/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package app

import (
	"fmt"
	"runtime"
	"sync"

	model2 "github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/platform"
)

type AppState struct {
	WaitGroup sync.WaitGroup // for graceful shutdown

	action chan string
	model  *model2.UIModel //gui.Model
	ui     model2.Gui_
	mgr    model2.PlatformManager
}

func NewApp() *AppState {
	s := &AppState{}
	s.action = make(chan string, 1)

	return s
}

func (s *AppState) initialize() error {
	runtime.LockOSThread()

	var err error
	s.mgr, err = platform.NewManager()
	return err
}

func (s *AppState) Shutdown() {
	// wait for SuperviseDockerNode to finish its work
	s.WaitGroup.Wait()
}

func (s *AppState) SetModel(ui *model2.UIModel) {
	s.model = ui
}

func (s *AppState) SetUI(ui model2.Gui_) {
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
	return s.model.Config.InitialState == model2.InitialStateNormalRun
}
