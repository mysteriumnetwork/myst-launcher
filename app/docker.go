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
	runtime.LockOSThread()
	utils.Win32Initialize()
	defer s.WaitGroup.Done()

	if utils.LauncherUpgradeAvailable() {
		ret := s.ui.YesNoModal("Launcher upgrade", "You are running a newer version of launcher.\r\nUpgrade launcher installation ?")
		if ret == model.IDYES {
			utils.UpdateExe()
		}
	}

	mystManager, err := myst.NewManagerWithDefaults()
	if err != nil {
		panic(err) // TODO handle gracefully
	}
	docker := myst.NewDockerMonitor(mystManager)

	t1 := time.NewTicker(15 * time.Second)
	s.model.Update()

	for {
		tryStartOrInstallDocker := func() bool {
			log.Println("tryStartOrInstallDocker")

			isRunning_ := docker.IsRunningSimple()
			if isRunning_ {
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
				return s.tryInstall()
			}

			isRunning, couldNotStart := docker.IsRunning()
			if isRunning {
				s.model.SetStateDocker(model.RunnableStateRunning)
				return false
			}
			s.model.SetStateDocker(model.RunnableStateStarting)
			if couldNotStart {
				s.model.SetStateDocker(model.RunnableStateUnknown)
				return s.tryInstall()
			}

			return false
		}
		if wantExit := tryStartOrInstallDocker(); wantExit {
			s.model.SetWantExit()
			return
		}

		// docker is running now
		s.upgradeImageAndRun(mystManager)

		select {
		case act := <-s.action:
			log.Println("action:", act)

			switch act {
			case "check":
				s.model.ImgVer.CurrentImgDigest = mystManager.GetCurrentImageDigest()
				myst.CheckVersionAndUpgrades(s.model, false)

			case "upgrade":
				s.upgrade(mystManager)

			case "restart":
				s.restart(mystManager)
				s.model.Config.Save()

			case "enable":
				s.model.Config.Enabled = true
				s.model.Config.Save()
				s.model.SetStateContainer(model.RunnableStateStarting)
				mystManager.Start(s.model.GetConfig())
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
	err := mystManager.Restart(s.model.GetConfig())
	if err != nil {
		log.Println("restart", err)
	}
}

func (s *AppState) upgrade(mystManager *myst.Manager) {
	s.model.SetStateContainer(model.RunnableStateInstalling)
	err := mystManager.Update(s.model.GetConfig())
	if err != nil {
		log.Println("upgrade", err)
	}

	s.model.ImgVer.CurrentImgDigest = mystManager.GetCurrentImageDigest()
	myst.CheckVersionAndUpgrades(s.model, false)
}

// check for image updates before starting container, offer upgrade interactively
func (s *AppState) upgradeImageAndRun(mystManager *myst.Manager) {
	if s.model.Config.Enabled {

		s.model.SetStateContainer(model.RunnableStateUnknown)
		s.model.ImgVer.CurrentImgDigest = mystManager.GetCurrentImageDigest()
		myst.CheckVersionAndUpgrades(s.model, true)

		containerAlreadyRunning, err := mystManager.Start(s.model.GetConfig())
		if err != nil {
			return
		}
		s.model.SetStateContainer(model.RunnableStateRunning)

		if !containerAlreadyRunning && s.didInstallation {
			s.didInstallation = false
			s.ui.ShowNotificationInstalled()
		}
	}

	s.model.ImgVer.CurrentImgDigest = mystManager.GetCurrentImageDigest()
	myst.CheckVersionAndUpgrades(s.model, false)

	if s.model.Config.AutoUpgrade && s.model.ImgVer.HasUpdate {
		s.upgrade(mystManager)
	}
}
