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

func (g *Gui) OpenUpgradeNetworkDlg() {
	var (
		dialog             *walk.Dialog
		acceptPB, cancelPB *walk.PushButton
		lbCurrentNet       *walk.Label
	)

	refresh := func() {
		dialog.Synchronize(func() {
			lbCurrentNet.SetText(g.model.Config.GetNetworkCaption())
			cancelPB.SetFocus()
		})
	}

	err := Dialog{
		AssignTo:      &dialog,
		Title:         "Update to MainNet..",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{400, 220},
		Icon:          g.icon,

		Layout: Grid{
			Columns: 2,
		},
		Children: []Widget{
			VSpacer{ColumnSpan: 2},

			Label{
				Text: "Current network",
			},
			Label{
				Text:     "-",
				AssignTo: &lbCurrentNet,
			},
			Label{
				Text: "Update to",
			},
			Label{
				Text: "MainNet",
			},
			Label{
				Text: "",
			},
			LinkLabel{
				Text: "<a>Information about MainNet</a>",
				OnLinkActivated: func(link *walk.LinkLabelLink) {
					utils.OpenUrlInBrowser("https://mysterium.network/")
				},
				Alignment: AlignHNearVNear,
			},

			VSpacer{ColumnSpan: 2},
			Composite{
				ColumnSpan: 2,
				Layout:     HBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &acceptPB,
						Text:     "Update to MainNet",
						OnClicked: func() {
							dialog.Accept()
							g.model.UpdateToMainnet()
						},
					},
					PushButton{
						AssignTo: &cancelPB,
						Text:     "Cancel",
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

		refresh()
	 	g.model.UIBus.Subscribe("model-change", refresh)
	})
	dialog.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		g.model.UIBus.Unsubscribe("model-change", refresh)
	})
	dialog.Run()
}
