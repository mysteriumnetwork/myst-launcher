/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package gui_win32

import (
	"log"

	"github.com/asaskevich/EventBus"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	model2 "github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/native"
)

const (
	frameInsNeed = 0
	frameIns     = 1
	frameState   = 2
)

type NotificationTypeID int

const (
	NotificationUpgrade          = NotificationTypeID(1)
	NotificationContainerStarted = NotificationTypeID(2)
)

type Gui struct {
	ni         *walk.NotifyIcon
	icon       *walk.Icon
	iconActive *walk.Icon
	mw         *walk.MainWindow
	dlg        *walk.MainWindow

	// menu
	actionFileMenu        *walk.Action
	actionMainMenu        *walk.Action
	actionOpenUI          *walk.Action
	actionUpgrade         *walk.Action
	actionLauncherUpgrade *walk.Action
	actionEnable          *walk.Action
	actionBackendNative   *walk.Action
	actionBackendDocker   *walk.Action

	isAutostartEnabled *walk.MutableCondition
	isNodeEnabled      *walk.MutableCondition

	//
	frameWidget *walk.Composite
	frame       Frame

	currentView model2.UIState
	logo        *walk.Icon
	icoActive   *walk.Icon

	model              *model2.UIModel
	lastNotificationID NotificationTypeID

	waitClick chan int
	bus       EventBus.Bus
}

func NewGui(m *model2.UIModel) *Gui {
	g := &Gui{}

	var err error
	g.icoActive, err = walk.NewIconFromResourceWithSize("ICON_ACTIVE", walk.Size{64, 64})
	if err != nil {
		log.Fatal(err)
	}

	g.icon, err = walk.NewIconFromResource("APPICON")
	if err != nil {
		log.Fatal(err)
	}
	g.iconActive, _ = walk.NewIconFromResource("ICON_ACTIVE")
	if err != nil {
		log.Fatal(err)
	}

	g.model = m
	g.isAutostartEnabled = walk.NewMutableCondition()
	g.isNodeEnabled = walk.NewMutableCondition()
	MustRegisterCondition("isAutostartEnabled", g.isAutostartEnabled)
	MustRegisterCondition("isNodeEnabled", g.isNodeEnabled)

	g.waitClick = make(chan int)
	g.bus = EventBus.New()
	return g
}

func (g *Gui) CreateMainWindow() {
	if err := (MainWindow{
		Visible:   false,
		AssignTo:  &g.dlg,
		Title:     "Mysterium Node Launcher",
		MinSize:   Size{380, 600},
		Size:      Size{380, 600},
		Icon:      g.icon,
		MenuItems: g.menu(),
		Layout:    VBox{},

		Children: []Widget{
			Composite{
				AssignTo: &g.frameWidget,
				Name:     "frameWidget",
				Layout:   HBox{MarginsZero: true, SpacingZero: true},
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}
	g.actionBackendNative.SetChecked(true)

	g.dlg.SetVisible(!g.model.App.GetInTray())
	g.changeView(frameState)

	// refresh on window restore
	g.dlg.Activating().Attach(func() {
		g.dlg.Synchronize(func() {
			g.refresh()
		})
	})
	g.model.UIBus.Subscribe("model-change", func() {
		g.dlg.Synchronize(func() {
			g.refresh()
		})
	})
	g.model.UIBus.Subscribe("state-change", func() {
		g.dlg.Synchronize(func() {
			g.refresh()
		})
	})
	g.model.UIBus.Subscribe("launcher-update", func() {
		g.dlg.Synchronize(func() {
			g.OpenUpgradeLauncherDlg()
		})
	})

	// prevent closing the app
	g.dlg.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if g.model.WantExit {
			walk.App().Exit(0)
			return
		}

		*canceled = true
		g.dlg.Hide()
	})

	g.bus.Subscribe("exit", func() {
		g.dlg.Synchronize(func() {
			g.dlg.Close()
			// g.mw.Close()
		})
	})

	// events
	g.model.UIBus.Subscribe("btn-config-click", func() {
		g.dlg.Synchronize(func() {
			g.NetworkingDlg()
		})
	})
	g.model.UIBus.Subscribe("btn-upgrade-network", func() {
		g.dlg.Synchronize(func() {
			g.OpenUpgradeNetworkDlg()
		})
	})
	g.model.UIBus.Subscribe("click-finish", func() {
		if g.model.WantExit {
			g.CloseUI()
		}
		g.DialogueComplete()
	})

}

func (g *Gui) enableMenu(enable bool) {
	//g.actionMainMenu.SetEnabled(enable)
	g.actionEnable.SetEnabled(enable)
	g.actionUpgrade.SetEnabled(enable)
}

func (g *Gui) changeView(state model2.UIState) {
	prev := g.currentView
	g.currentView = state
	if prev == state {
		return
	}

	if g.frame != nil {
		g.frame.Close()
		g.frame = nil
	}

	switch state {
	case frameInsNeed:
		g.frame = NewInstallationWelcomeFrame(g.frameWidget, g.model)
	case frameIns:
		g.frame = NewInstallationFrame(g.frameWidget, g.model)
	case frameState:
		g.frame = NewStateFrame(g.frameWidget, g.model)
	}
}

func (g *Gui) refresh() {
	switch g.model.State {

	case model2.UIStateInitial:
		g.enableMenu(true)
		g.changeView(frameState)

		g.isNodeEnabled.SetSatisfied(g.model.Config.Enabled)
		g.isAutostartEnabled.SetSatisfied(g.model.Config.AutoStart)
		g.actionLauncherUpgrade.SetVisible(g.model.LauncherHasUpdate)

		g.actionBackendNative.SetChecked(g.model.Config.Backend == "native")
		g.actionBackendDocker.SetChecked(g.model.Config.Backend == "docker")

	case model2.UIStateInstallNeeded:
		g.enableMenu(false)
		g.changeView(frameInsNeed)
		g.dlg.SetVisible(true)

	case model2.UIStateInstallInProgress:
		g.enableMenu(false)
		g.changeView(frameIns)
		g.dlg.SetVisible(true)

	case model2.UIStateInstallFinished:
		g.enableMenu(false)
		g.changeView(frameIns)

	case model2.UIStateInstallError:
		g.changeView(frameIns)
	}
}

func (g *Gui) bringMainToTop() {
	if g.dlg.Visible() {
		win.ShowWindow(g.dlg.Handle(), win.SW_SHOW)
		win.ShowWindow(g.dlg.Handle(), win.SW_SHOWNORMAL)

		native.SwitchToThisWindow(g.dlg.Handle(), true)

		win.SetWindowPos(g.dlg.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
		win.SetWindowPos(g.dlg.Handle(), win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
		win.SetWindowPos(g.dlg.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
	}
}

func (g *Gui) PopupMain() {
	if !g.dlg.Visible() {
		win.ShowWindow(g.dlg.Handle(), win.SW_RESTORE)
		return
	}
	g.bringMainToTop()
}

func (g *Gui) ShowMain() {
	if win.IsIconic(g.dlg.Handle()) {
		win.ShowWindow(g.dlg.Handle(), win.SW_RESTORE)
	} else {
		if g.dlg.Visible() {
			g.dlg.Hide()
			return
		} else {
			g.dlg.Show()
		}
	}
	g.bringMainToTop()
}

func (g *Gui) getModalOwner() walk.Form {
	if g != nil && g.dlg != nil {
		return g.dlg
	}
	return nil
}

func (g *Gui) ConfirmModal(title, message string) int {
	return walk.MsgBox(g.getModalOwner(), title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
}

func (g *Gui) YesNoModal(title, message string) int {
	return walk.MsgBox(g.getModalOwner(), title, message, walk.MsgBoxTopMost|walk.MsgBoxYesNo|walk.MsgBoxIconExclamation)
}

func (g *Gui) ErrorModal(title, message string) int {
	return walk.MsgBox(g.getModalOwner(), title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconError)
}

func (g *Gui) SetModalReturnCode(rc int) {}

func (g *Gui) Run() {
	g.mw.Run()
}

func (g *Gui) OpenDialogue(id int) {
	if id == 1 {
		g.dlg.Synchronize(func() {
			g.OpenUpgradeLauncherDlg()
		})
	}
}

// returns false, if dialogue was terminated
func (g *Gui) WaitDialogueComplete() bool {
	_, ok := <-g.waitClick
	return ok
}

func (g *Gui) TerminateWaitDialogueComplete() {
	close(g.waitClick)
}

func (g *Gui) DialogueComplete() {
	select {
	case g.waitClick <- 0:
	default:
	}
}

func (g *Gui) CloseUI() {
	g.model.WantExit = true
	g.bus.Publish("exit")
}
