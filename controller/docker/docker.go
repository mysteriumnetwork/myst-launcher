/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package docker

import (
	"log"
	"sync"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/app"
	model_ "github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/platform"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type Controller struct {
	a *app.AppState

	mgr model_.PlatformManager
	wg  sync.WaitGroup // for graceful shutdown
}

func NewController(a *app.AppState) *Controller {
	return &Controller{
		a: a,
	}
}

func (s *Controller) Shutdown() {
	// wait for SuperviseDockerNode to finish its work
	s.wg.Wait()
}

func (s *Controller) SuperviseDockerNode() {
	defer utils.PanicHandler("app")

	var err error
	s.mgr, err = platform.NewManager()
	if err != nil {
		panic(err) // TODO handle gracefully
	}
	s.wg.Add(1)
	defer s.wg.Done()

	model := s.a.GetModel()
	ui := s.a.GetUI()
	action := s.a.GetAction()

	mystManager, err := myst.NewManager(model)
	if err != nil {
		panic(err) // TODO handle gracefully
	}
	docker := NewDockerRunner(mystManager.GetDockerClient())

	t1 := time.NewTicker(15 * time.Second)
	model.Update()

	for {
		if wantExit := s.tryStartOrInstallDocker(docker); wantExit {
			model.SetWantExit()
			ui.CloseUI()
			return
		}
		model.SwitchState(model_.UIStateInitial)

		// docker is running now
		s.startContainer(mystManager)
		if model.Config.AutoUpgrade {
			s.upgradeContainer(mystManager, false)
		}

		select {
		case act := <-action:
			log.Println("action:", act)

			switch act {
			case "check":
				mystManager.CheckCurrentVersionAndUpgrades(true)

			case "upgrade":
				s.upgradeContainer(mystManager, false)

			case "restart":
				// restart to apply new settings
				s.restart(mystManager)
				model.Config.Save()

			case "enable":
				model.Config.Enabled = true
				model.Config.Save()
				model.SetStateContainer(model_.RunnableStateStarting)
				mystManager.Start()
				model.SetStateContainer(model_.RunnableStateRunning)

			case "disable":
				model.Config.Enabled = false
				model.Config.Save()
				model.SetStateContainer(model_.RunnableStateUnknown)
				mystManager.Stop()

			case "stop":
				return
			}

		// wait for ticker event if no action
		case <-t1.C:
		}
	}
}

// returns: will exit, if tryInstallDocker requests it
func (s *Controller) tryStartOrInstallDocker(docker *DockerRunner) bool {
	log.Println("tryStartOrInstallDocker")
	model := s.a.GetModel()
	ui := s.a.GetUI()

	if model.Config.InitialState == model_.InitialStateStage1 {
		return s.tryInstallDocker()
	}

	if docker.IsRunning() {
		model.SetStateDocker(model_.RunnableStateRunning)
		return false
	}

	// In case of suspend/resume some APIs may return unexpected error, so we need to retry it
	isUnderVM, needSetup, err := false, false, error(nil)
	err = utils.Retry(3, time.Second, func() error {
		isUnderVM, err = s.mgr.SystemUnderVm()
		if err != nil {
			return err
		}

		featuresOk, err := s.mgr.Features()
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
		log.Println("error", err)
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
		return s.tryInstallDocker()
	}

	isRunning, couldNotStart := docker.IsRunningOrTryStart()
	if isRunning {
		model.SetStateDocker(model_.RunnableStateRunning)
		return false
	}
	model.SetStateDocker(model_.RunnableStateStarting)
	if couldNotStart {
		model.SetStateDocker(model_.RunnableStateUnknown)
		return s.tryInstallDocker()
	}

	return false
}

func (s *Controller) restart(mystManager *myst.Manager) {
	model := s.a.GetModel()

	model.SetStateContainer(model_.RunnableStateInstalling)
	err := mystManager.Restart()
	if err != nil {
		log.Println("restart", err)
	}
}

func (s *Controller) upgradeContainer(mystManager *myst.Manager, refreshVersionCache bool) {
	model := s.a.GetModel()

	if !model.ImageInfo.HasUpdate {
		return
	}

	if refreshVersionCache {
		mystManager.CheckCurrentVersionAndUpgrades(refreshVersionCache)
	}
	model.SetStateContainer(model_.RunnableStateInstalling)
	err := mystManager.Update()
	if err != nil {
		log.Println("upgrade", err)
	}

	mystManager.CheckCurrentVersionAndUpgrades(false)
}

// check for image updates before starting container, offer upgrade interactively
func (c *Controller) startContainer(mystManager *myst.Manager) {
	model := c.a.GetModel()
	ui := c.a.GetUI()

	mystManager.CheckCurrentVersionAndUpgrades(false)

	model.SetStateContainer(model_.RunnableStateInstalling)
	if model.Config.Enabled {
		containerAlreadyRunning, err := mystManager.Start()
		if err != nil {
			model.SetStateContainer(model_.RunnableStateUnknown)
			log.Println("startContainer", err)

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
