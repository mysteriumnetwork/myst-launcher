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
	lg := log.New(log.Writer(), "[native] ", log.Ldate|log.Ltime)

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

func (r *Native_) IsRunning() bool {
	return r.runner.IsRunning()
}

// check for image updates before starting container //, offer upgrade interactively
func (c *Native_) StartContainer() {
	mdl := c.model
	cfg := &mdl.Config
	ui := c.ui

	if !mdl.Config.Enabled {
		return
	}

	tryInstallFirewallRules(ui)

	ok := c.runner.IsRunningOrTryStart()
	if ok {
		mdl.SetStateContainer(model.RunnableStateRunning)
	} else {
		mdl.SetStateContainer(model.RunnableStateUnknown)
	}

	if ok {
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
	
	ok := c.runner.IsRunningOrTryStart()
	if ok {
		c.model.SetStateContainer(model.RunnableStateRunning)
	} else {
		c.model.SetStateContainer(model.RunnableStateUnknown)
	}
}

func (c *Native_) UpgradeContainer(refreshVersionCache bool) {
	//c.lg.Println("!upgradeContainer")

	ok := c.CheckAndUpgradeNodeExe_(refreshVersionCache, true)
	if ok {
		c.model.SetStateContainer(model.RunnableStateRunning)
	} else {
		c.model.SetStateContainer(model.RunnableStateUnknown)
	}
}

func (c *Native_) CheckCurrentVersionAndUpgrades(refreshVersionCache bool) bool {
	return c.CheckAndUpgradeNodeExe_(refreshVersionCache, false)
}

// omit
func (c *Native_) TryInstallRuntime_() {}

// omit
func (c *Native_) TryInstallRuntime() bool {
	return false
}
