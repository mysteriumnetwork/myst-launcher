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

	"github.com/lxn/win"

	"github.com/mysteriumnetwork/myst-launcher/model"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

const (
	ofs = 0

	frameImage_  = 0 + ofs
	frameInsNeed = 1 + ofs
	frameIns     = 2 + ofs
	frameState   = 3 + ofs
)

type MWState struct {
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

var Mw MWState

func (mw *MWState) CreateDialogue() {
	UI.ReadConfig()

	if err := (MainWindow{
		AssignTo: &UI.dlg,
		Title:    "Mysterium Exit Node Launcher",
		MinSize:  Size{320, 440},
		Size:     Size{320, 640},
		MaxSize:  Size{320, 640},
		OnSizeChanged: func() {
			fmt.Println("OnSizeChanged", UI.dlg.Size())
			//UI.dlg.SetSize(walk.Size{
			//	Width:  320,
			//	Height: 440,
			//})
		},

		Icon: UI.icon,

		MenuItems: mw.menu(),
		Layout:    VBox{
			//Columns:     1,
			//MarginsZero: true,
			//Spacing:     11,
			//SpacingZero: true,
			//Alignment: AlignHNearVNear,
		},

		Children: []Widget{
			ImageView{
				AssignTo:  &mw.iv,
				Alignment: AlignHNearVFar,
			},

			Composite{
				Visible: false,
				Layout: VBox{
					MarginsZero: true,
				},

				Children: []Widget{
					GroupBox{
						Title:  "Installation",
						Layout: VBox{},
						Children: []Widget{
							HSpacer{ColumnSpan: 1},
							VSpacer{RowSpan: 1},
							Label{
								Text: "Installation status:",
							},
							TextEdit{
								Text: "This wizard will help with installation of missing components to run Mysterium Node.\r\n\r\n" +
									"Please press Install button to proceed with installation.",
								ReadOnly: true,
								MaxSize: Size{
									Height: 120,
								},
							},
							VSpacer{Row: 1},
							PushButton{
								AssignTo: &mw.btnBegin,
								Text:     "Install",
								OnClicked: func() {
									UI.BtnFinishOnClick()
								},
							},
						},
					},
				},
			},

			Composite{
				Visible: false,
				Layout: VBox{
					MarginsZero: true,
				},

				Children: []Widget{
					GroupBox{
						Title:  "Installation process",
						Layout: Grid{Columns: 2},
						Children: []Widget{
							VSpacer{RowSpan: 1, ColumnSpan: 2},
							Label{
								Text: "Check Windows version",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &mw.checkWindowsVersion,
							},

							Label{
								Text: "Check VT-x",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &mw.checkVTx,
							},
							Label{
								Text: "Check WSL",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &mw.enableWSL,
							},
							Label{
								Text: "Check Hyper-V",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &mw.enableHyperV,
							},

							Label{
								Text: "Install executable",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &mw.installExecutable,
							},
							Label{
								Text: "Reboot after WSL enable",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &mw.rebootAfterWSLEnable,
							},
							Label{
								Text: "Download files",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &mw.downloadFiles,
							},
							Label{
								Text: "Install WSL update",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &mw.installWSLUpdate,
							},
							Label{
								Text: "Install Docker",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &mw.installDocker,
							},
							Label{
								Text: "Check group membership (docker-users)",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &mw.checkGroupMembership,
							},

							VSpacer{
								ColumnSpan: 2,
								MinSize: Size{
									Height: 24,
								},
							},
							Label{
								Text:       "Installation status:",
								ColumnSpan: 2,
							},
							TextEdit{
								ColumnSpan: 2,
								RowSpan:    1,
								AssignTo:   &mw.lbInstallationStatus,
								ReadOnly:   true,
								MaxSize: Size{
									Height: 120,
								},
								VScroll: true,
							},

							VSpacer{ColumnSpan: 2, Row: 1},
							PushButton{
								ColumnSpan: 2,
								AssignTo:   &mw.btnFinish,
								Text:       "Finish",
								OnClicked: func() {
									UI.BtnFinishOnClick()
								},
							},
						},
					},
				},
			},

			Composite{
				Visible: false,
				Layout: VBox{
					MarginsZero: true,
				},
				Children: []Widget{
					GroupBox{
						Visible: true,
						Title:   "Status",
						Layout:  Grid{Columns: 2},
						Children: []Widget{
							VSpacer{ColumnSpan: 2},
							Label{
								Text: "Current node version",
							},
							Label{
								Text:     "-",
								AssignTo: &mw.lbVersionCurrent,
							},
							Label{
								Text: "Latest node version",
							},
							Label{
								Text:     "-",
								AssignTo: &mw.lbVersionLatest,
							},
							Label{
								Text: "Docker Hub image name",
							},
							Label{
								Text: model.State.ImageName,
							},
							Label{
								Text:       "-",
								ColumnSpan: 2,
							},

							Label{
								Text: "Docker",
							},
							Label{
								Text:     "-",
								AssignTo: &mw.lbDocker,
							},
							Label{
								Text: "Container",
							},
							Label{
								Text:     "-",
								AssignTo: &mw.lbContainer,
							},
							CheckBox{
								Text:     "Start launcher automatically",
								AssignTo: &mw.autoStart,
								OnCheckedChanged: func() {
									UI.Config.AutoStart = mw.autoStart.Checked()
									UI.SaveConfig()
								},
								ColumnSpan: 2,
							},
							PushButton{
								Enabled:  false,
								AssignTo: &mw.btnOpenNodeUI,
								Text:     "Open Node UI",
								OnClicked: func() {
									UI.OpenNodeUI()
								},
								ColumnSpan: 2,
							},
							VSpacer{ColumnSpan: 2},
						},
					},
				},
			},

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
			mw.iv.SetImage(i)
		}
	}

	if UI.InTray {
		UI.dlg.SetVisible(false)
	}

	// Events
	model.State.Bus.Subscribe("want-exit", func() {
		UI.dlg.Synchronize(func() {
			mw.btnFinish.SetEnabled(true)
		})
	})

	model.State.Bus.Subscribe("log", func(p []byte) {
		switch UI.state {
		case ModalStateInstallInProgress, ModalStateInstallError, ModalStateInstallFinished:
			UI.dlg.Synchronize(func() {
				mw.lbInstallationStatus.AppendText(string(p) + "\r\n")
			})
		}
	})
	model.State.Bus.Subscribe("show-dlg", func(d string, err error) {
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
		mw.actionEnable.SetEnabled(enable)
		mw.actionDisable.SetEnabled(enable)
		mw.actionUpgrade.SetEnabled(enable)
	}
	mw.currentView = frameState
	changeView := func(state modalState) {
		prev := mw.currentView
		mw.currentView = state
		if prev != state {
			UI.dlg.Children().At(int(prev)).SetVisible(false)
		}
		UI.dlg.Children().At(int(state)).SetVisible(true)
	}
	changeView(frameState)

	model.State.Bus.Subscribe("model-change", func() {
		UI.dlg.Synchronize(func() {
			switch UI.state {
			case ModalStateInitial:
				enableMenu(true)
				changeView(frameState)

				mw.autoStart.SetChecked(UI.Config.AutoStart)

				mw.lbDocker.SetText(model.State.StateDocker.String())
				mw.lbContainer.SetText(model.State.StateContainer.String())
				if !UI.Config.Enabled {
					mw.lbContainer.SetText("Disabled")
				}

				mw.btnOpenNodeUI.SetEnabled(model.State.IsRunning())
				mw.lbVersionLatest.SetText(UI.VersionLatest)
				mw.lbVersionCurrent.SetText(UI.VersionCurrent)

			case ModalStateInstallNeeded:
				enableMenu(false)
				changeView(frameInsNeed)
				mw.btnBegin.SetEnabled(true)

			case ModalStateInstallInProgress:
				enableMenu(false)
				changeView(frameIns)
				mw.btnFinish.SetEnabled(false)

			case ModalStateInstallFinished:
				enableMenu(false)
				changeView(frameIns)
				mw.btnFinish.SetEnabled(true)
				mw.btnFinish.SetText("Finish")

			case ModalStateInstallError:
				changeView(frameIns)
				mw.btnFinish.SetEnabled(true)
				mw.btnFinish.SetText("Exit installer")
			}

			switch UI.state {
			case ModalStateInstallInProgress, ModalStateInstallFinished, ModalStateInstallError:
				mw.checkWindowsVersion.SetChecked(UI.CheckWindowsVersion)
				mw.checkVTx.SetChecked(UI.CheckVTx)
				mw.enableWSL.SetChecked(UI.EnableWSL)
				mw.enableHyperV.SetChecked(UI.EnableHyperV)
				mw.installExecutable.SetChecked(UI.InstallExecutable)
				mw.rebootAfterWSLEnable.SetChecked(UI.RebootAfterWSLEnable)
				mw.downloadFiles.SetChecked(UI.DownloadFiles)
				mw.installWSLUpdate.SetChecked(UI.InstallWSLUpdate)
				mw.installDocker.SetChecked(UI.InstallDocker)
				mw.checkGroupMembership.SetChecked(UI.CheckGroupMembership)
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

func SetWS() {
	s := win.GetWindowLong(UI.mw.Handle(), win.GWL_STYLE) ^ (win.WS_THICKFRAME)
	fmt.Printf(">>> %X \r\n", s)
	ret := win.SetWindowLong(UI.mw.Handle(), win.GWL_STYLE, s)
	fmt.Printf(">>> %X \r\n", ret)
}
