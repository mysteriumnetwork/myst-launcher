/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package native

import (
	"log"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/app"
	model_ "github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type Controller struct {
	a      *app.AppState
	runner *NodeRunner
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) SetApp(a *app.AppState) {
	c.a = a
	c.runner = NewRunner(&a.GetModel().Config)
}

func (c *Controller) Shutdown() {}

// Supervise the node
func (c *Controller) Start() {
	defer utils.PanicHandler("app")
	log.Println("[native] start")

	model := c.a.GetModel()
	action := c.a.GetAction()

	t1 := time.NewTicker(15 * time.Second)
	model.Update()

	for {
		model.SwitchState(model_.UIStateInitial)

		c.startContainer()
		if model.Config.AutoUpgrade {
			c.upgradeContainer(false)
		}

		select {
		case act := <-action:
			log.Println("action:", act)

			switch act {
			case "check":
				c.upgradeContainer(true)

			case "upgrade":
				c.upgradeContainer(false)

			case "restart":
				// restart to apply new settings
				c.restartContainer()
				model.Config.Save()

			case "enable":
				model.SetStateContainer(model_.RunnableStateStarting)
				c.startContainer()
				model.SetStateContainer(model_.RunnableStateRunning)

			case "disable":
				model.SetStateContainer(model_.RunnableStateUnknown)
				c.stop()

			case "stop":
				log.Println("[native] stop")
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
	c.runner.IsRunningOrTryStart()
	model.SetStateContainer(model_.RunnableStateRunning)
}

func (c *Controller) stop() {
	c.runner.Stop()
}

func (c *Controller) upgradeContainer(refreshVersionCache bool) {
	model := c.a.GetModel()

	// if !model.ImageInfo.HasUpdate {
	// 	return
	// }
	model.SetStateContainer(model_.RunnableStateInstalling)
	c.CheckAndUpgradeNodeExe(refreshVersionCache)
	model.SetStateContainer(model_.RunnableStateRunning)
}

// check for image updates before starting container, offer upgrade interactively
func (c *Controller) startContainer() {
	log.Println("startContainer >")
	model := c.a.GetModel()
	//ui := c.a.GetUI()

	model.SetStateContainer(model_.RunnableStateInstalling)
	if model.Config.Enabled {
		// c.CheckAndUpgradeNodeExe(false)

		running := c.runner.IsRunningOrTryStart()
		if !running {
			// model.SetStateContainer(model_.RunnableStateUnknown)
		}
		model.SetStateContainer(model_.RunnableStateRunning)
	}
}
