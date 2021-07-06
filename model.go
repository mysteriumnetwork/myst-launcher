package main

import (
	"fmt"

	"github.com/lxn/walk"
)

type Model struct {
	mw *walk.MainWindow
	lv *LogView

	state int

	// docker
	lbDocker    *walk.Label
	lbContainer *walk.Label

	// inst
	lbInstallationState *walk.Label

	// common
	btnCmd  *walk.PushButton
	btnCmd2 *walk.PushButton

	dlg chan int
}

const (
	INSTALL_NEED = -1
	INSTALL_     = -2
	INSTALL_FIN  = -3
)

var mod Model

func init() {
	mod.dlg = make(chan int)
}

func (m *Model) SetState(s int) {
	m.state = s
	m.Invalidate()
}

const frameI = 1
const frameS = 2

func (m *Model) Invalidate() {
	if m.state == 0 {
		mod.mw.Children().At(frameI).SetVisible(false)
		mod.mw.Children().At(frameS).SetVisible(true)

	}
	if m.state == INSTALL_NEED {
		mod.mw.Children().At(frameI).SetVisible(true)
		mod.mw.Children().At(frameS).SetVisible(false)

		mod.btnCmd.SetEnabled(true)
		mod.btnCmd.SetText("Install")
		mod.btnCmd.SetFocus()

		mod.lbInstallationState.SetText("Docker desktop is requred to run exit node.\r\n" +
			"Press button to begin installation.")

		mod.lbDocker.SetText("OK")
	}
	if m.state == INSTALL_ {
		mod.btnCmd.SetEnabled(false)
		mod.lbInstallationState.SetText("Downloading installation packages.\r\n" + "_")
	}
	if m.state == INSTALL_FIN {
		m.lbInstallationState.SetText("Installation successfully finished!\r\n_")
		mod.btnCmd.SetEnabled(true)
		mod.btnCmd.SetText("Finish !")
	}
}

func (m *Model) BtnOnClick() {
	fmt.Println("BtnOnClick", m.state)

	if m.state == INSTALL_FIN {
		mod.SetState(0)
		m.dlg <- 0
	}
	if m.state == INSTALL_NEED {
		m.SetState(INSTALL_)
		m.dlg <- 0
	}

}

func (m *Model) WaitDialogueComplete() {
	//log.Println("WaitDialogueComplete>")
	<-m.dlg
}

func (m *Model) PrintProgress(progress int) {
	m.lv.AppendText(fmt.Sprintf("Download %d %%\r\n", progress))
}
