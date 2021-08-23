package gui

import (
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (g *Gui) UpgradeDlg(owner walk.Form) {
	var (
		dialog             *walk.Dialog
		acceptPB, cancelPB *walk.PushButton
		lbVersionCurrent   *walk.Label
		lbVersionLatest    *walk.Label
	)

	_, err := Dialog{
		AssignTo:      &dialog,
		Title:         "Would you like to upgrade?",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{400, 175},
		Icon:          UI.icon,

		Layout: Grid{
			Columns: 2,
		},
		Children: []Widget{
			VSpacer{ColumnSpan: 2},
			Label{
				Text: "Docker Hub image name",
			},
			Label{
				Text: UI.app.GetImageName(),
			},
			Label{
				Text: "Node version installed",
			},
			Label{
				Text:     "-",
				AssignTo: &lbVersionCurrent,
			},
			Label{
				Text: "Node version latest",
			},
			Label{
				Text:     "-",
				AssignTo: &lbVersionLatest,
			},
			VSpacer{ColumnSpan: 2},
			Composite{
				ColumnSpan: 2,
				Layout:     HBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &acceptPB,
						Text:     "Yes",
						OnClicked: func() {
							dialog.Accept()
							UI.app.TriggerAction("upgrade")
						},
					},
					PushButton{
						AssignTo: &cancelPB,
						Text:     "No",
						OnClicked: func() {
							dialog.Cancel()
						},
					},
				},
			},
		},
	}.Run(UI.dlg)
	if err != nil {
		fmt.Println(err)
	}
	refresh := func() {
		lbVersionCurrent.SetText(UI.VersionCurrent)
		lbVersionLatest.SetText(UI.VersionLatest)
		acceptPB.SetEnabled(UI.HasUpdate)
	}

	//dialog.Show()
	//dialog.SetX(UI.dlg.X() + 300)
	refresh()

	UI.app.Subscribe("model-change", refresh)
}
