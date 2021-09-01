/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package gui

import (
	"github.com/asaskevich/EventBus"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/model"
)

type UIModel struct {
	UIBus     EventBus.Bus
	waitClick chan int

	State    ModalState
	WantExit bool

	// common
	StateDocker    RunnableState
	StateContainer RunnableState

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

	App    model.AppInterface
	ImgVer model.ImageVersionInfo
	Config model.Config
}

func NewUIModel() *UIModel {
	m := &UIModel{}
	m.waitClick = make(chan int, 0)
	m.UIBus = EventBus.New()
	m.Config.Read()
	m.ImgVer.ImageName = _const.GetImageName()

	return m
}

func (m *UIModel) GetConfig() *model.Config {
	return &m.Config
}

func (m *UIModel) SetApp(app model.AppInterface) {
	m.App = app
}

func (m *UIModel) UpdateProperties(p UIProps) {
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

func (m *UIModel) SwitchState(s ModalState) {
	m.State = s
	m.Update()
}

func (m *UIModel) BtnFinishOnClick() {
	if m.WantExit {
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
	m.WantExit = true
	m.UIBus.Publish("want-exit")
}

func (m *UIModel) isExiting() bool {
	return m.State == ModalStateInstallError
}

func (m *UIModel) ExitApp() {
	close(m.waitClick)
	m.WantExit = true
	m.UIBus.Publish("exit")
}

func (m *UIModel) IsRunning() bool {
	return m.StateContainer == RunnableStateRunning
}

func (m *UIModel) SetStateDocker(r RunnableState) {
	m.StateDocker = r
	m.UIBus.Publish("model-change")
}

func (m *UIModel) SetStateContainer(r RunnableState) {
	m.StateContainer = r
	m.UIBus.Publish("model-change")
	m.UIBus.Publish("container-state")
}

func (m *UIModel) Publish(topic string, args ...interface{}) {
	m.UIBus.Publish(topic, args)
}

func (m *UIModel) TriggerAction(action string) {
	m.App.TriggerAction(action)
}
