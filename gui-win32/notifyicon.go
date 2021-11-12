/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package gui_win32

import (
	"log"

	"github.com/mysteriumnetwork/myst-launcher/model"

	"github.com/lxn/walk"
)

func (g *Gui) CreateNotifyIcon(ui *model.UIModel) {
	var err error

	g.mw, err = walk.NewMainWindow()
	if err != nil {
		log.Fatal("NewMainWindow", err)
	}
	g.ni, err = walk.NewNotifyIcon(g.mw)
	if err != nil {
		log.Println("NewNotifyIcon", err)
		return
	}

	g.bus.Subscribe("exit", func() {
		g.ni.Dispose()
	})
	ui.UIBus.Subscribe("container-state", func() {
		i := g.icon
		if ui.StateContainer == model.RunnableStateRunning {
			i = g.iconActive
		}

		g.ni.SetIcon(i)
	})

	if err := g.ni.SetIcon(g.icon); err != nil {
		log.Fatal(err)
	}
	if err := g.ni.SetToolTip("Mysterium Node Launcher"); err != nil {
		log.Fatal(err)
	}

	// When the left mouse button is pressed, bring up our balloon.
	g.ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		g.ShowMain()
	})

	g.ni.MessageClicked().Attach(func() {
		switch g.lastNotificationID {
		case NotificationUpgrade:
			g.OpenUpgradeDlg()

		case NotificationContainerStarted:
			OpenNodeUI()
		}
	})

	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		log.Fatal(err)
	}
	exitAction.Triggered().Attach(func() {
		g.TerminateWaitDialogueComplete()
		g.CloseUI()
	})

	openUIAction := walk.NewAction()
	openUIAction.SetText("Open &UI")
	openUIAction.Triggered().Attach(func() {
		OpenNodeUI()
	})

	nodeEnabled := walk.NewAction()
	nodeEnabled.SetText("Enable node")
	nodeEnabled.Triggered().Attach(func() {
		g.model.TriggerNodeEnableAction()
	})
	nodeEnabled.SetCheckedCondition(g.isNodeEnabled)

	g.ni.ContextMenu().Actions().Add(openUIAction)
	g.ni.ContextMenu().Actions().Add(nodeEnabled)
	g.ni.ContextMenu().Actions().Add(walk.NewSeparatorAction())
	g.ni.ContextMenu().Actions().Add(exitAction)

	g.ni.SetVisible(true)
}

func (g *Gui) ShowNotificationInstalled() {
	g.lastNotificationID = NotificationContainerStarted
	if g.ni == nil {
		return
	}

	g.ni.ShowCustom(
		"Mysterium Node successfully installed!",
		"Click this notification to open Node UI in browser",
		g.icon)
}

func (g *Gui) ShowNotificationUpgrade() {
	g.lastNotificationID = NotificationUpgrade
	if g.ni == nil {
		return
	}

	g.ni.ShowCustom(
		"Upgrade available",
		"Click this notification to see details.",
		g.icon)
}
