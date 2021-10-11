/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package model

import (
	"log"

	"github.com/asaskevich/EventBus"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type UIModel struct {
	UIBus EventBus.Bus

	State    UIState
	WantExit bool

	// common
	StateDocker    RunnableState
	StateContainer RunnableState

	// inst
	CheckWindowsVersion bool
	CheckVirt           bool
	CheckDocker         bool // darwin

	InstallExecutable    bool
	RebootAfterWSLEnable bool
	DownloadFiles        bool
	InstallWSLUpdate     bool
	InstallDocker        bool
	CheckGroupMembership bool

	App    AppInterface
	ImgVer ImageVersionInfo
	Config Config
}

func NewUIModel() *UIModel {
	m := &UIModel{}
	m.UIBus = EventBus.New()
	m.Config.Read()
	m.ImgVer.ImageName = _const.GetImageName()

	return m
}

func (m *UIModel) GetConfig() *Config {
	return &m.Config
}

func (m *UIModel) SetApp(app AppInterface) {
	m.App = app
}

func (m *UIModel) UpdateProperties(p UIProps) {
	for k, v := range p {
		switch k {
		case "CheckWindowsVersion":
			m.CheckWindowsVersion = v.(bool)
		case "CheckVTx":
			m.CheckVirt = v.(bool)
		case "CheckDocker":
			m.CheckDocker = v.(bool)
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
		default:
			log.Println("Unknown proprerty:", k)
		}
	}
	m.UIBus.Publish("model-change")
}

func (m *UIModel) Update() {
	m.UIBus.Publish("model-change")
}

func (m *UIModel) SwitchState(s UIState) {
	m.State = s
	m.UIBus.Publish("state-change")
}

func (m *UIModel) SetWantExit() {
	m.WantExit = true
	m.UIBus.Publish("want-exit")
}

func (m *UIModel) IsRunning() bool {
	return m.StateContainer == RunnableStateRunning
}

func (m *UIModel) OnConfigRead() {
	m.UIBus.Publish("config-read")
}

func (m *UIModel) SetStateDocker(r RunnableState) {
	if m.StateDocker != r {
		m.StateDocker = r
		m.UIBus.Publish("model-change")
	}
}

func (m *UIModel) SetStateContainer(r RunnableState) {
	if m.StateContainer != r {
		m.StateContainer = r
		m.UIBus.Publish("model-change")
	}
}

func (m *UIModel) Publish(topic string, args ...interface{}) {
	m.UIBus.Publish(topic, args...)
}

func (m *UIModel) TriggerAction(action string) {
	m.App.TriggerAction(action)
}

func (m *UIModel) TriggerAutostartAction() {

	m.Config.AutoStart = !m.Config.AutoStart
	m.Config.Save()
	m.UIBus.Publish("model-change")

	if m.Config.AutoStart {
		utils.CheckAndInstallExe()
	}
}

func (m *UIModel) TriggerNodeEnableAction() {
	if m.Config.Enabled {
		m.TriggerAction("disable")
	} else {
		m.TriggerAction("enable")
	}
}
