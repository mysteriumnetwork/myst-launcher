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
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

const group = "docker-users"

func (s *AppState) SuperviseDockerNode() {
	runtime.LockOSThread()
	utils.Win32Initialize()
	defer s.WaitGroup.Done()

	if utils.LauncherUpgradeAvailable() {
		ret := s.ui.YesNoModal("Launcher upgrade", "You are running a newer version of launcher.\r\n\r\nUpgrade launcher installation ?")
		if ret == gui.IDYES {
			exePath, _ := os.Executable()
			err := native.ShellExecuteAndWait(0, "runas", exePath, FlagInstall, "", syscall.SW_NORMAL)
			if err != nil {
				log.Println("Failed to install exe", err)
			}
		}
	}

	mystManager, err := myst.NewManagerWithDefaults()
	if err != nil {
		panic(err) // TODO handle gracefully
	}
	docker := myst.NewDockerMonitor(mystManager)

	t1 := time.Tick(15 * time.Second)
	s.Config.Read()
	s.mod.Update()

	for {
		tryStartOrInstallDocker := func() bool {
			if isRunning, _ := docker.IsRunning(); isRunning {
				s.mod.SetStateDocker(gui.RunnableStateRunning)
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

			if isUnderVM && !s.Config.CheckVMSettingsConfirm {
				ret := s.ui.YesNoModal("Requirements checker", "VM has been detected. \r\n\r\nPlease ensure that VT-x / EPT / IOMMU \r\nare enabled for this VM.\r\nRefer to VM settings.\r\n\r\nContinue ?")
				if ret == gui.IDNO {
					s.mod.ExitApp()
					return true
				}
				s.Config.CheckVMSettingsConfirm = true
				s.Config.Save()
			}
			if needSetup {
				return s.tryInstall()
			}

			isRunning, couldNotStart := docker.IsRunning()
			if isRunning {
				s.mod.SetStateDocker(gui.RunnableStateRunning)
				return false
			}
			s.mod.SetStateDocker(gui.RunnableStateStarting)
			if couldNotStart {
				s.mod.SetStateDocker(gui.RunnableStateUnknown)
				return s.tryInstall()
			}

			return false
		}
		wantExit := tryStartOrInstallDocker()
		if wantExit {
			s.mod.SetWantExit()
			return
		}

		// docker is running now
		s.upgradeImageAndRun(mystManager)

		select {
		case act := <-s.action:
			switch act {
			case "check":
				s.ImgVer.CurrentImgDigest = mystManager.GetCurrentImageDigest()
				ok := myst.CheckVersionAndUpgrades(&s.ImgVer, &s.Config)
				if !ok {
					break
				}
				s.Config.Save()
				s.mod.Update()

			case "upgrade":
				s.upgrade(mystManager)

			case "enable":
				s.Config.Enabled = true
				s.Config.Save()
				s.mod.SetStateContainer(gui.RunnableStateRunning)
				mystManager.Start(s.GetConfig())

			case "disable":
				s.Config.Enabled = false
				s.Config.Save()
				s.mod.SetStateContainer(gui.RunnableStateUnknown)
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
	s.mod.SetStateContainer(gui.RunnableStateUnknown)
	mystManager.Stop()

	s.mod.SetStateContainer(gui.RunnableStateInstalling)
	mystManager.Update(s.GetConfig())

	s.ImgVer.CurrentImgDigest = mystManager.GetCurrentImageDigest()
	ok := myst.CheckVersionAndUpgrades(&s.ImgVer, &s.Config)
	if ok {
		s.Config.Save()
	}
	s.mod.Update()
}

// check for image updates before starting container, offer upgrade interactively
func (s *AppState) upgradeImageAndRun(mystManager *myst.Manager) {
	imageDigest := mystManager.GetCurrentImageDigest()

	if s.ImgVer.CurrentImgDigest != imageDigest || s.ImgVer.VersionCurrent == "" || s.Config.NeedToCheckUpgrade() {
		// docker has a new image (in result the of external command)

		s.ImgVer.CurrentImgDigest = imageDigest
		ok := myst.CheckVersionAndUpgrades(&s.ImgVer, &s.Config)
		if ok {
			s.Config.Save()
		}
		s.mod.Update()
	}
	if s.Config.AutoUpgrade && s.ImgVer.HasUpdate {
		s.upgrade(mystManager)
	}

	if s.Config.Enabled {
		s.mod.SetStateContainer(gui.RunnableStateUnknown)
		containerAlreadyRunning, err := mystManager.Start(s.GetConfig())
		if err != nil {
			return
		}

		s.mod.SetStateContainer(gui.RunnableStateRunning)
		s.mod.Update()

		if !containerAlreadyRunning && s.didInstallation {
			s.didInstallation = false

			s.ui.ShowNotificationInstalled()
		}
	}
}
