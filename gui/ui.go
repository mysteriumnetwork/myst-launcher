/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package gui

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
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
	WaitGroup     sync.WaitGroup

	Bus       EventBus.Bus
	waitClick chan int
	UIAction  chan string

	icon *walk.Icon
	dlg  *walk.MainWindow
	ni   *walk.NotifyIcon
	mw   *walk.MainWindow

	state    modalState
	wantExit bool

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
	EnableHyperV         bool
	InstallExecutable    bool
	RebootAfterWSLEnable bool
	DownloadFiles        bool
	InstallWSLUpdate     bool
	InstallDocker        bool
	CheckGroupMembership bool
	installationStatus   string

	ImageName string
}

var UI UIModel

func init() {
	UI.Bus = EventBus.New()
	UI.waitClick = make(chan int, 0)
	UI.UIAction = make(chan string, 1)
	UI.icon, _ = walk.NewIconFromResourceId(2)
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
	if !m.dlg.Visible() {
		win.ShowWindow(m.dlg.Handle(), win.SW_SHOW)
		win.ShowWindow(m.dlg.Handle(), win.SW_SHOWNORMAL)

		native.SwitchToThisWindow(m.dlg.Handle(), false)

		win.SetWindowPos(m.dlg.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
		win.SetWindowPos(m.dlg.Handle(), win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
		win.SetWindowPos(m.dlg.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
		return
	}

	if !win.IsIconic(m.dlg.Handle()) {
		win.ShowWindow(m.dlg.Handle(), win.SW_MINIMIZE)

	} else {
		win.ShowWindow(m.dlg.Handle(), win.SW_RESTORE)
	}
}

func (m *UIModel) SwitchState(s modalState) {
	m.state = s
	m.Update()
}

func (m *UIModel) BtnFinishOnClick() {
	fmt.Println("BtnFinishOnClick >", m.wantExit)
	if m.wantExit {
		m.ExitApp()
		return
	}

	select {
	case m.waitClick <- 0:
	default:
		fmt.Println("BtnFinishOnClick > not sent")
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

// returns channel close status
func (m *UIModel) WaitDialogueComplete() bool {
	_, ok := <-m.waitClick
	return ok
}

func (m *UIModel) SetWantExit() {
	m.wantExit = true
	m.Bus.Publish("want-exit")
}

func (m *UIModel) isExiting() bool {
	return UI.state == ModalStateInstallError
}

func (m *UIModel) ExitApp() {
	close(m.waitClick)

	m.Bus.Publish("exit")
	m.wantExit = true

	m.dlg.Synchronize(func() {
		m.dlg.Close()
	})
}

func (m *UIModel) OpenNodeUI() {
	m.UIAction <- "open-ui"

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
	return walk.MsgBox(m.dlg, title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
}

func (m *UIModel) YesNoModal(title, message string) int {
	return walk.MsgBox(m.dlg, title, message, walk.MsgBoxTopMost|walk.MsgBoxYesNo|walk.MsgBoxIconExclamation)
}

func (m *UIModel) Run() {
	m.mw.Run()
}

func (m *UIModel) ShowNotification() {
	err := m.ni.ShowCustom(
		"Mysterium Node successfully installed!",
		"Click this notification to open Node UI in browser",
		m.icon)

	if err != nil {
	}
}
