/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package gui

import (
	"encoding/json"
	"github.com/mysteriumnetwork/myst-launcher/native"
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
	CFG           Config

	Bus       EventBus.Bus
	waitClick chan int
	Icon      *walk.Icon
	mw        *walk.MainWindow

	state modalState

	// common
	StateDocker    runnableState
	StateContainer runnableState

	// inst
	CheckWindowsVersion  bool
	CheckVTx             bool
	EnableWSL            bool
	InstallExecutable    bool
	RebootAfterWSLEnable bool
	DownloadFiles        bool
	InstallWSLUpdate     bool
	InstallDocker        bool
	CheckGroupMembership bool
	installationStatus   string
}

var SModel Model

func init() {
	SModel.Bus = EventBus.New()
	SModel.waitClick = make(chan int, 0)
}

func (m *Model) Write(p []byte) (int, error) {
	SModel.Bus.Publish("log", p)
	return len(p), nil
}

func (m *Model) TriggerUpdate() {
	SModel.Bus.Publish("state-change")
}

func (m *Model) ShowMain() {
	win.ShowWindow(m.mw.Handle(), win.SW_SHOW)
	win.ShowWindow(m.mw.Handle(), win.SW_SHOWNORMAL)

	native.SwitchToThisWindow(m.mw.Handle(), false)
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
	return SModel.state == ModalStateInstallError
}

func (m *Model) ExitApp() {
	m.Bus.Publish("exit")

	m.mw.Synchronize(func() {
		walk.App().Exit(0)
	})
}

func (m *Model) OpenNodeUI() {
	exe := "rundll32"
	cmdArgs := "url.dll,FileProtocolHandler http://localhost:4449/"
	native.ShellExecuteAndWait(0, "", exe, cmdArgs, "", syscall.SW_NORMAL)
}

func (m *Model) ReadConfig() {
	f := os.Getenv("USERPROFILE") + "\\.myst_node_launcher"
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		return
	}

	file, err := os.Open(f)
	if err != nil {
		return
	}
	json.NewDecoder(file).Decode(&SModel.CFG)
}

func (m *Model) SaveConfig() {
	f := os.Getenv("USERPROFILE") + "\\.myst_node_launcher"
	file, err := os.Create(f)
	if err != nil {
		log.Println(err)
		return
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", " ")
	err = enc.Encode(&SModel.CFG)
	log.Println(err)
}

func (m *Model) ConfirmModal(title, message string) int {
	return walk.MsgBox(m.mw, title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
}
