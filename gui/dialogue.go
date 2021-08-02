/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package gui

import (
	"fmt"
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

	//autoStart     *walk.CheckBox
	autoUpgrade   *walk.CheckBox
	btnOpenNodeUI *walk.PushButton

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
		Title:     "Mysterium Exit Node Launcher",
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

	enableMenu := func(enable bool) {
		//actionMainMenu.SetEnabled(enable)
		gui.actionEnable.SetEnabled(enable)
		gui.actionDisable.SetEnabled(enable)
		gui.actionUpgrade.SetEnabled(enable)
	}
	gui.currentView = frameState

	changeView := func(state modalState) {
		prev := gui.currentView
		gui.currentView = state
		if prev != state {
			UI.dlg.Children().At(int(prev)).SetVisible(false)
		}
		UI.dlg.Children().At(int(state)).SetVisible(true)
		UI.dlg.Children().At(int(state)).SetAlwaysConsumeSpace(true)
		UI.dlg.Children().At(int(state)).SetAlwaysConsumeSpace(false)
	}
	changeView(frameState)

	UI.app.Subscribe("model-change", func() {
		UI.dlg.Synchronize(func() {
			switch UI.state {
			case ModalStateInitial:
				enableMenu(true)
				changeView(frameState)

				gui.autoUpgrade.SetChecked(UI.app.GetConfig().AutoUpgrade)
				gui.lbDocker.SetText(UI.StateDocker.String())
				gui.lbContainer.SetText(UI.StateContainer.String())
				if !UI.app.GetConfig().Enabled {
					gui.lbContainer.SetText("Disabled")
				}
				gui.btnOpenNodeUI.SetEnabled(UI.IsRunning())
				//gui.lbVersionLatest.SetText(UI.VersionLatest)
				gui.lbVersionCurrent.SetText(UI.VersionCurrent)
				gui.lbVersionUpdatesAvail.SetText("-")
				if UI.VersionLatest != "" && !UI.VersionUpToDate {
					gui.lbVersionUpdatesAvail.SetText(`<a id="upgrade">Yes !</a> - click to see details`)
				}
				gui.btnOpenNodeUI.SetFocus()

			case ModalStateInstallNeeded:
				enableMenu(false)
				changeView(frameInsNeed)
				gui.btnBegin.SetEnabled(true)

			case ModalStateInstallInProgress:
				enableMenu(false)
				changeView(frameIns)
				gui.btnFinish.SetEnabled(false)

			case ModalStateInstallFinished:
				enableMenu(false)
				changeView(frameIns)
				gui.btnFinish.SetEnabled(true)
				gui.btnFinish.SetText("Finish")

			case ModalStateInstallError:
				changeView(frameIns)
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
		})
	})

	UI.dlg.Starting().Attach(func() {

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

func (g *Gui) Ask() {
	var (
		dialog             *walk.Dialog
		acceptPB, cancelPB *walk.PushButton
		lbVersionCurrent   *walk.Label
		lbVersionLatest    *walk.Label
	)

	err := Dialog{
		AssignTo:      &dialog,
		Title:         "Would you like to upgrade?",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{400, 175},
		Icon:          UI.icon,

		Layout: Grid{
			Columns: 2,
		},
		Children: []Widget{
			VSpacer{ColumnSpan: 2},
			Label{
				Text: "Docker Hub image name",
			},
			Label{
				Text: UI.app.GetImageName(),
			},
			Label{
				Text: "Node version installed",
			},
			Label{
				Text:     "-",
				AssignTo: &lbVersionCurrent,
			},
			Label{
				Text: "Node version latest",
			},
			Label{
				Text:     "-",
				AssignTo: &lbVersionLatest,
			},
			VSpacer{ColumnSpan: 2},
			Composite{
				ColumnSpan: 2,
				Layout:     HBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &acceptPB,
						Text:     "Yes",
						OnClicked: func() {
							dialog.Accept()
							UI.app.TriggerAction("upgrade")
						},
					},
					PushButton{
						AssignTo: &cancelPB,
						Text:     "No",
						OnClicked: func() {
							dialog.Cancel()
						},
					},
				},
			},
		},
	}.Create(UI.dlg)
	if err != nil {
		fmt.Println(err)
	}
	refresh := func() {
		lbVersionCurrent.SetText(UI.VersionCurrent)
		lbVersionLatest.SetText(UI.VersionLatest)
		acceptPB.SetEnabled(!UI.VersionUpToDate)
	}

	dialog.Show()
	dialog.SetX(UI.dlg.X() + 300)
	refresh()

	UI.app.Subscribe("model-change", refresh)
}
