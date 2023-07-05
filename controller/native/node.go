/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package native

import (
	"log"
	"sync"
	"time"

	model_ "github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type Controller struct {
	a model_.AppState

	finished bool
	wg       sync.WaitGroup

	runner *NodeRunner
	lg     *log.Logger
}

func (c *Controller) setFinished() {
	c.finished = true
	c.wg.Done()
}

func NewController() *Controller {
	lg := log.New(log.Writer(), "[native] ", log.Ldate|log.Ltime)
	return &Controller{lg: lg}
}

func (c *Controller) GetCaps() int {
	return 0
}

func (c *Controller) SetApp(a model_.AppState) {
	c.a = a
	c.runner = NewRunner(a.GetModel())
}

func (c *Controller) Shutdown() {
	if !c.finished {
		c.a.GetAction() <- model_.ActionStop
		c.wg.Wait()
	}
}

// Supervise the node
func (c *Controller) Start() {
	defer utils.PanicHandler("app-2")
	c.lg.Println("start")

	model := c.a.GetModel()
	action := c.a.GetAction()
	cfg := model.Config

	// copy version info to ui model
	model.ImageInfo.VersionCurrent = cfg.NodeExeVersion
	model.ImageInfo.VersionLatest = cfg.NodeLatestTag
	model.Update()

	c.wg.Add(1)
	defer c.setFinished()

	t1 := time.NewTicker(15 * time.Second)
	for {
		model.SwitchState(model_.UIStateInitial)

		c.startContainer()
		c.upgradeContainer(false)
		// if model.Config.AutoUpgrade {
		// c.upgradeContainer(false)
		// }

		select {
		case act := <-action:
			c.lg.Println("action:", act)

			switch act {
			case model_.ActionCheck:
				c.upgradeContainer(true)

			case model_.ActionUpgrade:
				c.upgradeContainer(false)

			case model_.ActionRestart:
				// restart to apply new settings
				c.restartContainer()
				model.Config.Save()

			case model_.ActionEnable:
				c.startContainer()

			case model_.ActionDisable:
				c.stop()

			case model_.ActionStopRunner:
				// terminate controller
				c.stop()
				return

			case model_.ActionStop:
				c.lg.Println("[native] stop")
				c.stop()
				return
			}

		// wait for ticker event if no action
		case <-t1.C:
		}
	}
}

func (c *Controller) restartContainer() {
	model := c.a.GetModel()
	model.SetStateContainer(model_.RunnableStateInstalling)

	c.runner.Stop()
	running := c.runner.IsRunningOrTryStart()
	if running {
		model.SetStateContainer(model_.RunnableStateRunning)
	} else {
		model.SetStateContainer(model_.RunnableStateUnknown)
	}
}

func (c *Controller) stop() {
	model := c.a.GetModel()
	model.SetStateContainer(model_.RunnableStateUnknown)

	c.runner.Stop()
}

func (c *Controller) upgradeContainer(refreshVersionCache bool) {
	//c.lg.Println("!upgradeContainer")
	// model := c.a.GetModel()
	// if !model.ImageInfo.HasUpdate {
	// 	return
	// }

	c.CheckAndUpgradeNodeExe(refreshVersionCache)
	//model.SetStateContainer(model_.RunnableStateRunning)
}

// check for image updates before starting container, offer upgrade interactively
func (c *Controller) startContainer() {
	//c.lg.Println("!run")
	model := c.a.GetModel()

	if !model.Config.Enabled {
		return
	}

	ui := c.a.GetUI()
	tryInstallFirewallRules(ui)

	running := c.runner.IsRunningOrTryStart()
	if running {
		model.SetStateContainer(model_.RunnableStateRunning)
	} else {
		model.SetStateContainer(model_.RunnableStateUnknown)
	}

	if running {
		cfg := &model.Config
		switch cfg.InitialState {
		case model_.InitialStateFirstRunAfterInstall, model_.InitialStateUndefined:
			cfg.InitialState = model_.InitialStateNormalRun
			cfg.Save()
			if ui != nil {
				ui.ShowNotificationInstalled()
			} else {
				c.lg.Println("node installed!")
			}
		}
	}
}
