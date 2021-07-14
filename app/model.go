/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package app

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/asaskevich/EventBus"
	"github.com/lxn/walk"
	"github.com/lxn/win"
)

type Config struct {
	AutoStart bool `json:"auto_start"`
}

type Model struct {
	InTray        bool
	InstallStage2 bool
	pipeListener  net.Listener
	cfg           Config

	bus       EventBus.Bus
	waitClick chan int
	Icon      *walk.Icon
	mw        *walk.MainWindow

	state modalState

	// common
	stateDocker    runnableState
	stateContainer runnableState

	// inst
	checkWindowsVersion  bool
	checkVTx             bool
	enableWSL            bool
	installExecutable    bool
	rebootAfterWSLEnable bool
	downloadFiles        bool
	installWSLUpdate     bool
	installDocker        bool
	checkGroupMembership bool
	installationStatus   string
}

type modalState int

const (
	// state
	initial           modalState = 0
	installNeeded     modalState = -1
	installInProgress modalState = -2
	installFinished   modalState = -3
	installError      modalState = -4
)

type runnableState int

const (
	stateUnknown    runnableState = 0
	stateStarting   runnableState = 1
	stateRunning    runnableState = 2
	stateInstalling runnableState = 3
)

var SModel Model

func init() {
	SModel.bus = EventBus.New()
	SModel.waitClick = make(chan int, 0)
}

func (m *Model) Write(p []byte) (int, error) {
	SModel.bus.Publish("log", p)
	return len(p), nil
}

func (m *Model) TriggerUpdate() {
	SModel.bus.Publish("state-change")
}

func (m *Model) ShowMain() {
	win.ShowWindow(m.mw.Handle(), win.SW_SHOW)
	win.ShowWindow(m.mw.Handle(), win.SW_SHOWNORMAL)

	SwitchToThisWindow(m.mw.Handle(), false)
	win.SetWindowPos(m.mw.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
	win.SetWindowPos(m.mw.Handle(), win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
	win.SetWindowPos(m.mw.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
}

func (m *Model) SwitchState(s modalState) {
	m.state = s
	m.TriggerUpdate()
}

func (m *Model) BtnOnClick() {
	select {
	case m.waitClick <- 0:
	default:
		//fmt.Println("no message sent > BtnOnClick")
	}
}

func (m *Model) WaitDialogueComplete() {
	<-m.waitClick
}

func (m *Model) SetProgress(progress int) {
}

func (m *Model) isExiting() bool {
	return SModel.state == installError
}

func (m *Model) ExitApp() {
	m.bus.Publish("exit")

	m.mw.Synchronize(func() {
		walk.App().Exit(0)
	})
}

func (m *Model) openNodeUI() {
	exe := "rundll32"
	cmdArgs := "url.dll,FileProtocolHandler http://localhost:4449/"
	_ShellExecuteAndWait(0, "", exe, cmdArgs, "", syscall.SW_NORMAL)
}

func (m *Model) readConfig() {
	f := os.Getenv("USERPROFILE") + "\\.myst_node_launcher"
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		return
	}

	file, err := os.Open(f)
	if err != nil {
		return
	}
	json.NewDecoder(file).Decode(&SModel.cfg)
}

func (m *Model) saveConfig() {
	f := os.Getenv("USERPROFILE") + "\\.myst_node_launcher"
	file, err := os.Create(f)
	if err != nil {
		log.Println(err)
		return
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", " ")
	err = enc.Encode(&SModel.cfg)
	log.Println(err)
}
