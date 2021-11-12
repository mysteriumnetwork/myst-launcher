/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package gui_win32

import (
	"log"
	"syscall"

	"github.com/asaskevich/EventBus"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	model2 "github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"github.com/mysteriumnetwork/myst-launcher/widget/impl"
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

	//
	actionFileMenu *walk.Action
	actionMainMenu *walk.Action
	actionOpenUI   *walk.Action
	actionUpgrade  *walk.Action
	actionEnable   *walk.Action

	isAutostartEnabled *walk.MutableCondition
	isNodeEnabled      *walk.MutableCondition

	// common
	stDocker         *impl.StatusViewImpl
	stContainer      *impl.StatusViewImpl
	lbDocker         *walk.Label
	lbContainer      *walk.Label
	lbVersionCurrent *walk.Label
	lbVersionLatest  *walk.Label
	lbImageName      *walk.Label

	autoUpgrade   *walk.CheckBox
	lbNetworkMode *walk.Label
	btnOpenNodeUI *walk.PushButton

	lbNetwork  *walk.Label
	btnMainNet *walk.PushButton

	// install
	lbInstallationStatus *walk.TextEdit
	btnBegin             *walk.PushButton

	checkWindowsVersion  *impl.StatusViewImpl
	checkVirt            *impl.StatusViewImpl
	installExecutable    *impl.StatusViewImpl
	rebootAfterWSLEnable *impl.StatusViewImpl
	downloadFiles        *impl.StatusViewImpl
	installWSLUpdate     *impl.StatusViewImpl
	installDocker        *impl.StatusViewImpl
	checkGroupMembership *impl.StatusViewImpl

	btnFinish *walk.PushButton
	img       *walk.ImageView

	currentView model2.UIState
	logo        *walk.Icon
	icoActive   *walk.Icon

	model              *model2.UIModel
	lastNotificationID NotificationTypeID

	waitClick chan int
	bus       EventBus.Bus

	cmp              *walk.Composite
	headerContainer  *walk.Composite
	lbNodeUI         *walk.LinkLabel
	lbMMN            *walk.LinkLabel
	lbUpdateLauncher *walk.LinkLabel
}

func NewGui(m *model2.UIModel) *Gui {
	g := &Gui{}

	var err error
	g.logo, err = walk.NewIconFromResourceWithSize("APPICON", walk.Size{64, 64})
	if err != nil {
		log.Fatal(err)
	}
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
			g.installationWelcome(),
			g.installationDlg(),
			g.stateDlg(),
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	// manual layout for head panel
	g.cmp.SetBounds(walk.Rectangle{0, 0, 400, 100})
	g.img.SetBounds(walk.Rectangle{5, 5, 70, 70})
	g.headerContainer.SetBounds(walk.Rectangle{80, 5, 270, 70})
	g.stContainer.SetBounds(walk.Rectangle{0, 0, 20, 20})
	g.lbContainer.SetBounds(walk.Rectangle{25, 0, 70, 20})
	g.lbNodeUI.SetBounds(walk.Rectangle{120, 4, 150, 20})
	g.lbMMN.SetBounds(walk.Rectangle{120, 24, 150, 20})
	g.lbUpdateLauncher.SetBounds(walk.Rectangle{120, 44, 150, 20})

	g.dlg.SetVisible(!g.model.App.GetInTray())
	g.setImage()

	// Events
	g.model.UIBus.Subscribe("want-exit", func() {
		g.dlg.Synchronize(func() {
			g.btnFinish.SetEnabled(true)
		})
	})

	g.model.UIBus.Subscribe("log", func(p []byte) {
		switch g.model.State {
		case model2.UIStateInstallInProgress, model2.UIStateInstallError, model2.UIStateInstallFinished:
			g.dlg.Synchronize(func() {
				g.lbInstallationStatus.AppendText(string(p) + "\r\n")
			})
		}
	})
	g.currentView = frameState
	g.changeView(frameState)

	// refresh on window restore
	g.dlg.Activating().Attach(func() {
		g.dlg.Synchronize(func() {
			g.refresh()
			g.setImage()
		})
	})
	g.model.UIBus.Subscribe("model-change", func() {
		g.dlg.Synchronize(func() {
			g.refresh()
			//g.setImage()
		})
	})
	g.model.UIBus.Subscribe("state-change", func() {
		g.dlg.Synchronize(func() {
			g.refresh()
		})
	})
	g.model.UIBus.Subscribe("launcher-update", func() {
		g.OpenDialogue(1)
	})

	// prevent closing the app
	g.dlg.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if g.model.WantExit {
			walk.App().Exit(0)
		}
		*canceled = true
		g.dlg.Hide()
	})

	g.bus.Subscribe("exit", func() {
		g.dlg.Synchronize(func() {
			g.dlg.Close()
		})
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

	if prev != state {
		g.dlg.Children().At(int(prev)).SetVisible(false)
	}
	g.dlg.Children().At(int(state)).SetVisible(true)
	g.dlg.Children().At(int(state)).SetAlwaysConsumeSpace(true)
	g.dlg.Children().At(int(state)).SetAlwaysConsumeSpace(false)
}

func (g *Gui) refresh() {
	if !g.dlg.Visible() {
		return
	}
	switch g.model.State {

	case model2.UIStateInitial:
		g.enableMenu(true)
		g.changeView(frameState)

		g.isNodeEnabled.SetSatisfied(g.model.Config.Enabled)
		g.isAutostartEnabled.SetSatisfied(g.model.Config.AutoStart)

		g.autoUpgrade.SetChecked(g.model.GetConfig().AutoUpgrade)
		if !g.model.GetConfig().EnablePortForwarding {
			g.lbNetworkMode.SetText(`Port restricted cone NAT`)
		} else {
			g.lbNetworkMode.SetText(`Manual port forwarding`)
		}

		g.lbDocker.SetText(g.model.StateDocker.String())
		g.lbContainer.SetText(g.model.StateContainer.String())
		setState2(g.stDocker, g.model.StateDocker)
		setState2(g.stContainer, g.model.StateContainer)

		g.lbContainer.SetText(g.model.StateContainer.String())
		if !g.model.GetConfig().Enabled {
			g.lbContainer.SetText("Disabled")
		}
		g.btnOpenNodeUI.SetEnabled(g.model.IsRunning())

		g.lbVersionCurrent.SetText(g.model.ImgVer.VersionCurrent)
		g.lbVersionLatest.SetText(g.model.ImgVer.VersionLatest)

		g.lbImageName.SetText(g.model.Config.GetFullImageName())
		g.btnOpenNodeUI.SetFocus()

		g.lbUpdateLauncher.SetVisible(g.model.LauncherHasUpdate)

		g.lbNetwork.SetText(g.model.Config.GetNetworkCaption())
		g.btnMainNet.SetVisible(g.model.Config.Network != "mainnet")

	case model2.UIStateInstallNeeded:
		g.enableMenu(false)
		g.changeView(frameInsNeed)
		g.btnBegin.SetEnabled(true)

	case model2.UIStateInstallInProgress:
		g.enableMenu(false)
		g.changeView(frameIns)
		g.btnFinish.SetEnabled(false)

	case model2.UIStateInstallFinished:
		g.enableMenu(false)
		g.changeView(frameIns)
		g.btnFinish.SetEnabled(true)
		g.btnFinish.SetText("Finish")

	case model2.UIStateInstallError:
		g.changeView(frameIns)
		g.btnFinish.SetEnabled(true)
		g.btnFinish.SetText("Exit installer")
	}

	switch g.model.State {
	case model2.UIStateInstallInProgress, model2.UIStateInstallFinished, model2.UIStateInstallError:
		setState(g.checkWindowsVersion, g.model.CheckWindowsVersion)
		setState(g.checkVirt, g.model.CheckVirt)
		setState(g.installExecutable, g.model.InstallExecutable)
		setState(g.rebootAfterWSLEnable, g.model.RebootAfterWSLEnable)
		setState(g.downloadFiles, g.model.DownloadFiles)
		setState(g.installWSLUpdate, g.model.InstallWSLUpdate)
		setState(g.installDocker, g.model.InstallDocker)
		setState(g.checkGroupMembership, g.model.CheckGroupMembership)
	}
}

func setState(b *impl.StatusViewImpl, st model2.InstallStep) {
	switch st {
	case model2.StepInProgress:
		b.SetState(3)
	case model2.StepFinished:
		b.SetState(2)
	case model2.StepFailed:
		b.SetState(1)
	case model2.StepNone:
		b.SetState(0)
	default:
		b.SetState(0)
	}
}
func setState2(b *impl.StatusViewImpl, st model2.RunnableState) {
	switch st {
	case model2.RunnableStateRunning:
		b.SetState(2)
	case model2.RunnableStateInstalling:
		b.SetState(3)
	case model2.RunnableStateStarting:
		b.SetState(3)
	case model2.RunnableStateUnknown:
		b.SetState(1)
	}
}

func (g *Gui) setImage() {
	if !g.dlg.Visible() {
		return
	}

	img, err := walk.ImageFrom(g.logo)
	if err != nil {
		return
	}
	g.img.SetImage(img)
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

func openUrlInBrowser(url string) {
	native.ShellExecuteAndWait(
		0,
		"",
		"rundll32",
		"url.dll,FileProtocolHandler "+url,
		"",
		syscall.SW_NORMAL)
}

func OpenNodeUI() {
	openUrlInBrowser("http://localhost:4449/")
}

func OpenMMN() {
	openUrlInBrowser("https://mystnodes.com/")
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
