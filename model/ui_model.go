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
	CheckWindowsVersion  InstallStep
	CheckVirt            InstallStep
	CheckDocker          InstallStep // darwin
	InstallExecutable    InstallStep
	RebootAfterWSLEnable InstallStep
	DownloadFiles        InstallStep
	InstallWSLUpdate     InstallStep
	InstallDocker        InstallStep
	CheckGroupMembership InstallStep

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

func (m *UIModel) ResetProperties() {
	m.CheckWindowsVersion = StepNone
	m.CheckVirt = StepNone
	m.CheckDocker = StepNone
	m.InstallExecutable = StepNone
	m.RebootAfterWSLEnable = StepNone
	m.DownloadFiles = StepNone
	m.InstallWSLUpdate = StepNone
	m.InstallDocker = StepNone
	m.CheckGroupMembership = StepNone

	m.UIBus.Publish("model-change")
}

func (m *UIModel) UpdateProperties(p UIProps) {
	for k, v := range p {
		v := v.(InstallStep)

		switch k {
		case "CheckWindowsVersion":
			m.CheckWindowsVersion = v
		case "CheckVTx":
			m.CheckVirt = v
		case "CheckDocker":
			m.CheckDocker = v
		case "InstallExecutable":
			m.InstallExecutable = v
		case "RebootAfterWSLEnable":
			m.RebootAfterWSLEnable = v
		case "DownloadFiles":
			m.DownloadFiles = v
		case "InstallWSLUpdate":
			m.InstallWSLUpdate = v
		case "InstallDocker":
			m.InstallDocker = v
		case "CheckGroupMembership":
			m.CheckGroupMembership = v
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
		m.UIBus.Publish("container-state")
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
