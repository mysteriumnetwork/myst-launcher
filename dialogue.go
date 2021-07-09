/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package main

import (
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func createDialogue() {
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
						AssignTo: &model.lbInstallationState,
					},
					Label{
						Text:     "-",
						AssignTo: &model.lbInstallationState2,
					},
					ProgressBar{
						AssignTo: &model.progressBar,
						Enabled:  false,
						Value:    50,
					},
					VSpacer{Row: 1},
					PushButton{
						AssignTo: &model.btnCmd,
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
						AssignTo: &model.lbDocker,
					},
					Label{
						Text: "Container",
					},
					Label{
						Text:     "-",
						AssignTo: &model.lbContainer,
					},
					VSpacer{
						ColumnSpan: 2,
						Size:       8,
					},
					PushButton{
						Enabled:  false,
						AssignTo: &model.btnOpenNodeUI,
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

	// prevent closing the app
	model.mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if model.isExiting() {
			walk.App().Exit(0)
		}
		*canceled = true
		model.mw.Hide()
	})
}
