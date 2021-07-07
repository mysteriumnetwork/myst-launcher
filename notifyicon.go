// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/lxn/walk"
)

func createNotifyIcon() {
	mw, err := walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}

	ni, err := walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}
	defer ni.Dispose()

	if err := ni.SetIcon(mod.icon); err != nil {
		log.Fatal(err)
	}
	if err := ni.SetToolTip("Mysterium Network - Node Launcher"); err != nil {
		log.Fatal(err)
	}

	// When the left mouse button is pressed, bring up our balloon.
	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		mod.mw.Show()
	})
	ni.MessageClicked().Attach(func() {
	})

	// We put an exit action into the context menu.
	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		log.Fatal(err)
	}
	exitAction.Triggered().Attach(func() {
		walk.App().Exit(0)
	})

	if err := ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}
	if err := ni.SetVisible(true); err != nil {
		log.Fatal(err)
	}

	// Now that the icon is visible, we can bring up an info balloon.
	//if err := ni.ShowInfo("Walk NotifyIcon Example", "Click the icon to show again."); err != nil {
	//	log.Fatal(err)
	//}

	// Run the message loop.
	mw.Run()
}
