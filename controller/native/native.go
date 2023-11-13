package native

import (
	"log"

	"github.com/mysteriumnetwork/myst-launcher/model"
)

type Native_ struct {
	lg *log.Logger

	model  *model.UIModel
	ui     model.Gui_
	runner *NodeRunner
}

func NewSvc(m *model.UIModel, ui model.Gui_) *Native_ {
	lg := log.New(log.Writer(), "[native_] ", log.Ldate|log.Ltime)

	m.Caps = 1
	m.UIBus.Publish("model-change")

	return &Native_{
		lg:     lg,
		model:  m,
		ui:     ui,
		runner: NewRunner(m),
	}
}

func (c *Native_) TryStartRuntime() bool {
	return true
}

func (c *Native_) CheckSysRequirements() bool {
	return true
}

// check for image updates before starting container, offer upgrade interactively
func (c *Native_) StartContainer() {
	mdl := c.model
	ui := c.ui

	if !mdl.Config.Enabled {
		return
	}
	tryInstallFirewallRules(ui)

	running := c.runner.IsRunningOrTryStart()
	if running {
		mdl.SetStateContainer(model.RunnableStateRunning)
	} else {
		mdl.SetStateContainer(model.RunnableStateUnknown)
	}

	if running {
		cfg := &mdl.Config
		switch cfg.InitialState {
		case model.InitialStateFirstRunAfterInstall, model.InitialStateUndefined:
			cfg.InitialState = model.InitialStateNormalRun
			cfg.Save()

			if ui != nil {
				ui.ShowNotificationInstalled()
				ui.OpenNodeUI()
			} else {
				c.lg.Println("node installed!")
			}
		}
	}
}

func (c *Native_) StopContainer() {
	c.model.SetStateContainer(model.RunnableStateUnknown)
	c.runner.Stop()
}

func (c *Native_) RestartContainer() {
	c.model.SetStateContainer(model.RunnableStateInstalling)

	c.runner.Stop()
	running := c.runner.IsRunningOrTryStart()
	if running {
		c.model.SetStateContainer(model.RunnableStateRunning)
	} else {
		c.model.SetStateContainer(model.RunnableStateUnknown)
	}
}

func (c *Native_) UpgradeContainer(refreshVersionCache bool) {
	//c.lg.Println("!upgradeContainer")

	// if !c.model.ImageInfo.HasUpdate {
	// 	return
	// }

	c.CheckAndUpgradeNodeExe(refreshVersionCache)
	//model.SetStateContainer(model_.RunnableStateRunning)
}

func (c *Native_) TryInstallRuntime_() {}

func (c *Native_) TryInstallRuntime() bool {
	return false
}
