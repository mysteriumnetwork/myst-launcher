/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package app

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const group = "docker-users"

func (s *AppState) SuperviseDockerNode() {
	runtime.LockOSThread()
	utils.Win32Initialize()
	defer s.WaitGroup.Done()

	fmt.Println("SuperviseDockerNode >")

	if utils.LauncherUpgradeAvailable() {
		//	ret := s.ui.YesNoModal("Launcher upgrade", "You are running a newer version of launcher.\r\n\r\nUpgrade launcher installation ?")
		//	if ret == gui.IDYES {
		//		utils.UpdateExe()
		//	}
	}

	mystManager, err := myst.NewManagerWithDefaults()
	if err != nil {
		panic(err) // TODO handle gracefully
	}
	docker := myst.NewDockerMonitor(mystManager)

	t1 := time.Tick(15 * time.Second)
	s.model.Update()

	for {
		tryStartOrInstallDocker := func() bool {

			if isRunning, _ := docker.IsRunning(); isRunning {
				s.model.SetStateDocker(gui.RunnableStateRunning)
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
					return nil
				}
				return nil
			})
			if err != nil {
				log.Println("error", err)
				s.ui.ErrorModal("Application error", err.Error())
				return true
			}
			fmt.Println("tryStartOrInstallDocker")

			if isUnderVM && !s.model.Config.CheckVMSettingsConfirm {
				ret := s.ui.YesNoModal("Requirements checker", "VM has been detected. \r\n\r\nPlease ensure that VT-x / EPT / IOMMU \r\nare enabled for this VM.\r\nRefer to VM settings.\r\n\r\nContinue ?")
				if ret == gui.IDNO {
					s.model.ExitApp()
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
				s.model.SetStateDocker(gui.RunnableStateRunning)
				return false
			}
			s.model.SetStateDocker(gui.RunnableStateStarting)
			if couldNotStart {
				s.model.SetStateDocker(gui.RunnableStateUnknown)
				return s.tryInstall()
			}

			return false
		}
		fmt.Println("tryStartOrInstallDocker")
		wantExit := tryStartOrInstallDocker()
		if wantExit {
			s.model.SetWantExit()
			return
		}

		// docker is running now
		s.upgradeImageAndRun(mystManager)

		select {
		case act := <-s.action:
			switch act {
			case "check":
				s.model.ImgVer.CurrentImgDigest = mystManager.GetCurrentImageDigest()
				myst.CheckVersionAndUpgrades(s.model)

			case "upgrade":
				s.upgrade(mystManager)

			case "enable":
				s.model.Config.Enabled = true
				s.model.Config.Save()
				s.model.SetStateContainer(gui.RunnableStateRunning)
				mystManager.Start(s.model.GetConfig())

			case "disable":
				s.model.Config.Enabled = false
				s.model.Config.Save()
				s.model.SetStateContainer(gui.RunnableStateUnknown)
				mystManager.Stop()

			case "stop":
				fmt.Println("action: stop")
				return
			}

		// wait for ticker event if no action
		case <-t1:
		}
	}
}

func (s *AppState) upgrade(mystManager *myst.Manager) {
	s.model.SetStateContainer(gui.RunnableStateUnknown)
	mystManager.Stop()

	s.model.SetStateContainer(gui.RunnableStateInstalling)
	mystManager.Update(s.model.GetConfig())

	s.model.ImgVer.CurrentImgDigest = mystManager.GetCurrentImageDigest()
	myst.CheckVersionAndUpgrades(s.model)
}

// check for image updates before starting container, offer upgrade interactively
func (s *AppState) upgradeImageAndRun(mystManager *myst.Manager) {
	imageDigest := mystManager.GetCurrentImageDigest()

	if s.model.ImgVer.CurrentImgDigest != imageDigest || s.model.ImgVer.VersionCurrent == "" || s.model.Config.NeedToCheckUpgrade() {
		// docker has a new image (in result the of external command)

		s.model.ImgVer.CurrentImgDigest = imageDigest
		myst.CheckVersionAndUpgrades(s.model)
	}
	if s.model.Config.AutoUpgrade && s.model.ImgVer.HasUpdate {
		s.upgrade(mystManager)
	}

	if s.model.Config.Enabled {
		s.model.SetStateContainer(gui.RunnableStateUnknown)
		containerAlreadyRunning, err := mystManager.Start(s.model.GetConfig())
		if err != nil {
			return
		}

		s.model.SetStateContainer(gui.RunnableStateRunning)
		s.model.Update()

		if !containerAlreadyRunning && s.didInstallation {
			s.didInstallation = false

			s.ui.ShowNotificationInstalled()
		}
	}
}
