/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package app

import (
	"fmt"
	model2 "github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/wmi"
	"os"
	"runtime"
	"sync"
)

var f *os.File

func init() {

	//f, _ = os.Create("log_" + time.Now().GoString())
}

type AppState struct {
	// flags
	InTray        bool
	InstallStage1 bool
	InstallStage2 bool

	WaitGroup sync.WaitGroup // for graceful shutdown

	action chan string
	model  *model2.UIModel //gui.Model
	ui     model2.Gui_

	// state
	didDockerInstallation bool
	wmi                   *wmi.Manager
}

func NewApp() *AppState {
	s := &AppState{}
	s.action = make(chan string, 1)

	return s
}

func (s *AppState) initialize() error {
	runtime.LockOSThread()

	var err error
	s.wmi, err = wmi.NewSysManager()
	if err != nil {
		fmt.Println(">>>", err)
	}
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
		//f.WriteString(string(bCopy))
		//f.Sync()
	}

	s.model.Publish("log", bCopy)
	return len(bCopy), nil
}

func (s *AppState) TriggerAction(action string) {
	s.action <- action
}

func (s *AppState) GetInTray() bool {
	return s.InTray
}
