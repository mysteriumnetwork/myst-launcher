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

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	gui2 "github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/native"
)

const (
	ofs          = 0
	frameImage_  = 0 + ofs
	frameInsNeed = 1 + ofs
	frameIns     = 2 + ofs
	frameState   = 3 + ofs
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
	actionDisable  *walk.Action

	// common
	lbDocker              *walk.Label
	lbContainer           *walk.Label
	lbVersionLatest       *walk.Label
	lbVersionCurrent      *walk.Label
	lbVersionUpdatesAvail *walk.LinkLabel
	lbImageName           *walk.Label

	autoUpgrade          *walk.CheckBox
	manualPortForwarding *walk.CheckBox
	lbNetworkMode        *walk.LinkLabel
	btnOpenNodeUI        *walk.PushButton

	// install
	lbInstallationStatus *walk.TextEdit
	btnBegin             *walk.PushButton

	checkWindowsVersion  *walk.CheckBox
	checkVTx             *walk.CheckBox
	enableWSL            *walk.CheckBox
	enableHyperV         *walk.CheckBox
	installExecutable    *walk.CheckBox
	rebootAfterWSLEnable *walk.CheckBox
	downloadFiles        *walk.CheckBox
	installWSLUpdate     *walk.CheckBox
	installDocker        *walk.CheckBox
	checkGroupMembership *walk.CheckBox

	btnFinish *walk.PushButton
	img       *walk.ImageView

	currentView gui2.ModalState
	ico         *walk.Icon
	icoActive   *walk.Icon

	model              *gui2.UIModel
	LastNotificationID NotificationTypeID
}

func NewGui(m *gui2.UIModel) *Gui {
	g := &Gui{}
	g.icon, _ = walk.NewIconFromResourceId(2)
	g.iconActive, _ = walk.NewIconFromResourceId(3)
	g.model = m

	return g
}

func (g *Gui) CreateMainWindow() {
	if err := (MainWindow{
		Visible:   false,
		AssignTo:  &g.dlg,
		Title:     "Mysterium Node Launcher",
		MinSize:   Size{380, 640},
		Size:      Size{380, 640},
		Icon:      g.icon,
		MenuItems: g.menu(),
		Layout:    VBox{},

		Children: []Widget{
			ImageView{
				AssignTo:  &g.img,
				Alignment: AlignHNearVFar,
			},
			g.installationWelcome(),
			g.installationDlg(),
			g.stateDlg(),
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}
	g.dlg.SetVisible(!g.model.App.GetInTray())

	var err error
	g.ico, err = walk.NewIconFromResourceIdWithSize(2, walk.Size{
		Width:  64,
		Height: 64,
	})
	if err != nil {
		log.Fatal(err)
	}
	g.icoActive, err = walk.NewIconFromResourceIdWithSize(3, walk.Size{
		Width:  64,
		Height: 64,
	})
	if err != nil {
		log.Fatal(err)
	}
	g.setImage()

	g.model.UIBus.Subscribe("container-state", func() {
		g.dlg.Synchronize(func() {
			g.setImage()
		})
	})
	// Events
	g.model.UIBus.Subscribe("want-exit", func() {
		g.dlg.Synchronize(func() {
			g.btnFinish.SetEnabled(true)
		})
	})

	g.model.UIBus.Subscribe("log", func(p []byte) {
		switch g.model.State {
		case gui2.ModalStateInstallInProgress, gui2.ModalStateInstallError, gui2.ModalStateInstallFinished:
			g.dlg.Synchronize(func() {
				g.lbInstallationStatus.AppendText(string(p) + "\r\n")
			})
		}
	})
	g.currentView = frameState
	g.changeView(frameState)

	g.dlg.Activating().Attach(func() {
		if g.dlg.Visible() {
			// refresh on window restore
			g.dlg.Synchronize(func() {
				g.refresh()
			})
		}
	})
	g.model.UIBus.Subscribe("model-change", func() {
		if !g.dlg.Visible() {
			return
		}
		g.dlg.Synchronize(func() {
			g.refresh()
		})
	})

	// prevent closing the app
	g.dlg.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {

		if g.model.WantExit {
			walk.App().Exit(0)
		}
		*canceled = true
		g.dlg.Hide()
	})

	g.model.UIBus.Subscribe("exit", func() {
		g.dlg.Synchronize(func() {
			g.dlg.Close()
		})
	})
}

func (g *Gui) enableMenu(enable bool) {
	//actionMainMenu.SetEnabled(enable)
	g.actionEnable.SetEnabled(enable)
	g.actionDisable.SetEnabled(enable)
	g.actionUpgrade.SetEnabled(enable)
}

func (g *Gui) changeView(state gui2.ModalState) {
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
	switch g.model.State {

	case gui2.ModalStateInitial:
		g.enableMenu(true)
		g.changeView(frameState)

		g.autoUpgrade.SetChecked(g.model.GetConfig().AutoUpgrade)
		if !g.model.GetConfig().EnablePortForwarding {
			g.lbNetworkMode.SetText(`<a id="net">Port restricted cone NAT</a>`)
		} else {
			g.lbNetworkMode.SetText(`<a id="net">Manual port forwarding</a>`)
		}

		g.lbDocker.SetText(g.model.StateDocker.String())
		g.lbContainer.SetText(g.model.StateContainer.String())
		if !g.model.GetConfig().Enabled {
			g.lbContainer.SetText("Disabled")
		}
		g.btnOpenNodeUI.SetEnabled(g.model.IsRunning())
		//g.lbVersionLatest.SetText(g.model.VersionLatest)

		g.lbVersionCurrent.SetText(g.model.ImgVer.VersionCurrent)
		g.lbVersionUpdatesAvail.SetText("-")
		if g.model.ImgVer.HasUpdate {
			g.lbVersionUpdatesAvail.SetText(`<a id="upgrade">Yes !</a> - click to see details`)
		}
		g.lbImageName.SetText(g.model.ImgVer.ImageName)
		g.btnOpenNodeUI.SetFocus()

	case gui2.ModalStateInstallNeeded:
		g.enableMenu(false)
		g.changeView(frameInsNeed)
		g.btnBegin.SetEnabled(true)

	case gui2.ModalStateInstallInProgress:
		g.enableMenu(false)
		g.changeView(frameIns)
		g.btnFinish.SetEnabled(false)

	case gui2.ModalStateInstallFinished:
		g.enableMenu(false)
		g.changeView(frameIns)
		g.btnFinish.SetEnabled(true)
		g.btnFinish.SetText("Finish")

	case gui2.ModalStateInstallError:
		g.changeView(frameIns)
		g.btnFinish.SetEnabled(true)
		g.btnFinish.SetText("Exit installer")
	}

	switch g.model.State {
	case gui2.ModalStateInstallInProgress, gui2.ModalStateInstallFinished, gui2.ModalStateInstallError:
		g.checkWindowsVersion.SetChecked(g.model.CheckWindowsVersion)
		g.checkVTx.SetChecked(g.model.CheckVTx)
		g.enableWSL.SetChecked(g.model.EnableWSL)
		g.enableHyperV.SetChecked(g.model.EnableHyperV)
		g.installExecutable.SetChecked(g.model.InstallExecutable)
		g.rebootAfterWSLEnable.SetChecked(g.model.RebootAfterWSLEnable)
		g.downloadFiles.SetChecked(g.model.DownloadFiles)
		g.installWSLUpdate.SetChecked(g.model.InstallWSLUpdate)
		g.installDocker.SetChecked(g.model.InstallDocker)
		g.checkGroupMembership.SetChecked(g.model.CheckGroupMembership)
	}
}

func (g *Gui) setImage() {
	ico := g.ico
	if g.model.StateContainer == gui2.RunnableStateRunning {
		ico = g.icoActive
	}
	img, err := walk.ImageFrom(ico)
	if err != nil {
		return
	}
	g.img.SetImage(img)
}

func (g *Gui) ShowMain() {
	if !g.dlg.Visible() {
		win.ShowWindow(g.dlg.Handle(), win.SW_SHOW)
		win.ShowWindow(g.dlg.Handle(), win.SW_SHOWNORMAL)

		native.SwitchToThisWindow(g.dlg.Handle(), false)

		win.SetWindowPos(g.dlg.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
		win.SetWindowPos(g.dlg.Handle(), win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
		win.SetWindowPos(g.dlg.Handle(), win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE)
		return
	}

	if !win.IsIconic(g.dlg.Handle()) {
		win.ShowWindow(g.dlg.Handle(), win.SW_MINIMIZE)
	} else {
		win.ShowWindow(g.dlg.Handle(), win.SW_RESTORE)
	}
}

func OpenNodeUI() {
	native.ShellExecuteAndWait(
		0,
		"",
		"rundll32",
		"url.dll,FileProtocolHandler http://localhost:4449/",
		"",
		syscall.SW_NORMAL)
}

func (g *Gui) ShowNotificationInstalled() {
	//g.LastNotificationID = NotificationContainerStarted
	err := g.ni.ShowCustom(
		"Mysterium Node successfully installed!",
		"Click this notification to open Node UI in browser",
		g.icon)

	if err != nil {
	}
}

func (g *Gui) ShowNotificationUpgrade() {
	//g.LastNotificationID = NotificationUpgrade
	err := g.ni.ShowCustom(
		"Upgrade available",
		"Click this notification to see details.",
		g.icon)

	if err != nil {
	}
}

func (g *Gui) ConfirmModal(title, message string) int {
	res := make(chan int)
	go func() {
		res <- walk.MsgBox(g.dlg, title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconExclamation)
	}()
	return <-res
}

func (g *Gui) YesNoModal(title, message string) int {
	res := make(chan int)
	go func() {
		res <- walk.MsgBox(g.dlg, title, message, walk.MsgBoxTopMost|walk.MsgBoxYesNo|walk.MsgBoxIconExclamation)
	}()
	return <-res
}

func (g *Gui) ErrorModal(title, message string) int {
	res := make(chan int)
	go func() {
		res <- walk.MsgBox(g.dlg, title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconError)
	}()
	return <-res
}

func (g *Gui) Run() {
	g.mw.Run()
}
