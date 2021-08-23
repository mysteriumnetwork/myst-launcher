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
		editRedirectionPortSize *walk.TextEdit
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
		editRedirectionPortSize.SetEnabled(manualPortForwarding.Checked())

		if manualPortForwarding.Checked() != UI.app.GetConfig().EnablePortForwarding {
			canSave = true
		}
		if editRedirectionPortSize.Text() != strconv.Itoa(UI.app.GetConfig().PortRangeSize) {
			canSave = true
		}
		if editRedirectionPortFrom.Text() != strconv.Itoa(UI.app.GetConfig().PortRangeFrom) {
			canSave = true
		}
		acceptPB.SetEnabled(canSave)
	}

	validatePortRange := func(portRangeFrom, len int) bool {
		if portRangeFrom < 1000 || len <= 1 {
			return false
		}
		if portRangeFrom > 65535 {
			return false
		}
		if portRangeFrom+len > 65535 {
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
				Text:       "Redirection port range",
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
				Text: "Number of ports",
			},
			TextEdit{
				AssignTo:  &editRedirectionPortSize,
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
							portRangeFrom, err := strconv.Atoi(editRedirectionPortFrom.Text())
							if err != nil {
								return
							}
							portRangeLen, err := strconv.Atoi(editRedirectionPortSize.Text())
							if err != nil {
								return
							}

							if !validatePortRange(portRangeFrom, portRangeLen) {
								walk.MsgBox(dialog, "Port range", "Wrong port range.\nPorts should be in range of 1000..65535", walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
								return
							}
							UI.app.GetConfig().EnablePortForwarding = manualPortForwarding.Checked()
							UI.app.GetConfig().PortRangeFrom = portRangeFrom
							UI.app.GetConfig().PortRangeSize = portRangeLen

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
	editRedirectionPortSize.SetText(strconv.Itoa(UI.app.GetConfig().PortRangeSize))
	editRedirectionPortFrom.SetText(strconv.Itoa(UI.app.GetConfig().PortRangeFrom))
	acceptPB.SetEnabled(canSave)
	loaded = true

	refreshState()
}
