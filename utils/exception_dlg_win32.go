//go:build windows
// +build windows

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package utils

import (
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/mysteriumnetwork/myst-launcher/native"
)

func OpenExceptionDlg(name, trace string) {
	var (
		dialog   *walk.Dialog
		acceptPB *walk.PushButton
	)

	dlg := Dialog{
		AssignTo:      &dialog,
		Title:         "Mysterium Launcher",
		DefaultButton: &acceptPB,
		MinSize:       Size{450, 200},

		Layout: Grid{
			Columns: 2,
		},
		Children: []Widget{

			Label{
				Text:       "Mysterium Launcher has stopped working.",
				ColumnSpan: 2,
			},
			VSpacer{
				ColumnSpan: 2,
				Size:       15,
			},
			Label{
				Text:       "A problem caused the program to stop working correctly.\r\nPlease send the report to Mysterium support: help@mysterium.network",
				ColumnSpan: 2,
			},

			VSpacer{ColumnSpan: 2},

			Composite{

				Layout: HBox{
					MarginsZero: true,
				},

				Children: []Widget{
					HSpacer{},

					PushButton{
						MinSize:   Size{Width: 120},
						Alignment: AlignHNearVCenter,
						Text:      "Copy the report",
						OnClicked: func() {
							walk.Clipboard().SetText(trace)
						},
					},

					PushButton{
						MinSize:   Size{Width: 60},
						Alignment: AlignHNearVCenter,
						AssignTo:  &acceptPB,
						Text:      "Exit",

						OnClicked: func() {
							dialog.Accept()
						},
					},
				},
			},
		},
	}

	dlg.Create(nil)
	dialog.Activating().Once(func() {
		dialog.SetX(400)
	})

	go func() {
		time.Sleep(100 * time.Millisecond)
		dialog.Accept()
		native.SwitchToThisWindow(dialog.Handle(), true)
	}()

	dialog.Run()

}
