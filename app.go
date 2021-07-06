// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate goversioninfo -icon=ico/icon_512x512.ico -manifest=logview.exe.manifest
package main

import (
	"log"
	"os/exec"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	icon, _ := walk.NewIconFromResourceId(2)
	if err := (MainWindow{
		AssignTo: &mod.mw,
		Title:    "Mysterium Exit Node launcher",
		MinSize:  Size{320, 240},
		Size:     Size{400, 600},
		Icon:     icon,

		Layout: VBox{
			//MarginsZero: true,
		},
		Children: []Widget{
			//CustomWidget{},
			//CustomWidget{},
			VSpacer{RowSpan: 1},
			//Composite{},
			//Composite{

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

					VSpacer{
						Row: 1,
						//RowSpan: 1,
					},
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
					PushButton{
						Enabled:  false,
						AssignTo: &mod.btnCmd2,
						Text:     "- Open UI -",
						OnClicked: func() {
							//rundll32 url.dll,FileProtocolHandler https://www.google.com
							cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:4449")
							if err := cmd.Start(); err != nil {
							}
						},
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
	//lv.PostAppendText("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\r\n")
	log.SetOutput(mod.lv)
	//mod.mw.SetIcon(icon)

	go func() {
		checkSystemsAndTry()
	}()
	mod.mw.Run()
}
