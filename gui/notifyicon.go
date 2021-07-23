/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package gui

import (
	"log"

	"github.com/lxn/walk"
)

func CreateNotifyIcon() {
	var err error

	UI.mw, err = walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}

	UI.ni, err = walk.NewNotifyIcon(UI.mw)
	if err != nil {
		log.Fatal(err)
	}
	UI.Bus.Subscribe("exit", func() {
		UI.ni.Dispose()
	})

	if err := UI.ni.SetIcon(UI.icon); err != nil {
		log.Fatal(err)
	}
	if err := UI.ni.SetToolTip("Mysterium Network - Node Launcher"); err != nil {
		log.Fatal(err)
	}

	// When the left mouse button is pressed, bring up our balloon.
	UI.ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		UI.ShowMain()
	})
	UI.ni.MessageClicked().Attach(func() {
		UI.OpenNodeUI()
	})

	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		log.Fatal(err)
	}
	exitAction.Triggered().Attach(func() {
		UI.ExitApp()
	})

	openUIAction := walk.NewAction()
	if err := openUIAction.SetText("Open &UI"); err != nil {
		log.Fatal(err)
	}
	openUIAction.Triggered().Attach(func() {
		UI.OpenNodeUI()
	})

	if err := UI.ni.ContextMenu().Actions().Add(openUIAction); err != nil {
		log.Fatal(err)
	}
	if err := UI.ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}
	if err := UI.ni.SetVisible(true); err != nil {
		log.Fatal(err)
	}
}
