/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package gui

import (
	"syscall"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/native"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

type UIModel struct {
	waitClick chan int

	icon       *walk.Icon
	iconActive *walk.Icon

	dlg *walk.MainWindow
	ni  *walk.NotifyIcon
	mw  *walk.MainWindow

	state    modalState
	wantExit bool

	// common
	StateDocker      RunnableState
	StateContainer   RunnableState
	VersionLatest    string
	VersionCurrent   string
	HasUpdate        bool
	CurrentImgDigest string

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

	app                model.AppInterface
	LastNotificationID NotificationTypeID
}

type NotificationTypeID int

const (
	NotificationUpgrade          = NotificationTypeID(1)
	NotificationContainerStarted = NotificationTypeID(2)
)

var UI UIModel

func init() {
	UI.waitClick = make(chan int, 0)
	UI.icon, _ = walk.NewIconFromResourceId(2)
	UI.iconActive, _ = walk.NewIconFromResourceId(3)
}

func (m *UIModel) SetApp(app model.AppInterface) {
	UI.app = app
}

func (m *UIModel) Update() {
	UI.app.Publish("model-change")
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
	if m.wantExit {
		m.ExitApp()
		return
	}

	select {
	case m.waitClick <- 0:
	default:
	}
}

func (m *UIModel) BtnUpgradeOnClick() {
	m.AskDlg()
}

func (m *UIModel) BtnDisableOnClick() {
	m.app.TriggerAction("disable")
}

func (m *UIModel) BtnEnableOnClick() {
	m.app.TriggerAction("enable")
}

// returns channel close status
func (m *UIModel) WaitDialogueComplete() bool {
	_, ok := <-m.waitClick
	return ok
}

func (m *UIModel) SetWantExit() {
	m.wantExit = true
	m.app.Publish("want-exit")
}

func (m *UIModel) isExiting() bool {
	return UI.state == ModalStateInstallError
}

func (m *UIModel) ExitApp() {
	close(m.waitClick)

	m.app.Publish("exit")
	m.wantExit = true

	m.dlg.Synchronize(func() {
		m.dlg.Close()
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

func (m *UIModel) ConfirmModal(title, message string) int {
	return walk.MsgBox(m.dlg, title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
}

func (m *UIModel) YesNoModal(title, message string) int {
	return walk.MsgBox(m.dlg, title, message, walk.MsgBoxTopMost|walk.MsgBoxYesNo|walk.MsgBoxIconExclamation)
}

func (m *UIModel) ErrorModal(title, message string) int {
	return walk.MsgBox(UI.dlg, title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconError)
}

func (m *UIModel) Run() {
	m.mw.Run()
}

func (m *UIModel) ShowNotificationInstalled() {
	err := m.ni.ShowCustom(
		"Mysterium Node successfully installed!",
		"Click this notification to open Node UI in browser",
		m.icon)

	if err != nil {
	}
}

func (m *UIModel) ShowNotificationUpgrade() {
	err := m.ni.ShowCustom(
		"Upgrade available",
		"Click this notification to see details.",
		m.icon)

	if err != nil {
	}
}

func (m *UIModel) IsRunning() bool {
	return UI.StateContainer == RunnableStateRunning
}

func (m *UIModel) SetStateDocker(r RunnableState) {
	UI.StateDocker = r
	m.app.Publish("model-change")
}

func (m *UIModel) SetStateContainer(r RunnableState) {
	UI.StateContainer = r
	m.app.Publish("model-change")
	m.app.Publish("container-state")
}

func (m *UIModel) AskDlg() {
	gui.UpgradeDlg(UI.dlg)
}
