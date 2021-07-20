/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package gui

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/mysteriumnetwork/myst-launcher/native"

	"github.com/asaskevich/EventBus"
	"github.com/lxn/walk"
	"github.com/lxn/win"
)

type Config struct {
	AutoStart bool `json:"auto_start"`
	Enabled   bool `json:"enabled"`
}

type UIModel struct {
	InTray        bool
	InstallStage2 bool
	pipeListener  net.Listener
	CFG           Config

	Bus       EventBus.Bus
	waitClick chan int
	UIAction  chan string
	Icon      *walk.Icon
	mw        *walk.MainWindow

	state modalState

	// common
	StateDocker     runnableState
	StateContainer  runnableState
	VersionLatest   string
	VersionCurrent  string
	VersionUpToDate bool

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

var UI UIModel

func init() {
	UI.Bus = EventBus.New()
	UI.waitClick = make(chan int, 0)
	UI.UIAction = make(chan string, 1)
}

func (m *UIModel) Write(b []byte) (int, error) {
	// copy to avoid data corruption
	// see https://stackoverflow.com/a/20688698/4413696
	bCopy := make([]byte, len(b))
	copy(bCopy, b)

	UI.Bus.Publish("log", bCopy)
	return len(bCopy), nil
}

func (m *UIModel) Update() {
	UI.Bus.Publish("state-change")
}

func (m *UIModel) ShowMain() {
	win.ShowWindow(m.mw.Handle(), win.SW_SHOW)
	win.ShowWindow(m.mw.Handle(), win.SW_SHOWNORMAL)

	native.SwitchToThisWindow(m.mw.Handle(), false)
	win.SetWindowPos(m.mw.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
	win.SetWindowPos(m.mw.Handle(), win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
	win.SetWindowPos(m.mw.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
}

func (m *UIModel) SwitchState(s modalState) {
	m.state = s
	m.Update()
}

func (m *UIModel) BtnOnClick() {
	select {
	case m.waitClick <- 0:
	default:
		//fmt.Println("no message sent > BtnOnClick")
	}
}

func (m *UIModel) BtnUpgradeOnClick() {
	m.UIAction <- "upgrade"
}

func (m *UIModel) BtnDisableOnClick() {
	m.UIAction <- "disable"
}

func (m *UIModel) BtnEnableOnClick() {
	m.UIAction <- "enable"
}

func (m *UIModel) WaitDialogueComplete() {
	<-m.waitClick
}

func (m *UIModel) isExiting() bool {
	return UI.state == ModalStateInstallError
}

func (m *UIModel) ExitApp() {
	m.Bus.Publish("exit")

	m.mw.Synchronize(func() {
		walk.App().Exit(0)
	})
}

func (m *UIModel) OpenNodeUI() {
	native.ShellExecuteAndWait(
		0,
		"",
		"rundll32",
		"url.dll,FileProtocolHandler http://localhost:4449/",
		"",
		syscall.SW_NORMAL)
}

func (m *UIModel) ReadConfig() {
	f := os.Getenv("USERPROFILE") + "\\.myst_node_launcher"
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		// create default settings
		UI.CFG.AutoStart = true
		UI.CFG.Enabled = true
		m.SaveConfig()
		return
	}

	file, err := os.Open(f)
	if err != nil {
		return
	}

	// default value
	UI.CFG.Enabled = true
	json.NewDecoder(file).Decode(&UI.CFG)
}

func (m *UIModel) SaveConfig() {
	f := os.Getenv("USERPROFILE") + "\\.myst_node_launcher"
	file, err := os.Create(f)
	if err != nil {
		log.Println(err)
		return
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", " ")
	enc.Encode(&UI.CFG)
}

func (m *UIModel) ConfirmModal(title, message string) int {
	return walk.MsgBox(m.mw, title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
}
