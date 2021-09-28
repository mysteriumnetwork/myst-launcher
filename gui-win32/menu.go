package gui_win32

import (
	. "github.com/lxn/walk/declarative"
)

func (g *Gui) menu() []MenuItem {
	return []MenuItem{
		Menu{
			AssignActionTo: &g.actionFileMenu,
			Text:           "&File",
			Items: []MenuItem{
				Action{
					Text: "E&xit",
					OnTriggered: func() {
						g.TerminateWaitDialogueComplete()
						g.CloseUI()
					},
				},
			},
		},

		Menu{
			AssignActionTo: &g.actionMainMenu,
			Text:           "&Node",
			Items: []MenuItem{
				Action{
					Text:        "&Open Node UI",
					AssignTo:    &g.actionOpenUI,
					OnTriggered: func() { OpenNodeUI() },
				},
				Separator{},
				Action{
					Text:     "Check updates",
					AssignTo: &g.actionUpgrade,
					OnTriggered: func() {
						g.OpenUpgradeDlg()
					},
				},

				Separator{},
				Action{
					Text:     "Disable node",
					AssignTo: &g.actionDisable,
					OnTriggered: func() {
						//g.model.BtnDisableOnClick()
						g.model.TriggerAction("disable")
					},
				},
				Action{
					Text:     "Enable node",
					AssignTo: &g.actionEnable,
					OnTriggered: func() {
						//g.model.BtnEnableOnClick()
						g.model.TriggerAction("enable")
					},
				},
			},
		},
	}
}
