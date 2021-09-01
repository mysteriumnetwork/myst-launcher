package gui_win32

import (
	"strconv"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (g *Gui) NetworkingDlg() {
	var (
		dialog                  *walk.Dialog
		acceptPB, cancelPB      *walk.PushButton
		manualPortForwarding    *walk.CheckBox
		editRedirectionPortEnd  *walk.TextEdit
		editRedirectionPortFrom *walk.TextEdit
		canSave                 bool
	)
	loaded := false

	refreshState := func() {
		if !loaded {
			return
		}
		canSave = false

		editRedirectionPortFrom.SetEnabled(manualPortForwarding.Checked())
		editRedirectionPortEnd.SetEnabled(manualPortForwarding.Checked())

		conf := g.model.GetConfig()
		if manualPortForwarding.Checked() != conf.EnablePortForwarding {
			canSave = true
		}
		if editRedirectionPortEnd.Text() != strconv.Itoa(conf.PortRangeEnd) {
			canSave = true
		}
		if editRedirectionPortFrom.Text() != strconv.Itoa(conf.PortRangeBegin) {
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
		Icon:          g.icon,

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
							g.model.GetConfig().EnablePortForwarding = manualPortForwarding.Checked()
							g.model.GetConfig().PortRangeBegin = portRangeBegin
							g.model.GetConfig().PortRangeEnd = portRangeLen

							dialog.Accept()
							g.model.App.TriggerAction("upgrade")
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
	err := d.Create(g.dlg)
	if err != nil {
		return
	}
	dialog.Show()
	dialog.SetX(g.dlg.X() + 150)

	manualPortForwarding.SetChecked(g.model.GetConfig().EnablePortForwarding)
	editRedirectionPortFrom.SetText(strconv.Itoa(g.model.GetConfig().PortRangeBegin))
	editRedirectionPortEnd.SetText(strconv.Itoa(g.model.GetConfig().PortRangeEnd))
	acceptPB.SetEnabled(canSave)
	loaded = true

	refreshState()
}
