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
	frameW = 1
	frameI = 2
	frameS = 3
)

func CreateDialogue() {
	var (
		// common
		lbDocker      *walk.Label
		lbContainer   *walk.Label
		autoStart     *walk.CheckBox
		btnOpenNodeUI *walk.PushButton

		// install
		lbInstallationStatus *walk.TextEdit
		btnBegin             *walk.PushButton

		checkWindowsVersion  *walk.CheckBox
		checkVTx             *walk.CheckBox
		enableWSL            *walk.CheckBox
		installExecutable    *walk.CheckBox
		rebootAfterWSLEnable *walk.CheckBox
		downloadFiles        *walk.CheckBox
		installWSLUpdate     *walk.CheckBox
		installDocker        *walk.CheckBox
		checkGroupMembership *walk.CheckBox

		btnFinish *walk.PushButton

		iv  *walk.ImageView
		iv2 *walk.ImageView
		iv3 *walk.ImageView
	)
	SModel.ReadConfig()

	if err := (MainWindow{
		AssignTo: &SModel.mw,
		Title:    "Mysterium Exit Node Launcher",
		MinSize:  Size{320, 240},
		Size:     Size{400, 600},
		Icon:     SModel.Icon,

		Layout: VBox{},
		Children: []Widget{
			VSpacer{RowSpan: 1},

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
							ImageView{
								AssignTo:  &iv3,
								Alignment: AlignHNearVFar,
							},
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
								AssignTo: &btnBegin,
								Text:     "Install",
								OnClicked: func() {
									SModel.BtnOnClick()
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
							ImageView{
								AssignTo:   &iv2,
								Alignment:  AlignHNearVFar,
								ColumnSpan: 2,
							},
							VSpacer{RowSpan: 1, ColumnSpan: 2},

							Label{
								Text: "Check Windows version",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &checkWindowsVersion,
							},

							Label{
								Text: "Check VT-x",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &checkVTx,
							},
							Label{
								Text: "Enable WSL",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &enableWSL,
							},
							Label{
								Text: "Install executable",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &installExecutable,
							},
							Label{
								Text: "Reboot after WSL enable",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &rebootAfterWSLEnable,
							},
							Label{
								Text: "Download files",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &downloadFiles,
							},
							Label{
								Text: "Install WSL update",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &installWSLUpdate,
							},
							Label{
								Text: "Install Docker",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &installDocker,
							},
							Label{
								Text: "Check group membership (docker-users)",
							},
							CheckBox{
								Enabled:  false,
								AssignTo: &checkGroupMembership,
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
								AssignTo:   &lbInstallationStatus,
								ReadOnly:   true,
								MaxSize: Size{
									Height: 120,
								},
								VScroll: true,
							},

							VSpacer{ColumnSpan: 2, Row: 1},
							PushButton{
								ColumnSpan: 2,
								AssignTo:   &btnFinish,
								Text:       "Finish",
								OnClicked: func() {
									SModel.BtnOnClick()
								},
							},
						},
					},
				},
			},

			GroupBox{
				Visible: false,
				Title:   "Status",
				Layout:  Grid{Columns: 2},
				Children: []Widget{
					ImageView{
						AssignTo: &iv,
					},
					VSpacer{ColumnSpan: 2},
					Label{
						Text: "Docker",
					},
					Label{
						Text:     "-",
						AssignTo: &lbDocker,
					},
					Label{
						Text: "Container",
					},
					Label{
						Text:     "-",
						AssignTo: &lbContainer,
					},
					CheckBox{
						Text:     "Start Node automatically",
						AssignTo: &autoStart,
						OnCheckedChanged: func() {
							SModel.CFG.AutoStart = autoStart.Checked()
							SModel.SaveConfig()
						},
					},
					PushButton{
						Enabled:  false,
						AssignTo: &btnOpenNodeUI,
						Text:     "Open Node UI",
						OnClicked: func() {
							SModel.OpenNodeUI()
						},
						ColumnSpan: 2,
					},
					VSpacer{ColumnSpan: 2},
				},
			},
			VSpacer{RowSpan: 1},
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
			iv.SetImage(i)
			iv2.SetImage(i)
			iv3.SetImage(i)
		}
	}

	if SModel.InTray {
		SModel.mw.SetVisible(false)
	}

	SModel.Bus.Subscribe("log", func(p []byte) {
		switch SModel.state {
		case ModalStateInstallInProgress, ModalStateInstallError, ModalStateInstallFinished:
			SModel.mw.Synchronize(func() {
				lbInstallationStatus.AppendText(string(p))
				lbInstallationStatus.AppendText("\r\n")
			})
		}
	})
	SModel.Bus.Subscribe("state-change", func() {
		SModel.mw.Synchronize(func() {
			switch SModel.state {
			case ModalStateInitial:
				SModel.mw.Children().At(frameW).SetVisible(false)
				SModel.mw.Children().At(frameI).SetVisible(false)
				SModel.mw.Children().At(frameS).SetVisible(true)
				autoStart.SetChecked(SModel.CFG.AutoStart)

				switch SModel.StateDocker {
				case RunnableStateRunning:
					lbDocker.SetText("Running [OK]")
				case RunnableStateInstalling:
					lbDocker.SetText("Installing..")
				case RunnableStateStarting:
					lbDocker.SetText("Starting..")
				case RunnableStateUnknown:
					lbDocker.SetText("-")
				}
				switch SModel.StateContainer {
				case RunnableStateRunning:
					lbContainer.SetText("Running [OK]")
				case RunnableStateInstalling:
					lbContainer.SetText("Installing..")
				case RunnableStateStarting:
					lbContainer.SetText("Starting..")
				case RunnableStateUnknown:
					lbContainer.SetText("-")
				}
				btnOpenNodeUI.SetEnabled(SModel.StateContainer == RunnableStateRunning)

			case ModalStateInstallNeeded:
				SModel.mw.Children().At(frameW).SetVisible(true)
				SModel.mw.Children().At(frameI).SetVisible(false)
				SModel.mw.Children().At(frameS).SetVisible(false)
				btnBegin.SetEnabled(true)

			case ModalStateInstallInProgress:
				SModel.mw.Children().At(frameW).SetVisible(false)
				SModel.mw.Children().At(frameI).SetVisible(true)
				SModel.mw.Children().At(frameS).SetVisible(false)

				checkWindowsVersion.SetChecked(SModel.CheckWindowsVersion)
				checkVTx.SetChecked(SModel.CheckVTx)
				enableWSL.SetChecked(SModel.EnableWSL)
				installExecutable.SetChecked(SModel.InstallExecutable)
				rebootAfterWSLEnable.SetChecked(SModel.RebootAfterWSLEnable)
				downloadFiles.SetChecked(SModel.DownloadFiles)
				installWSLUpdate.SetChecked(SModel.InstallWSLUpdate)
				installDocker.SetChecked(SModel.InstallDocker)
				checkGroupMembership.SetChecked(SModel.CheckGroupMembership)
				btnFinish.SetEnabled(false)

			case ModalStateInstallFinished:
				btnFinish.SetEnabled(true)
				btnFinish.SetText("Finish")

			case ModalStateInstallError:
				SModel.mw.Children().At(frameI).SetVisible(true)
				SModel.mw.Children().At(frameS).SetVisible(false)
				btnFinish.SetEnabled(true)
				btnFinish.SetText("Exit installer")
			}
		})
	})

	// prevent closing the app
	SModel.mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if SModel.isExiting() {
			walk.App().Exit(0)
		}
		*canceled = true
		SModel.mw.Hide()
	})
}
