/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

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
		dialog.Synchronize(func() {
			lbVersionCurrent.SetText(g.model.ImageInfo.VersionCurrent)
			lbVersionLatest.SetText(g.model.ImageInfo.VersionLatest)
			lbImageName.SetText(g.model.Config.GetFullImageName())
			acceptPB.SetEnabled(g.model.ImageInfo.HasUpdate)
		})
	}

	err := Dialog{
		AssignTo:      &dialog,
		Title:         "Would you like to upgrade node?",
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
	dialog.Activating().Once(func() {
		dialog.SetX(g.dlg.X() + 300)
		g.model.App.TriggerAction("check")

		refresh()
		g.model.UIBus.Subscribe("model-change", refresh)
	})
	dialog.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		g.model.UIBus.Unsubscribe("model-change", refresh)
	})
	dialog.Run()
}
