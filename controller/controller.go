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
	runner model.RunnerController

	waitForShutdownReady bool
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
		if c.runner != nil {
			c.stopBackendControl()
		}
		c.runner = NewBackend(c.model.Config.Backend, c.model, c.ui)

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

	c.model.UIBus.SubscribeAsync("ready-to-shutdown", func() {
		log.Println("control> ready-to-shutdown")
		c.action <- model.ActionUpgradeGracefuly
	}, false)
}

func (c *Controller) Shutdown() {
	c.lg.Println("Shutdown >")
	c.stopBackendControl()

	// TODO: unsubscribe
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

		if c.model.Config.AgreementConsentDate.IsZero() {
			return
		}

		if !c.runner.TryStartRuntime() {
			c.runner.TryInstallRuntime_()
			return
		}
		// now we have runtime (docker) running

		hasUpdates := false
		if !c.waitForShutdownReady {
			hasUpdates = c.runner.CheckCurrentVersionAndUpgrades(false)
		}

		if !c.runner.IsRunning() {
			if c.model.Config.AutoUpgrade && hasUpdates {
				c.runner.UpgradeContainer(false)
			}
		}

		c.runner.StartContainer()
		c.lg.Println("startNode >", c.model.Config.AutoUpgrade, hasUpdates)

		if c.model.Config.AutoUpgrade && hasUpdates {

			// start and wait for SHUTDOWN READY event
			if !c.waitForShutdownReady {
				c.waitForShutdownReady = true
				c.model.Sh.Start()
			}
		}
		if !c.model.Config.AutoUpgrade && c.waitForShutdownReady {
			c.waitForShutdownReady = false
			c.model.Sh.Stop()
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
				// log.Println("<-", act)

				switch act {
				case model.ActionCheck:
					c.runner.CheckCurrentVersionAndUpgrades(true)

				case model.ActionUpgradeGracefuly:
					if c.waitForShutdownReady {
						c.waitForShutdownReady = false
						c.model.Sh.Stop()
						c.runner.UpgradeContainer(false)
					}

				case model.ActionUpgrade:
					c.runner.UpgradeContainer(false)

				case model.ActionInstall:
					c.runner.TryInstallRuntime()
					return

				case model.ActionStop:
					c.runner.StopContainer()
					return

				case model.ActionDisable:
					c.model.SetStateContainer(model.RunnableStateUnknown)
					c.runner.StopContainer()

				case model.ActionEnable:
					c.runner.StartContainer()

				case model.ActionRestart:
					c.model.SetStateContainer(model.RunnableStateUnknown)
					c.runner.RestartContainer()
				}
			}
		}
	}()
}
