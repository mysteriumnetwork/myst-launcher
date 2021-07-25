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
	lbDocker         *walk.Label
	lbContainer      *walk.Label
	lbVersionLatest  *walk.Label
	lbVersionCurrent *walk.Label

	autoStart     *walk.CheckBox
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
}

var gui Gui

func CreateDialogue() {

	if err := (MainWindow{
		AssignTo:  &UI.dlg,
		Title:     "Mysterium Exit Node Launcher",
		MinSize:   Size{320, 440},
		Size:      Size{320, 640},
		MaxSize:   Size{320, 640},
		Icon:      UI.icon,
		MenuItems: gui.menu(),
		Layout:    VBox{
			//MarginsZero: true,
		},

		Children: []Widget{
			ImageView{
				AssignTo:  &gui.iv,
				Alignment: AlignHNearVFar,
			},
			gui.instWiz(),
			gui.installationDlg(),
			gui.stateDlg(),
			//VSpacer{RowSpan: 1},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	icon, err := walk.NewIconFromResourceIdWithSize(2, walk.Size{
		Width:  64,
		Height: 64,
	})
	if err == nil {
		i, err := walk.ImageFrom(icon)
		if err == nil {
			_ = i
			gui.iv.SetImage(i)
		}
	}

	if UI.app.GetInTray() {
		UI.dlg.SetVisible(false)
	}

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
	UI.app.Subscribe("show-dlg", func(d string, err error) {
		switch d {
		case "is-up-to-date":
			walk.MsgBox(UI.dlg, "Update", "Node is up to date.", walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconInformation)

		case "error":
			txt := err.Error() + "\r\n" + "Application will exit now"
			walk.MsgBox(UI.dlg, "Application error", txt, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconError)
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
	}
	changeView(frameState)

	UI.app.Subscribe("model-change", func() {
		UI.dlg.Synchronize(func() {
			switch UI.state {
			case ModalStateInitial:
				enableMenu(true)
				changeView(frameState)

				gui.autoStart.SetChecked(UI.app.GetConfig().AutoStart)

				gui.lbDocker.SetText(UI.StateDocker.String())
				gui.lbContainer.SetText(UI.StateContainer.String())
				if !UI.app.GetConfig().Enabled {
					gui.lbContainer.SetText("Disabled")
				}

				gui.btnOpenNodeUI.SetEnabled(UI.IsRunning())
				gui.lbVersionLatest.SetText(UI.VersionLatest)
				gui.lbVersionCurrent.SetText(UI.VersionCurrent)

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

	// prevent closing the app
	UI.dlg.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if UI.wantExit {
			walk.App().Exit(0)
		}
		*canceled = true
		UI.dlg.Hide()
	})
}
