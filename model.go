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
	"os/exec"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

type Model struct {
	state        modalState
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
	btnCmd        *walk.PushButton
	btnOpenNodeUI *walk.PushButton

	dlg chan int
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

var model Model

func init() {
	model.dlg = make(chan int)
}

func (m *Model) ShowMain() {
	win.ShowWindow(m.mw.Handle(), win.SW_SHOW)
	win.ShowWindow(m.mw.Handle(), win.SW_SHOWNORMAL)

	SwitchToThisWindow(m.mw.Handle(), false)
	win.SetWindowPos(m.mw.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
	win.SetWindowPos(m.mw.Handle(), win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
	win.SetWindowPos(m.mw.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
}

func (m *Model) SetState(s modalState) {
	m.state = s
	m.Invalidate()
}

const frameI = 1
const frameS = 2

func (m *Model) Invalidate() {
	switch m.state {
	case initial:
		m.mw.Children().At(frameI).SetVisible(false)
		m.mw.Children().At(frameS).SetVisible(true)
	case installNeeded:
		m.mw.Children().At(frameI).SetVisible(true)
		m.mw.Children().At(frameS).SetVisible(false)
		m.HideProgress()

		m.btnCmd.SetEnabled(true)
		m.btnCmd.SetText("Install")
		m.btnCmd.SetFocus()

		m.lbInstallationState.SetText("Docker desktop is required to run exit node.")
		m.lbInstallationState2.SetText("Press button to begin installation.")

		m.lbDocker.SetText("OK")
	case installInProgress:
		m.btnCmd.SetEnabled(false)
		m.lbInstallationState.SetText("Downloading installation packages.")
		m.lbInstallationState2.SetText("-")
	case installFinished:
		m.lbInstallationState.SetText("Installation successfully finished!")
		m.btnCmd.SetEnabled(true)
		m.btnCmd.SetText("Finish !")
	case installError:
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
	return model.state == installError
}

func (m *Model) openNodeUI() {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:4449")
	if err := cmd.Start(); err != nil {
	}
}
