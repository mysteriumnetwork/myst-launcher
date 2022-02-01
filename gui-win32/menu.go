/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

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
					Text: "About",
					OnTriggered: func() {
						g.OpenAboutDlg()
					},
				},
				Action{
					Text:     "Launcher update available",
					AssignTo: &g.actionLauncherUpgrade,
					OnTriggered: func() {
						g.OpenUpgradeLauncherDlg()
					},
				},

				Separator{},
				Action{
					Checked: Bind("isAutostartEnabled"),
					Text:    "Autostart launcher",
					OnTriggered: func() {
						g.model.TriggerAutostartAction()
					},
				},
				Separator{},

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
				Action{
					Text:     "Check updates",
					AssignTo: &g.actionUpgrade,
					OnTriggered: func() {
						g.OpenUpgradeDlg()
					},
				},

				Separator{},
				Action{
					Checked:  Bind("isNodeEnabled"),
					Text:     "Enable node",
					AssignTo: &g.actionEnable,
					OnTriggered: func() {
						g.model.TriggerNodeEnableAction()
					},
				},
			},
		},
	}
}
