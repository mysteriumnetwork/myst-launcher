/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package main

import (
	"log"
	"os/exec"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func createDialogue() {
	if err := (MainWindow{
		AssignTo: &mod.mw,
		Title:    "Mysterium Exit Node Launcher",
		MinSize:  Size{320, 240},
		Size:     Size{400, 600},
		Icon:     mod.icon,

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
						AssignTo: &mod.lbInstallationState,
					},
					Label{
						Text:     "-",
						AssignTo: &mod.lbInstallationState2,
					},
					ProgressBar{
						AssignTo: &mod.progressBar,
						Enabled:  false,
						Value:    50,
					},
					VSpacer{Row: 1},
					PushButton{
						AssignTo: &mod.btnCmd,
						Text:     "-",
						OnClicked: func() {
							mod.BtnOnClick()
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
						AssignTo: &mod.lbDocker,
					},
					Label{
						Text: "Container",
					},
					Label{
						Text:     "-",
						AssignTo: &mod.lbContainer,
					},
					VSpacer{
						ColumnSpan: 2,
						Size:       8,
					},
					PushButton{
						Enabled:  false,
						AssignTo: &mod.btnCmd2,
						Text:     "Open Node UI",
						OnClicked: func() {
							cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:4449")
							if err := cmd.Start(); err != nil {
							}
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

	var err error
	mod.lv, err = NewLogView(mod.mw)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(mod.lv)

	// prevent closing the app
	mod.mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if mod.isExiting() {
			// works only from main thread
			walk.App().Exit(0)
		}
		*canceled = true
		mod.mw.Hide()
	})
}
