/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package docker

import (
	"log"
	"time"

	model_ "github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/platform"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type Controller struct {
	a model_.AppState

	finished    bool
	mgr         model_.PlatformManager
	mystManager *myst.Manager
	lg          *log.Logger
}

func NewController() *Controller {
	lg := log.New(log.Writer(), "[docker] ", log.Ldate|log.Ltime)
	return &Controller{lg: lg}
}

func (c *Controller) GetCaps() int {
	return 1
}

func (c *Controller) SetApp(a model_.AppState) {
	c.a = a
}

func (c *Controller) GetFinished() bool {
	return c.finished
}

func (c *Controller) SetFinished() {
	c.finished = true
}

func (c *Controller) Start() {
	defer utils.PanicHandler("app-1")
	c.lg.Println("start")

	var err error
	c.mgr, err = platform.NewManager()
	if err != nil {
		panic(err) // TODO handle gracefully
	}

	mdl := c.a.GetModel()
	ui := c.a.GetUI()
	action := c.a.GetAction()

	mdl.Update()

	mystManager, err := myst.NewManager(mdl)
	if err != nil {
		panic(err) // TODO handle gracefully
	}
	c.mystManager = mystManager
	docker := NewDockerRunner(mystManager.GetDockerClient())

	defer c.SetFinished()

	t1 := time.NewTicker(15 * time.Second)
	for {
		if wantExit := c.tryStartOrInstallDocker(docker); wantExit {
			c.SetFinished()
			ui.CloseUI()
			return
		}
		mdl.SwitchState(model_.UIStateInitial)

		// docker is running now
		c.startContainer()
		if mdl.Config.AutoUpgrade {
			c.upgradeContainer(false)
		}

		// c.lg.Println("wait action >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
		select {
		case act := <-action:
			c.lg.Println("action:", act)

			switch act {
			case model_.ActionCheck:
				mystManager.CheckCurrentVersionAndUpgrades(true)

			case model_.ActionUpgrade:
				c.upgradeContainer(false)

			case model_.ActionRestart:
				// restart to apply new settings
				c.restartContainer()
				mdl.Config.Save()

			case model_.ActionEnable:
				mdl.SetStateContainer(model_.RunnableStateStarting)
				mystManager.Start()
				mdl.SetStateContainer(model_.RunnableStateRunning)

			case model_.ActionDisable:
				mdl.SetStateContainer(model_.RunnableStateUnknown)
				mystManager.Stop()

			case model_.ActionStopRunner:
				// terminate node
				mdl.SetStateContainer(model_.RunnableStateUnknown)
				mystManager.Stop()
				return

			case model_.ActionStop:
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
	mdl := c.a.GetModel()
	ui := c.a.GetUI()

	if mdl.Config.InitialState == model_.InitialStateStage1 {
		return c.tryInstallDocker()
	}

	if docker.IsRunning() {
		mdl.SetStateDocker(model_.RunnableStateRunning)
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

	if isUnderVM && !mdl.Config.CheckVMSettingsConfirm {
		ret := ui.YesNoModal("Requirements checker", "VM has been detected.\r\nPlease ensure that VT-x / EPT / IOMMU \r\nare enabled for this VM.\r\nRefer to VM settings.\r\n\r\nContinue ?")
		if ret == model_.IDNO {
			ui.TerminateWaitDialogueComplete()
			return true
		}
		mdl.Config.CheckVMSettingsConfirm = true
		mdl.Config.Save()
	}

	if needSetup {
		return c.tryInstallDocker()
	}

	isRunning, couldNotStart := docker.IsRunningOrTryStart()
	if isRunning {
		mdl.SetStateDocker(model_.RunnableStateRunning)
		return false
	}
	mdl.SetStateDocker(model_.RunnableStateStarting)
	if couldNotStart {
		mdl.SetStateDocker(model_.RunnableStateUnknown)
		return c.tryInstallDocker()
	}

	return false
}

func (c *Controller) restartContainer() {
	mdl := c.a.GetModel()

	mdl.SetStateContainer(model_.RunnableStateInstalling)
	err := c.mystManager.Restart()
	if err != nil {
		c.lg.Println("restart", err)
	}
}

func (c *Controller) upgradeContainer(refreshVersionCache bool) {
	mdl := c.a.GetModel()

	if !mdl.ImageInfo.HasUpdate {
		return
	}

	if refreshVersionCache {
		c.mystManager.CheckCurrentVersionAndUpgrades(refreshVersionCache)
	}
	mdl.SetStateContainer(model_.RunnableStateInstalling)
	err := c.mystManager.Update()
	if err != nil {
		c.lg.Println("upgrade", err)
	}

	c.mystManager.CheckCurrentVersionAndUpgrades(false)
}

// check for image updates before starting container, offer upgrade interactively
func (c *Controller) startContainer() {
	mdl := c.a.GetModel()
	ui := c.a.GetUI()

	c.mystManager.CheckCurrentVersionAndUpgrades(false)

	mdl.SetStateContainer(model_.RunnableStateInstalling)
	if !mdl.Config.Enabled {
		return
	}

	containerAlreadyRunning, err := c.mystManager.Start()
	if err != nil {
		mdl.SetStateContainer(model_.RunnableStateUnknown)
		c.lg.Println("startContainer", err)
		return
	}

	mdl.SetStateContainer(model_.RunnableStateRunning)

	if !containerAlreadyRunning && mdl.Config.InitialState == model_.InitialStateFirstRunAfterInstall {
		mdl.Config.InitialState = model_.InitialStateNormalRun
		mdl.Config.Save()

		ui.ShowNotificationInstalled()
	}
}
