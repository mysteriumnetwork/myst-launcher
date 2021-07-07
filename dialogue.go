package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func createDialogue() {
	if err := (MainWindow{
		AssignTo: &mod.mw,
		Title:    "Mysterium Exit Node launcher",
		MinSize:  Size{320, 240},
		Size:     Size{400, 600},
		Icon:     mod.icon,

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
					ProgressBar{
						AssignTo: &mod.progressBar,
						Visible:  false,
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
					PushButton{
						Enabled:  false,
						AssignTo: &mod.btnCmd2,
						Text:     "Install",
						OnClicked: func() {
							fmt.Println(os.Args[0])
							runMeElevated(os.Args[0], flagInstall, "")
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
		if reason == walk.CloseReasonUser {
		}
		*canceled = true
		mod.mw.Hide()
	})
}
