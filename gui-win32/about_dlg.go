package gui_win32

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type AboutDlgData struct {
	Version string
}

func (g *Gui) OpenAboutDlg() {
	var (
		db       *walk.DataBinder
		dialog   *walk.Dialog
		acceptPB *walk.PushButton
	)

	data := &AboutDlgData{}
	data.Version = "Version: " + g.model.Config.ProductVersion

	dlg := Dialog{
		AssignTo:      &dialog,
		Title:         "About",
		DefaultButton: &acceptPB,
		MinSize:       Size{300, 175},
		Icon:          g.icon,

		DataBinder: DataBinder{
			AssignTo:   &db,
			DataSource: data,
		},
		Layout: Grid{
			Columns: 2,
		},
		Children: []Widget{
			VSpacer{ColumnSpan: 2},
			Label{
				Text: "Mysterium Launcher",
				ColumnSpan: 2,
				Alignment: AlignHCenterVCenter,
			},
			Label{
				ColumnSpan: 2,
				Text: Bind("Version"),
				Alignment: AlignHCenterVCenter,
			},

			VSpacer{ColumnSpan: 2},
			Composite{
				ColumnSpan: 2,
				Layout:     HBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							dialog.Accept()
						},
					},
				},
			},
		},
	}

	dlg.Create(g.dlg)
	dialog.Activating().Once(func() {
		dialog.SetX(400)
	})
	dialog.Run()
}
