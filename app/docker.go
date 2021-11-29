/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package app

import (
	"log"
	"runtime"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

func (s *AppState) SuperviseDockerNode() {
	defer utils.PanicHandler("app")

	runtime.LockOSThread()
	utils.Win32Initialize()
	defer s.WaitGroup.Done()

	mystManager, err := myst.NewManager(s.model)
	if err != nil {
		panic(err) // TODO handle gracefully
	}
	docker := NewDockerMonitor(mystManager.GetDockerClient())

	t1 := time.NewTicker(15 * time.Second)
	s.model.Update()

	for {
		tryStartOrInstallDocker := func() bool {
			log.Println("tryStartOrInstallDocker")

			if docker.IsRunning() {
				s.model.SetStateDocker(model.RunnableStateRunning)
				return false
			}

			// In case of suspend/resume some APIs may return unexpected error, so we need to retry it
			isUnderVM, needSetup, err := false, false, error(nil)
			err = utils.Retry(3, time.Second, func() error {
				isUnderVM, err = utils.SystemUnderVm()
				if err != nil {
					return err
				}
				features, err := utils.QueryFeatures()
				if err != nil {
					return err
				}
				if len(features) > 0 {
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
				s.ui.ErrorModal("Application error", err.Error())
				return true
			}

			if isUnderVM && !s.model.Config.CheckVMSettingsConfirm {
				ret := s.ui.YesNoModal("Requirements checker", "VM has been detected.\r\nPlease ensure that VT-x / EPT / IOMMU \r\nare enabled for this VM.\r\nRefer to VM settings.\r\n\r\nContinue ?")
				if ret == model.IDNO {
					s.ui.TerminateWaitDialogueComplete()
					s.ui.CloseUI()
					return true
				}
				s.model.Config.CheckVMSettingsConfirm = true
				s.model.Config.Save()
			}

			if needSetup {
				return s.tryInstallDocker()
			}

			isRunning, couldNotStart := docker.IsRunningOrTryStart()
			if isRunning {
				s.model.SetStateDocker(model.RunnableStateRunning)
				return false
			}
			s.model.SetStateDocker(model.RunnableStateStarting)
			if couldNotStart {
				s.model.SetStateDocker(model.RunnableStateUnknown)
				return s.tryInstallDocker()
			}

			return false
		}
		if wantExit := tryStartOrInstallDocker(); wantExit {
			s.model.SetWantExit()
			return
		}

		// docker is running now
		s.startContainer(mystManager)
		if s.model.Config.AutoUpgrade {
			s.upgrade(mystManager)
		}

		select {
		case act := <-s.action:
			log.Println("action:", act)

			switch act {
			case "check":
				mystManager.CheckCurrentVersionAndUpgrades()

			case "upgrade":
				s.upgrade(mystManager)

			case "restart":
				// restart to apply new settings
				s.restart(mystManager)
				s.model.Config.Save()

			case "enable":
				s.model.Config.Enabled = true
				s.model.Config.Save()
				s.model.SetStateContainer(model.RunnableStateStarting)
				mystManager.Start()
				s.model.SetStateContainer(model.RunnableStateRunning)

			case "disable":
				s.model.Config.Enabled = false
				s.model.Config.Save()
				s.model.SetStateContainer(model.RunnableStateUnknown)
				mystManager.Stop()

			case "stop":
				return
			}

		// wait for ticker event if no action
		case <-t1.C:
		}
	}
}

func (s *AppState) restart(mystManager *myst.Manager) {
	s.model.SetStateContainer(model.RunnableStateInstalling)
	err := mystManager.Restart()
	if err != nil {
		log.Println("restart", err)
	}
}

func (s *AppState) upgrade(mystManager *myst.Manager) {
	if !s.model.ImageInfo.HasUpdate {
		return
	}

	s.model.SetStateContainer(model.RunnableStateInstalling)
	err := mystManager.Update()
	if err != nil {
		log.Println("upgrade", err)
	}

	mystManager.CheckCurrentVersionAndUpgrades()
}

// check for image updates before starting container, offer upgrade interactively
func (s *AppState) startContainer(mystManager *myst.Manager) {

	mystManager.CheckCurrentVersionAndUpgrades()

	if s.model.Config.Enabled {
		containerAlreadyRunning, err := mystManager.Start()
		if err != nil {
			s.model.SetStateContainer(model.RunnableStateUnknown)
			log.Println("startContainer", err)
			return
		}
		s.model.SetStateContainer(model.RunnableStateRunning)

		if !containerAlreadyRunning && s.didDockerInstallation {
			s.didDockerInstallation = false
			s.ui.ShowNotificationInstalled()
		}
	}

}
