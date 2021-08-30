/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package gui_win32

import (
	"github.com/asaskevich/EventBus"

	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/myst"
)

type UIModel struct {
	UIBus     EventBus.Bus
	waitClick chan int

	state    gui.ModalState
	wantExit bool

	// common
	StateDocker    gui.RunnableState
	StateContainer gui.RunnableState

	// inst
	CheckWindowsVersion  bool
	CheckVTx             bool
	EnableWSL            bool
	EnableHyperV         bool
	InstallExecutable    bool
	RebootAfterWSLEnable bool
	DownloadFiles        bool
	InstallWSLUpdate     bool
	InstallDocker        bool
	CheckGroupMembership bool

	app    model.AppInterface
	imgVer *myst.ImageVersionInfo
}

func NewUIModel() *UIModel {
	m := &UIModel{}
	m.waitClick = make(chan int, 0)
	m.UIBus = EventBus.New()
	return m
}

func (m *UIModel) SetImageVersionInfo(ivi *myst.ImageVersionInfo) {
	m.imgVer = ivi
}

func (m *UIModel) SetApp(app model.AppInterface) {
	m.app = app
}

func (m *UIModel) UpdateProperties(p gui.UIProps) {
	for k, v := range p {
		switch k {
		case "CheckWindowsVersion":
			m.CheckWindowsVersion = v.(bool)
		case "CheckVTx":
			m.CheckVTx = v.(bool)
		case "EnableWSL":
			m.EnableWSL = v.(bool)
		case "EnableHyperV":
			m.EnableHyperV = v.(bool)
		case "InstallExecutable":
			m.InstallExecutable = v.(bool)
		case "RebootAfterWSLEnable":
			m.RebootAfterWSLEnable = v.(bool)
		case "DownloadFiles":
			m.DownloadFiles = v.(bool)
		case "InstallWSLUpdate":
			m.InstallWSLUpdate = v.(bool)
		case "InstallDocker":
			m.InstallDocker = v.(bool)
		case "CheckGroupMembership":
			m.CheckGroupMembership = v.(bool)
		}
	}
	m.UIBus.Publish("model-change")
}

func (m *UIModel) Update() {
	m.UIBus.Publish("model-change")
}

func (m *UIModel) SwitchState(s gui.ModalState) {
	m.state = s
	m.Update()
}

func (m *UIModel) BtnFinishOnClick() {
	if m.wantExit {
		m.ExitApp()
		return
	}

	select {
	case m.waitClick <- 0:
	default:
	}
}

// returns channel close status
func (m *UIModel) WaitDialogueComplete() bool {
	_, ok := <-m.waitClick
	return ok
}

func (m *UIModel) SetWantExit() {
	m.wantExit = true
	m.UIBus.Publish("want-exit")
}

func (m *UIModel) isExiting() bool {
	return m.state == gui.ModalStateInstallError
}

func (m *UIModel) ExitApp() {
	close(m.waitClick)
	m.wantExit = true
	m.UIBus.Publish("exit")
}

func (m *UIModel) IsRunning() bool {
	return m.StateContainer == gui.RunnableStateRunning
}

func (m *UIModel) SetStateDocker(r gui.RunnableState) {
	m.StateDocker = r
	m.UIBus.Publish("model-change")
}

func (m *UIModel) SetStateContainer(r gui.RunnableState) {
	m.StateContainer = r
	m.UIBus.Publish("model-change")
	m.UIBus.Publish("container-state")
}

func (m *UIModel) GetImageName() string {
	return m.imgVer.ImageName
}

func (m *UIModel) Publish(topic string, args ...interface{}) {
	m.UIBus.Publish(topic, args)
}

func (m *UIModel) TriggerAction(action string) {
	m.app.TriggerAction(action)
}
