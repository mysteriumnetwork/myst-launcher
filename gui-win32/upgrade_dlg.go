package gui_win32

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (g *Gui) OpenUpgradeDlg() {
	var (
		dialog             *walk.Dialog
		acceptPB, cancelPB *walk.PushButton
		lbVersionCurrent   *walk.Label
		lbVersionLatest    *walk.Label
		lbImageName        *walk.Label
	)

	refresh := func() {
		lbVersionCurrent.SetText(g.model.ImgVer.VersionCurrent)
		lbVersionLatest.SetText(g.model.ImgVer.VersionLatest)
		lbImageName.SetText(g.model.ImgVer.ImageName)
		acceptPB.SetEnabled(g.model.ImgVer.HasUpdate)
	}

	err := Dialog{
		AssignTo:      &dialog,
		Title:         "Would you like to upgrade?",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{400, 175},
		Icon:          g.icon,

		Layout: Grid{
			Columns: 2,
		},
		Children: []Widget{
			VSpacer{ColumnSpan: 2},
			Label{
				Text: "Docker Hub image name",
			},
			Label{
				Text:     "-",
				AssignTo: &lbImageName,
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
							g.model.App.TriggerAction("upgrade")
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
	}.Create(g.dlg)
	if err != nil {
		return
	}
	dialog.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		g.model.UIBus.Unsubscribe("model-change", refresh)
	})
	dialog.Disposing()

	dialog.Show()
	dialog.SetX(g.dlg.X() + 300)
	refresh()

	g.model.UIBus.Subscribe("model-change", refresh)
}
