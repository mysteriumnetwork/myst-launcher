/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package main

import (
	"fmt"
	"net"

	"github.com/lxn/win"

	"github.com/lxn/walk"
)

type Model struct {
	state        int
	inTray       bool
	pipeListener net.Listener

	icon *walk.Icon
	mw   *walk.MainWindow
	lv   *LogView

	// docker
	lbDocker    *walk.Label
	lbContainer *walk.Label

	// inst
	lbInstallationState  *walk.Label
	lbInstallationState2 *walk.Label
	progressBar          *walk.ProgressBar

	// common
	btnCmd  *walk.PushButton
	btnCmd2 *walk.PushButton

	dlg chan int
}

const (
	// state
	ST_STATUS_FRAME       = 0
	ST_INSTALL_NEED       = -1
	ST_INSTALL_INPROGRESS = -2
	ST_INSTALL_FIN        = -3
	ST_INSTALL_ERR        = -4
)

var mod Model

func init() {
	mod.dlg = make(chan int)
}

func (m *Model) ShowMain() {
	win.ShowWindow(m.mw.Handle(), win.SW_SHOW)
	win.ShowWindow(m.mw.Handle(), win.SW_SHOWNORMAL)

	SwitchToThisWindow(m.mw.Handle(), false)
	win.SetWindowPos(m.mw.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
	win.SetWindowPos(m.mw.Handle(), win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
	win.SetWindowPos(m.mw.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
}

func (m *Model) SetState(s int) {
	m.state = s
	m.Invalidate()
}

const frameI = 1
const frameS = 2

func (m *Model) Invalidate() {
	if m.state == 0 {
		m.mw.Children().At(frameI).SetVisible(false)
		m.mw.Children().At(frameS).SetVisible(true)
	}
	if m.state == ST_INSTALL_NEED {
		m.mw.Children().At(frameI).SetVisible(true)
		m.mw.Children().At(frameS).SetVisible(false)
		m.HideProgress()

		m.btnCmd.SetEnabled(true)
		m.btnCmd.SetText("Install")
		m.btnCmd.SetFocus()

		m.lbInstallationState.SetText("Docker desktop is required to run exit node.")
		m.lbInstallationState2.SetText("Press button to begin installation.")

		m.lbDocker.SetText("OK")
	}
	if m.state == ST_INSTALL_INPROGRESS {
		m.btnCmd.SetEnabled(false)
		m.lbInstallationState.SetText("Downloading installation packages.")
		m.lbInstallationState2.SetText("-")
	}
	if m.state == ST_INSTALL_FIN {
		m.lbInstallationState.SetText("Installation successfully finished!")
		m.btnCmd.SetEnabled(true)
		m.btnCmd.SetText("Finish !")
	}
	if m.state == ST_INSTALL_ERR {
		m.lbInstallationState.SetText("Installation failed")
		m.btnCmd.SetEnabled(true)
		m.btnCmd.SetText("Exit installer")
	}
}

func (m *Model) BtnOnClick() {
	m.dlg <- 0
}

func (m *Model) WaitDialogueComplete() {
	<-m.dlg
}

func (m *Model) HideProgress() {
	m.progressBar.SetVisible(false)
}

func (m *Model) PrintProgress(progress int) {
	m.lv.AppendText(fmt.Sprintf("Download %d %%\r\n", progress))
	m.progressBar.SetVisible(true)
	m.progressBar.SetValue(progress)
}

func (m *Model) isExiting() bool {
	return mod.state == ST_INSTALL_ERR
}
