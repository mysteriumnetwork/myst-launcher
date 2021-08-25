/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package gui

import (
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

const (
	ofs          = 0
	frameImage_  = 0 + ofs
	frameInsNeed = 1 + ofs
	frameIns     = 2 + ofs
	frameState   = 3 + ofs
)

type Gui struct {
	actionFileMenu *walk.Action
	actionMainMenu *walk.Action

	actionUpgrade *walk.Action
	actionEnable  *walk.Action
	actionDisable *walk.Action

	// common
	lbDocker              *walk.Label
	lbContainer           *walk.Label
	lbVersionLatest       *walk.Label
	lbVersionCurrent      *walk.Label
	lbVersionUpdatesAvail *walk.LinkLabel

	autoUpgrade          *walk.CheckBox
	manualPortForwarding *walk.CheckBox
	lbNetworkMode        *walk.LinkLabel
	btnOpenNodeUI        *walk.PushButton

	// install
	lbInstallationStatus *walk.TextEdit
	btnBegin             *walk.PushButton

	checkWindowsVersion  *walk.CheckBox
	checkVTx             *walk.CheckBox
	enableWSL            *walk.CheckBox
	enableHyperV         *walk.CheckBox
	installExecutable    *walk.CheckBox
	rebootAfterWSLEnable *walk.CheckBox
	downloadFiles        *walk.CheckBox
	installWSLUpdate     *walk.CheckBox
	installDocker        *walk.CheckBox
	checkGroupMembership *walk.CheckBox

	btnFinish *walk.PushButton
	iv        *walk.ImageView

	currentView modalState
	ico         *walk.Icon
	icoActive   *walk.Icon
}

var gui Gui

func CreateDialogue() {
	if err := (MainWindow{
		Visible:   false,
		AssignTo:  &UI.dlg,
		Title:     "Mysterium Node Launcher",
		MinSize:   Size{380, 640},
		Size:      Size{380, 640},
		Icon:      UI.icon,
		MenuItems: gui.menu(),
		Layout:    VBox{},

		Children: []Widget{
			ImageView{
				AssignTo:  &gui.iv,
				Alignment: AlignHNearVFar,
			},
			gui.installationWelcome(),
			gui.installationDlg(),
			gui.stateDlg(),
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}
	UI.dlg.SetVisible(!UI.app.GetInTray())

	var err error
	gui.ico, err = walk.NewIconFromResourceIdWithSize(2, walk.Size{
		Width:  64,
		Height: 64,
	})
	if err != nil {
		log.Fatal(err)
	}
	gui.icoActive, err = walk.NewIconFromResourceIdWithSize(3, walk.Size{
		Width:  64,
		Height: 64,
	})
	if err != nil {
		log.Fatal(err)
	}
	gui.SetImage()
	UI.app.Subscribe("container-state", func() {
		UI.dlg.Synchronize(func() {
			gui.SetImage()
		})
	})

	// Events
	UI.app.Subscribe("want-exit", func() {
		UI.dlg.Synchronize(func() {
			gui.btnFinish.SetEnabled(true)
		})
	})

	UI.app.Subscribe("log", func(p []byte) {
		switch UI.state {
		case ModalStateInstallInProgress, ModalStateInstallError, ModalStateInstallFinished:
			UI.dlg.Synchronize(func() {
				gui.lbInstallationStatus.AppendText(string(p) + "\r\n")
			})
		}
	})
	gui.currentView = frameState
	gui.changeView(frameState)

	UI.dlg.Activating().Attach(func() {
		if UI.dlg.Visible() {
			// refresh on window restore
			UI.dlg.Synchronize(func() {
				gui.refresh()
			})
		}
	})
	UI.app.Subscribe("model-change", func() {
		if !UI.dlg.Visible() {
			return
		}
		UI.dlg.Synchronize(func() {
			gui.refresh()
		})
	})

	// prevent closing the app
	UI.dlg.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if UI.wantExit {
			walk.App().Exit(0)
		}
		*canceled = true
		UI.dlg.Hide()
	})
}

func (g *Gui) enableMenu(enable bool) {
	//actionMainMenu.SetEnabled(enable)
	gui.actionEnable.SetEnabled(enable)
	gui.actionDisable.SetEnabled(enable)
	gui.actionUpgrade.SetEnabled(enable)
}

func (g *Gui) changeView(state modalState) {
	prev := gui.currentView
	gui.currentView = state
	if prev != state {
		UI.dlg.Children().At(int(prev)).SetVisible(false)
	}
	UI.dlg.Children().At(int(state)).SetVisible(true)
	UI.dlg.Children().At(int(state)).SetAlwaysConsumeSpace(true)
	UI.dlg.Children().At(int(state)).SetAlwaysConsumeSpace(false)
}

func (g *Gui) refresh() {
	switch UI.state {
	case ModalStateInitial:
		g.enableMenu(true)
		g.changeView(frameState)

		gui.autoUpgrade.SetChecked(UI.app.GetConfig().AutoUpgrade)
		if !UI.app.GetConfig().EnablePortForwarding {
			gui.lbNetworkMode.SetText(`<a id="net">Port restricted cone NAT</a>`)
		} else {
			gui.lbNetworkMode.SetText(`<a id="net">Manual port forwarding</a>`)
		}

		gui.lbDocker.SetText(UI.StateDocker.String())
		gui.lbContainer.SetText(UI.StateContainer.String())
		if !UI.app.GetConfig().Enabled {
			gui.lbContainer.SetText("Disabled")
		}
		gui.btnOpenNodeUI.SetEnabled(UI.IsRunning())
		//gui.lbVersionLatest.SetText(UI.VersionLatest)
		gui.lbVersionCurrent.SetText(UI.VersionCurrent)
		gui.lbVersionUpdatesAvail.SetText("-")
		if UI.HasUpdate {
			gui.lbVersionUpdatesAvail.SetText(`<a id="upgrade">Yes !</a> - click to see details`)
		}
		gui.btnOpenNodeUI.SetFocus()

	case ModalStateInstallNeeded:
		g.enableMenu(false)
		g.changeView(frameInsNeed)
		gui.btnBegin.SetEnabled(true)

	case ModalStateInstallInProgress:
		g.enableMenu(false)
		g.changeView(frameIns)
		gui.btnFinish.SetEnabled(false)

	case ModalStateInstallFinished:
		g.enableMenu(false)
		g.changeView(frameIns)
		gui.btnFinish.SetEnabled(true)
		gui.btnFinish.SetText("Finish")

	case ModalStateInstallError:
		g.changeView(frameIns)
		gui.btnFinish.SetEnabled(true)
		gui.btnFinish.SetText("Exit installer")
	}

	switch UI.state {
	case ModalStateInstallInProgress, ModalStateInstallFinished, ModalStateInstallError:
		gui.checkWindowsVersion.SetChecked(UI.CheckWindowsVersion)
		gui.checkVTx.SetChecked(UI.CheckVTx)
		gui.enableWSL.SetChecked(UI.EnableWSL)
		gui.enableHyperV.SetChecked(UI.EnableHyperV)
		gui.installExecutable.SetChecked(UI.InstallExecutable)
		gui.rebootAfterWSLEnable.SetChecked(UI.RebootAfterWSLEnable)
		gui.downloadFiles.SetChecked(UI.DownloadFiles)
		gui.installWSLUpdate.SetChecked(UI.InstallWSLUpdate)
		gui.installDocker.SetChecked(UI.InstallDocker)
		gui.checkGroupMembership.SetChecked(UI.CheckGroupMembership)
	}
}

func (g *Gui) SetImage() {
	ico := gui.ico
	if UI.StateContainer == RunnableStateRunning {
		ico = gui.icoActive
	}
	img, err := walk.ImageFrom(ico)
	if err != nil {
		return
	}
	gui.iv.SetImage(img)
}
