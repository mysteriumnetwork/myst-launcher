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

func (g *Gui) OpenAgreementDlg(txt string) {
	var (
		dialog               *walk.Dialog
		acceptPB, cancelPB   *walk.PushButton
		lbInstallationStatus *walk.TextEdit
		lbNodeUI             *walk.LinkLabel
	)

	refresh := func() {
		dialog.Synchronize(func() {
			acceptPB.SetEnabled(true)
		})
	}

	err := Dialog{
		AssignTo:      &dialog,
		Title:         "Accept Terms and Conditions Agreement",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{500, 500},
		Icon:          g.icon,

		Layout: Grid{
			Columns: 2,
		},
		Children: []Widget{
			// VSpacer{ColumnSpan: 2},

			Label{
				Text: "Agreement:",
			},
			LinkLabel{
				Text:     "<a>Hyperlink to agreement</a>",
				AssignTo: &lbNodeUI,
				OnLinkActivated: func(link *walk.LinkLabelLink) {
					utils.OpenUrlInBrowser("https://github.com/mysteriumnetwork/node/tree/master/TERMS_EXIT_NODE.md")
				},
				Alignment: AlignHNearVNear,
			},
			TextEdit{
				ColumnSpan: 2,
				RowSpan:    2,
				AssignTo:   &lbInstallationStatus,
				ReadOnly:   true,
				VScroll:    true,
				Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},
			},

			VSpacer{ColumnSpan: 2},
			Composite{
				ColumnSpan: 2,
				Layout:     VBox{},

				Children: []Widget{
					PushButton{
						AssignTo: &acceptPB,
						Text:     "Accept Agreement",
						OnClicked: func() {
							g.model.Publish("accept-agreement")
							dialog.Accept()
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
		x := g.dlg.X()
		if x < 0 {
			x = 0
		}
		dialog.SetX(x + 300)

		refresh()
		lbInstallationStatus.SetText(txt)
	})
	dialog.Run()
}
