/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package main

import (
	"fmt"
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

const (
	frameW = 1
	frameI = 2
	frameS = 3
)

func createDialogue() {
	var (
		// common
		lbDocker      *walk.Label
		lbContainer   *walk.Label
		autoStart     *walk.CheckBox
		btnOpenNodeUI *walk.PushButton

		// install
		lbInstallationStatus *walk.Label
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
	)

	if err := (MainWindow{
		AssignTo: &model.mw,
		Title:    "Mysterium Exit Node Launcher",
		MinSize:  Size{320, 240},
		Size:     Size{400, 600},
		Icon:     model.icon,

		Layout: VBox{
			//MarginsZero: true,
		},
		Children: []Widget{
			VSpacer{RowSpan: 1},

			GroupBox{
				Visible: false,
				Title:   "Installation needed",
				Layout:  VBox{},
				Children: []Widget{
					VSpacer{RowSpan: 1},
					Label{
						Text: "Status:",
						Font: Font{
							Family:    "Arial",
							PointSize: 9,
							Bold:      true,
						},
					},
					Label{
						Text: "This wizard will install missing components to run Mysterium Node.\r\nPress install button to proceed with installation",
					},
					VSpacer{Row: 1},
					PushButton{
						AssignTo: &btnBegin,
						Text:     "Install",
						OnClicked: func() {
							model.BtnOnClick()
						},
					},
				},
			},

			GroupBox{
				Visible: false,
				Title:   "Installation process",
				//Layout:  VBox{},

				Layout: Grid{Columns: 2},
				Children: []Widget{
					VSpacer{ColumnSpan: 2},
					Label{
						Text: "check Windows version",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &checkWindowsVersion,
					},

					Label{
						Text: "check VT-x",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &checkVTx,
					},
					Label{
						Text: "enable WSL",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &enableWSL,
					},
					Label{
						Text: "install executable",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &installExecutable,
					},
					Label{
						Text: "reboot after WSL enable",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &rebootAfterWSLEnable,
					},
					Label{
						Text: "download files",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &downloadFiles,
					},
					Label{
						Text: "install WSL update",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &installWSLUpdate,
					},
					Label{
						Text: "install Docker",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &installDocker,
					},
					Label{
						Text: "checkGroupMembership",
					},
					CheckBox{
						Enabled:  false,
						AssignTo: &checkGroupMembership,
					},

					VSpacer{
						ColumnSpan: 2,
						Size:       8,
					},
					Label{
						ColumnSpan: 2,
						AssignTo:   &lbInstallationStatus,
					},

					VSpacer{ColumnSpan: 2},
					PushButton{
						AssignTo:   &btnFinish,
						Text:       "Finish",
						ColumnSpan: 2,
						OnClicked: func() {
							model.BtnOnClick()
						},
					},
				},
			},

			GroupBox{
				Visible: false,
				Title:   "Status",
				Layout:  Grid{Columns: 2},
				Children: []Widget{
					//VSpacer{ColumnSpan: 2},
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
							model.cfg.AutoStart = autoStart.Checked()
							model.saveConfig()
						},
					},

					VSpacer{
						ColumnSpan: 2,
						Size:       8,
					},
					PushButton{
						Enabled:  false,
						AssignTo: &btnOpenNodeUI,
						Text:     "Open Node UI",
						OnClicked: func() {
							model.openNodeUI()
						},
						ColumnSpan: 2,
					},
					//VSpacer{ColumnSpan: 2},
				},
			},
			VSpacer{RowSpan: 1},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}
	if model.inTray {
		model.mw.SetVisible(false)
	}
	model.readConfig()

	go func() {
		for {
			select {
			case sig := <-model.signal:
				fmt.Println("received signal", sig)

				model.mw.Synchronize(func() {
					switch model.state {
					case initial:
						model.mw.Children().At(frameW).SetVisible(false)
						model.mw.Children().At(frameI).SetVisible(false)
						model.mw.Children().At(frameS).SetVisible(true)
						autoStart.SetChecked(model.cfg.AutoStart)

					case installNeeded:
						model.mw.Children().At(frameW).SetVisible(true)
						model.mw.Children().At(frameI).SetVisible(false)
						model.mw.Children().At(frameS).SetVisible(false)
						//model.HideProgress()
						//progressBar.SetVisible(false)

						btnBegin.SetEnabled(true)
						//btnCmd.SetText("Install")
						//btnCmd.SetFocus()
						//lbInstallationState.SetText("Docker desktop is required to run exit node.")
						//lbInstallationStatus.SetText("Press button to begin installation.")

					case installInProgress:
						model.mw.Children().At(frameW).SetVisible(false)
						model.mw.Children().At(frameI).SetVisible(true)
						model.mw.Children().At(frameS).SetVisible(false)

						checkWindowsVersion.SetChecked(model.checkWindowsVersion)
						checkVTx.SetChecked(model.checkVTx)
						enableWSL.SetChecked(model.enableWSL)
						installExecutable.SetChecked(model.installExecutable)
						rebootAfterWSLEnable.SetChecked(model.rebootAfterWSLEnable)
						downloadFiles.SetChecked(model.downloadFiles)
						installWSLUpdate.SetChecked(model.installWSLUpdate)
						installDocker.SetChecked(model.installDocker)
						checkGroupMembership.SetChecked(model.checkGroupMembership)

						lbInstallationStatus.SetText(model.installationStatus)
						btnFinish.SetEnabled(false)

					case installFinished:
						lbInstallationStatus.SetText(model.installationStatus)
						btnFinish.SetEnabled(true)
						btnFinish.SetText("Finish")

					case installError:
						model.mw.Children().At(frameI).SetVisible(true)
						model.mw.Children().At(frameS).SetVisible(false)
						lbInstallationStatus.SetText(model.installationStatus)

						btnFinish.SetEnabled(true)
						btnFinish.SetText("Exit installer")
					}

					switch model.stateDocker {
					case stateRunning:
						lbDocker.SetText("Running [OK]")
					case stateInstalling:
						lbDocker.SetText("Installing..")
					case stateStarting:
						lbDocker.SetText("Starting..")
					case stateUnknown:
						lbDocker.SetText("-")
					}
					switch model.stateContainer {
					case stateRunning:
						lbContainer.SetText("Running [OK]")
					case stateInstalling:
						lbContainer.SetText("Installing..")
					case stateStarting:
						lbContainer.SetText("Starting..")
					case stateUnknown:
						lbContainer.SetText("-")
					}
					btnOpenNodeUI.SetEnabled(model.stateContainer == stateRunning)
				})
			}
		}
	}()

	// prevent closing the app
	model.mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if model.isExiting() {
			walk.App().Exit(0)
		}
		*canceled = true
		model.mw.Hide()
	})
}
