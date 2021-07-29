package gui

import (
	. "github.com/lxn/walk/declarative"
)

func (mw *Gui) menu() []MenuItem {
	return []MenuItem{
		Menu{
			AssignActionTo: &mw.actionFileMenu,
			Text:           "&File",
			Items: []MenuItem{
				Action{
					Text:        "E&xit",
					OnTriggered: func() { UI.ExitApp() },
				},
			},
		},

		Menu{
			AssignActionTo: &mw.actionMainMenu,
			Text:           "&Node",
			Items: []MenuItem{
				Action{
					Text:        "&Open Node UI",
					AssignTo:    &mw.actionUpgrade,
					OnTriggered: func() { UI.OpenNodeUI() },
				},
				Separator{},
				Action{
					Text:        "Check updates",
					AssignTo:    &mw.actionUpgrade,
					OnTriggered: func() { UI.BtnUpgradeOnClick() },
				},

				Separator{},
				Action{
					Text:        "Disable node",
					AssignTo:    &mw.actionDisable,
					OnTriggered: func() { UI.BtnDisableOnClick() },
				},
				Action{
					Text:        "Enable node",
					AssignTo:    &mw.actionEnable,
					OnTriggered: func() { UI.BtnEnableOnClick() },
				},
			},
		},
	}
}
