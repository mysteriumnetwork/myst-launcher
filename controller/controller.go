/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package controller

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/model"
)

// Main controller
type Controller struct {
	lg *log.Logger
	wg sync.WaitGroup

	model  *model.UIModel
	ui     model.Gui_
	action chan string
	d      model.RunnerController
}

func NewController(m *model.UIModel, ui model.Gui_, a model.AppState) *Controller {
	lg := log.New(log.Writer(), "[ctrl] ", log.Ldate|log.Ltime)

	c := &Controller{
		lg:     lg,
		model:  m,
		ui:     ui,
		action: make(chan string, 1),
	}
	return c
}

func (c *Controller) Start() {
	//c.lg.Println("start")

	restartBackendControl := func() {
		if c.d != nil {
			c.stopBackendControl()
		}
		c.d = NewBackend(c.model.Config.Backend, c.model, c.ui)

		c.model.SwitchState(model.UIStateInitial)
		c.startBackendControl()
	}
	restartBackendControl()
	c.model.Bus2.Subscribe("backend", restartBackendControl)

	c.model.UIBus.SubscribeAsync("install-dlg-exit", func(id int) {
		fmt.Println("install-dlg-exit>", id, c.model.State)

		switch c.model.State {
		case model.UIStateInstallNeeded:
			c.action <- model.ActionInstall

		case model.UIStateInstallFinished:
			c.model.SwitchState(model.UIStateInitial)
			// restart controller as it stops after installation
			restartBackendControl()

		case model.UIStateInstallError:
			c.ui.CloseUI()

		default:
			c.ui.CloseUI()
		}
	}, false)

	c.model.UIBus.SubscribeAsync("dlg-exit", func() {
		c.ui.CloseUI()
	}, false)
}

func (c *Controller) Shutdown() {
	c.lg.Println("Shutdown >")
	c.stopBackendControl()
}

func (c *Controller) TriggerAction(action string) {
	c.action <- action
}

/////////////////////////////////////////////////////////////////////////

func (c *Controller) stopBackendControl() {
	c.lg.Println("stop")

	c.action <- model.ActionStop
	c.wg.Wait()
}

func (c *Controller) startBackendControl() {
	c.lg.Println("start")
	c.wg.Add(1)

	startNode := func() {
		c.lg.Println("startNode >")

		if !c.d.TryStartRuntime() {
			c.d.TryInstallRuntime_()
			return
		}
		// now we have runtime (docker) running

		c.d.StartContainer()
		if c.model.Config.AutoUpgrade {
			c.d.UpgradeContainer(false)
		}
	}

	go func() {
		defer c.wg.Done()
		t1 := time.NewTicker(15 * time.Second)

		startNode()
		for {
			select {
			case <-t1.C:
				startNode()

			case act := <-c.action:
				fmt.Println("<-", act)

				switch act {
                case model.ActionCheck:
                    c.d.CheckCurrentVersionAndUpgrades(true)

                case model.ActionUpgrade:
                    c.d.UpgradeContainer(false)

				case model.ActionInstall:
					c.d.TryInstallRuntime()
					return

				case model.ActionStop:
					c.d.StopContainer()
					return

				case model.ActionDisable:
					c.model.SetStateContainer(model.RunnableStateUnknown)
					c.d.StopContainer()

				case model.ActionEnable:
					c.d.StartContainer()

				case model.ActionRestart:
					c.model.SetStateContainer(model.RunnableStateUnknown)
					c.d.RestartContainer()
				}
			}
		}
	}()
}
