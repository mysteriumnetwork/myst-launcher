package gui

import (
	"fmt"
	"strconv"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (g *Gui) NetworkingDlg(owner walk.Form) {
	var (
		dialog             *walk.Dialog
		acceptPB, cancelPB *walk.PushButton

		manualPortForwarding  *walk.CheckBox
		lbRedirectionPortSize *walk.TextEdit
		lbRedirectionPortFrom *walk.TextEdit

		//lbVersionLatest       *walk.Label
		canSave bool
	)
	loaded := false

	refreshState := func() {
		if !loaded {
			return
		}
		canSave = false
		fmt.Println("refresh >>", manualPortForwarding.Checked(), UI.app.GetConfig().EnablePortForwarding)

		lbRedirectionPortFrom.SetEnabled(manualPortForwarding.Checked())
		lbRedirectionPortSize.SetEnabled(manualPortForwarding.Checked())

		if manualPortForwarding.Checked() != UI.app.GetConfig().EnablePortForwarding {
			canSave = true
		}

		fmt.Println("refresh 2>>", lbRedirectionPortSize.Text(), strconv.Itoa(UI.app.GetConfig().PortRangeSize))
		if lbRedirectionPortSize.Text() != strconv.Itoa(UI.app.GetConfig().PortRangeSize) {
			canSave = true
		}

		fmt.Println("refresh 3>>", lbRedirectionPortFrom.Text(), strconv.Itoa(UI.app.GetConfig().PortRangeFrom))
		if lbRedirectionPortFrom.Text() != strconv.Itoa(UI.app.GetConfig().PortRangeFrom) {
			canSave = true
		}

		fmt.Println("refresh >>>", canSave)
		acceptPB.SetEnabled(canSave)
	}

	d := Dialog{
		AssignTo:      &dialog,
		Title:         "Networking settings",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{400, 175},
		Icon:          UI.icon,

		Layout: Grid{
			Columns: 2,
		},
		Children: []Widget{
			VSpacer{ColumnSpan: 2},
			CheckBox{
				Text:           "Use manual port forwarding",
				TextOnLeftSide: true,
				AssignTo:       &manualPortForwarding,
				OnCheckedChanged: func() {
					refreshState()
				},
				MaxSize: Size{Height: 15},
			},
			HSpacer{Size: 1},

			VSpacer{ColumnSpan: 2, Size: 10},
			Label{
				Text: "Redirection port range [from]",
			},
			TextEdit{
				AssignTo:  &lbRedirectionPortFrom,
				MaxLength: 5,
				MaxSize:   Size{Width: 50},
				OnTextChanged: func() {
					refreshState()
				},
			},
			Label{
				Text: "Redirection port range [size]",
			},
			TextEdit{
				AssignTo:  &lbRedirectionPortSize,
				MaxLength: 5,
				OnTextChanged: func() {
					refreshState()
				},
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
							fmt.Println(">>", manualPortForwarding.Checked())

							UI.app.GetConfig().EnablePortForwarding = manualPortForwarding.Checked()
							s, err := strconv.Atoi(lbRedirectionPortSize.Text())
							if err != nil {
								return
							}
							UI.app.GetConfig().PortRangeSize = s
							s, err = strconv.Atoi(lbRedirectionPortFrom.Text())
							if err != nil {
								return
							}
							UI.app.GetConfig().PortRangeFrom = s

							dialog.Accept()
							UI.app.TriggerAction("upgrade")
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
	}
	err := d.Create(owner)
	if err != nil {
		return
	}
	dialog.Show()
	dialog.SetX(UI.dlg.X() + 150)

	manualPortForwarding.SetChecked(UI.app.GetConfig().EnablePortForwarding)
	lbRedirectionPortSize.SetText(strconv.Itoa(UI.app.GetConfig().PortRangeSize))
	lbRedirectionPortFrom.SetText(strconv.Itoa(UI.app.GetConfig().PortRangeFrom))
	acceptPB.SetEnabled(canSave)
	loaded = true

	//refreshState()
	//UI.app.Subscribe("model-change", refreshState)
}
