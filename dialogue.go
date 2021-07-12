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
	frameI = 1
	frameS = 2
)

func createDialogue() {
	var (
		// common
		lbDocker      *walk.Label
		lbContainer   *walk.Label
		autoStart     *walk.CheckBox
		btnOpenNodeUI *walk.PushButton

		// install
		lbInstallationState  *walk.Label
		lbInstallationState2 *walk.Label
		progressBar          *walk.ProgressBar
		btnCmd               *walk.PushButton
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
				Title:   "Installation",
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
						Text:     "- installation info -",
						AssignTo: &lbInstallationState,
					},
					Label{
						Text:     "-",
						AssignTo: &lbInstallationState2,
					},
					ProgressBar{
						AssignTo: &progressBar,
						Enabled:  false,
						Value:    50,
					},
					VSpacer{Row: 1},
					PushButton{
						AssignTo: &btnCmd,
						Text:     "-",
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
					VSpacer{ColumnSpan: 2},
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
	autoStart.SetChecked(model.cfg.AutoStart)

	go func() {
		for {
			select {
			case sig := <-model.signal:
				fmt.Println("received signal", sig)

				switch model.state {
				case initial:
					model.mw.Children().At(frameI).SetVisible(false)
					model.mw.Children().At(frameS).SetVisible(true)
				case installNeeded:
					model.mw.Children().At(frameI).SetVisible(true)
					model.mw.Children().At(frameS).SetVisible(false)
					//model.HideProgress()
					progressBar.SetVisible(false)

					btnCmd.SetEnabled(true)
					btnCmd.SetText("Install")
					btnCmd.SetFocus()
					lbInstallationState.SetText("Docker desktop is required to run exit node.")
					lbInstallationState2.SetText("Press button to begin installation.")

				case installInProgress:
					//btnCmd.SetEnabled(false)
					lbInstallationState.SetText("Downloading installation packages.")
					lbInstallationState2.SetText("-")
					progressBar.SetVisible(model.progressVisible)
					progressBar.SetValue(model.progress)

				case installFinished:
					lbInstallationState.SetText("Installation successfully finished!")
					btnCmd.SetEnabled(true)
					btnCmd.SetText("Finish !")
				case installError:
					model.mw.Children().At(frameI).SetVisible(true)
					model.mw.Children().At(frameS).SetVisible(false)
					//model.HideProgress()
					progressBar.SetVisible(false)

					lbInstallationState.SetText("Installation failed")
					btnCmd.SetEnabled(true)
					btnCmd.SetText("Exit installer")
				}

				model.mw.Synchronize(func() {
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
