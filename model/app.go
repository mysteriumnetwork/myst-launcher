/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package model

type App interface {
	TriggerAction(action string)
	GetInTray() bool
}

type Controller interface {
	Start()
	Shutdown()
	TriggerAction(action string)
}

type RunnerController interface {
	TryInstallRuntime() bool
	TryInstallRuntime_()
	TryStartRuntime() bool
	CheckSysRequirements() bool

	IsRunning() bool
	StartContainer()
	StopContainer()
	RestartContainer()

	UpgradeContainer(refreshVersionCache bool)
	CheckCurrentVersionAndUpgrades(refreshVersionCache bool) bool
}

const (
	ActionCheck            = "check"
	ActionUpgrade          = "upgrade"
	ActionUpgradeGracefuly = "upgrade-graceful"
	ActionRestart          = "restart"
	ActionEnable           = "enable"
	ActionDisable          = "disable"
	ActionStopRunner       = "stop-runner"
	ActionStop             = "stop"
	ActionInstall          = "install" // install runner/container
)

type AppState interface {
	StartAppController()
	StopAppController()

	GetModel() *UIModel
	GetUI() Gui_
}
