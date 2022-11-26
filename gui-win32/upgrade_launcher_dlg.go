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

func (g *Gui) OpenUpgradeLauncherDlg() {
	var (
		dialog             *walk.Dialog
		acceptPB, cancelPB *walk.PushButton
		lbVersionCurrent   *walk.Label
		lbVersionLatest    *walk.Label
		pBarDownload       *walk.ProgressBar
	)

	refresh := func() {
		dialog.Synchronize(func() {
			lbVersionCurrent.SetText(g.model.ProductVersion)
			lbVersionLatest.SetText(g.model.ProductVersionLatest)
			acceptPB.SetEnabled(g.model.LauncherHasUpdate)
		})
	}
	setProgress := func(i int) {
		dialog.Synchronize(func() {
			pBarDownload.SetValue(i)
			if i == 100 || i == -1 {
				dialog.Accept()
			}
		})
	}

	err := Dialog{
		AssignTo:      &dialog,
		Title:         "Update of Mysterium Launcher",
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
				Text:       "Mysterium Launcher",
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
				Layout:     VBox{},

				Children: []Widget{
					ProgressBar{
						MaxValue: 100,
						MinValue: 0,
						AssignTo: &pBarDownload,
					},

					PushButton{
						AssignTo:    &acceptPB,
						Text:        "Download and install update",
						// ToolTipText: "",
						OnClicked: func() {
							g.model.Publish("launcher-update-ok", int(2))
							// dialog.Accept()
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
		g.model.UIBus.Subscribe("launcher-update-download", setProgress)

	})
	dialog.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		g.model.UIBus.Unsubscribe("model-change", refresh)
		g.model.UIBus.Unsubscribe("launcher-update-download", setProgress)
		g.model.Publish("launcher-update-ok", int(0))
	})
	dialog.Run()
}
