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
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

func (g *Gui) OpenUpgradeLauncherDlg() {
	var (
		dialog             *walk.Dialog
		acceptPB, cancelPB *walk.PushButton
		lbVersionCurrent   *walk.Label
		lbVersionLatest    *walk.Label
	)

	refresh := func() {
		dialog.Synchronize(func() {
			lbVersionCurrent.SetText(g.model.ProductVersion)
			lbVersionLatest.SetText(g.model.ProductVersionLatest)
			acceptPB.SetEnabled(g.model.LauncherHasUpdate)
		})
	}

	err := Dialog{
		AssignTo:      &dialog,
		Title:         "New version of launcher available",
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
				ColumnSpan: 2,
				Text:       "Mysterium Network launcher",
			},

			Label{
				Text: "Installed Version",
			},
			Label{
				Text:     "-",
				AssignTo: &lbVersionCurrent,
			},
			Label{
				Text: "Latest Version",
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
						AssignTo:    &acceptPB,
						Text:        "Download latest version",
						ToolTipText: "Open link in browser",
						OnClicked: func() {
							dialog.Accept()
							utils.OpenUrlInBrowser(g.model.ProductVersionLatestUrl)
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
		refresh()
		g.model.UIBus.Subscribe("model-change", refresh)
	})
	dialog.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		g.model.UIBus.Unsubscribe("model-change", refresh)
	})
	dialog.Run()
}
