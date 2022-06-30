/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package docker

import (
	"log"
	"os"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/app"
	model_ "github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/platform"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type Controller struct {
	a *app.AppState

	mgr         model_.PlatformManager
	mystManager *myst.Manager
	lg          *log.Logger
}

func NewController() *Controller {
	lg := log.New(os.Stdout, "[docker] ", log.Ldate|log.Ltime)
	return &Controller{lg: lg}
}

func (c *Controller) GetCaps() int {
	return 1
}

func (c *Controller) SetApp(a *app.AppState) {
	c.a = a
}

// func (c *Controller) Shutdown() {}

func (c *Controller) Start() {
	defer utils.PanicHandler("app")
	c.lg.Println("start")

	var err error
	c.mgr, err = platform.NewManager()
	if err != nil {
		panic(err) // TODO handle gracefully
	}

	model := c.a.GetModel()
	ui := c.a.GetUI()
	action := c.a.GetAction()

	model.Update()

	mystManager, err := myst.NewManager(model)
	if err != nil {
		panic(err) // TODO handle gracefully
	}
	c.mystManager = mystManager
	docker := NewDockerRunner(mystManager.GetDockerClient())

	t1 := time.NewTicker(15 * time.Second)
	for {
		if wantExit := c.tryStartOrInstallDocker(docker); wantExit {
			model.SetWantExit()
			ui.CloseUI()
			return
		}
		model.SwitchState(model_.UIStateInitial)

		// docker is running now
		c.startContainer()
		if model.Config.AutoUpgrade {
			c.upgradeContainer(false)
		}

		select {
		case act := <-action:
			c.lg.Println("action:", act)

			switch act {
			case app.ActionCheck:
				mystManager.CheckCurrentVersionAndUpgrades(true)

			case app.ActionUpgrade:
				c.upgradeContainer(false)

			case app.ActionRestart:
				// restart to apply new settings
				c.restartContainer()
				model.Config.Save()

			case app.ActionEnable:
				model.SetStateContainer(model_.RunnableStateStarting)
				mystManager.Start()
				model.SetStateContainer(model_.RunnableStateRunning)

			case app.ActionDisable:
				model.SetStateContainer(model_.RunnableStateUnknown)
				mystManager.Stop()

			case app.ActionStopRunner:
				// terminate node
				model.SetStateContainer(model_.RunnableStateUnknown)
				mystManager.Stop()
				return

			case app.ActionStop:
				c.lg.Println("[docker] stop")
				return
			}

		// wait for ticker event if no action
		case <-t1.C:
		}
	}
}

// returns: will exit, if tryInstallDocker requests it
func (c *Controller) tryStartOrInstallDocker(docker *DockerRunner) bool {
	c.lg.Println("tryStartOrInstallDocker")
	model := c.a.GetModel()
	ui := c.a.GetUI()

	if model.Config.InitialState == model_.InitialStateStage1 {
		return c.tryInstallDocker()
	}

	if docker.IsRunning() {
		model.SetStateDocker(model_.RunnableStateRunning)
		return false
	}

	// In case of suspend/resume some APIs may return unexpected error, so we need to retry it
	isUnderVM, needSetup, err := false, false, error(nil)
	err = utils.Retry(3, time.Second, func() error {
		isUnderVM, err = c.mgr.SystemUnderVm()
		if err != nil {
			return err
		}

		featuresOk, err := c.mgr.Features()
		if err != nil {
			return err
		}
		if !featuresOk {
			needSetup = true
		}

		hasDocker, _ := utils.HasDocker()
		if !hasDocker {
			needSetup = true
		}
		return nil
	})
	if err != nil {
		c.lg.Println("error", err)
		ui.ErrorModal("Application error", err.Error())
		return true
	}

	if isUnderVM && !model.Config.CheckVMSettingsConfirm {
		ret := ui.YesNoModal("Requirements checker", "VM has been detected.\r\nPlease ensure that VT-x / EPT / IOMMU \r\nare enabled for this VM.\r\nRefer to VM settings.\r\n\r\nContinue ?")
		if ret == model_.IDNO {
			ui.TerminateWaitDialogueComplete()
			return true
		}
		model.Config.CheckVMSettingsConfirm = true
		model.Config.Save()
	}
	if needSetup {
		return c.tryInstallDocker()
	}

	isRunning, couldNotStart := docker.IsRunningOrTryStart()
	if isRunning {
		model.SetStateDocker(model_.RunnableStateRunning)
		return false
	}
	model.SetStateDocker(model_.RunnableStateStarting)
	if couldNotStart {
		model.SetStateDocker(model_.RunnableStateUnknown)
		return c.tryInstallDocker()
	}

	return false
}

func (c *Controller) restartContainer() {
	model := c.a.GetModel()

	model.SetStateContainer(model_.RunnableStateInstalling)
	err := c.mystManager.Restart()
	if err != nil {
		c.lg.Println("restart", err)
	}
}

func (c *Controller) upgradeContainer(refreshVersionCache bool) {
	model := c.a.GetModel()

	if !model.ImageInfo.HasUpdate {
		return
	}

	if refreshVersionCache {
		c.mystManager.CheckCurrentVersionAndUpgrades(refreshVersionCache)
	}
	model.SetStateContainer(model_.RunnableStateInstalling)
	err := c.mystManager.Update()
	if err != nil {
		c.lg.Println("upgrade", err)
	}

	c.mystManager.CheckCurrentVersionAndUpgrades(false)
}

// check for image updates before starting container, offer upgrade interactively
func (c *Controller) startContainer() {
	model := c.a.GetModel()
	ui := c.a.GetUI()

	c.mystManager.CheckCurrentVersionAndUpgrades(false)

	model.SetStateContainer(model_.RunnableStateInstalling)
	if model.Config.Enabled {
		containerAlreadyRunning, err := c.mystManager.Start()
		if err != nil {
			model.SetStateContainer(model_.RunnableStateUnknown)
			c.lg.Println("startContainer", err)

			return
		}
		model.SetStateContainer(model_.RunnableStateRunning)

		if !containerAlreadyRunning && model.Config.InitialState == model_.InitialStateFirstRunAfterInstall {
			model.Config.InitialState = model_.InitialStateNormalRun
			model.Config.Save()

			ui.ShowNotificationInstalled()
		}
	}
}
