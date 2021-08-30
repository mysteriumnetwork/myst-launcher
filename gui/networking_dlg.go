package gui

import (
	"strconv"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (g *Gui) NetworkingDlg(owner walk.Form) {
	var (
		dialog             *walk.Dialog
		acceptPB, cancelPB *walk.PushButton

		manualPortForwarding    *walk.CheckBox
		editRedirectionPortEnd  *walk.TextEdit
		editRedirectionPortFrom *walk.TextEdit

		canSave bool
	)
	loaded := false

	refreshState := func() {
		if !loaded {
			return
		}
		canSave = false

		editRedirectionPortFrom.SetEnabled(manualPortForwarding.Checked())
		editRedirectionPortEnd.SetEnabled(manualPortForwarding.Checked())

		if manualPortForwarding.Checked() != UI.app.GetConfig().EnablePortForwarding {
			canSave = true
		}
		if editRedirectionPortEnd.Text() != strconv.Itoa(UI.app.GetConfig().PortRangeEnd) {
			canSave = true
		}
		if editRedirectionPortFrom.Text() != strconv.Itoa(UI.app.GetConfig().PortRangeBegin) {
			canSave = true
		}
		acceptPB.SetEnabled(canSave)
	}

	validatePortRange := func(portRangeFrom, portRangeEnd int) bool {
		if portRangeEnd-portRangeFrom+1 < 100 {
			return false
		}
		if portRangeFrom > 65535 {
			return false
		}
		if portRangeEnd > 65535 {
			return false
		}
		return true
	}

	d := Dialog{
		Functions:     map[string]func(args ...interface{}) (interface{}, error){},
		AssignTo:      &dialog,
		Title:         "Networking settings (advanced settings)",
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
				Text:       "Redirection port range (UDP)",
				ColumnSpan: 2,
				MinSize:    Size{Height: 20},
			},
			Label{
				Text: "Range begin",
			},
			TextEdit{
				AssignTo:  &editRedirectionPortFrom,
				MaxLength: 5,
				MaxSize:   Size{Width: 50},
				OnTextChanged: func() {
					refreshState()
				},
			},
			Label{
				Text: "Range end",
			},
			TextEdit{
				AssignTo:  &editRedirectionPortEnd,
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
							portRangeBegin, err := strconv.Atoi(editRedirectionPortFrom.Text())
							if err != nil {
								return
							}
							portRangeLen, err := strconv.Atoi(editRedirectionPortEnd.Text())
							if err != nil {
								return
							}

							if !validatePortRange(portRangeBegin, portRangeLen) {
								walk.MsgBox(dialog,
									"Port range",
									"Wrong port range.\nPorts shall be in range of 1000..65535.\nNumber of ports in the range shall be at least 100.",
									walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
								return
							}
							UI.app.GetConfig().EnablePortForwarding = manualPortForwarding.Checked()
							UI.app.GetConfig().PortRangeBegin = portRangeBegin
							UI.app.GetConfig().PortRangeEnd = portRangeLen

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
	editRedirectionPortFrom.SetText(strconv.Itoa(UI.app.GetConfig().PortRangeBegin))
	editRedirectionPortEnd.SetText(strconv.Itoa(UI.app.GetConfig().PortRangeEnd))
	acceptPB.SetEnabled(canSave)
	loaded = true

	refreshState()
}
