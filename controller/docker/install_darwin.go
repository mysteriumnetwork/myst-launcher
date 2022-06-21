//go:build darwin
// +build darwin

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package docker

import (
	"bytes"
	"fmt"
	"log"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

// returns exit mode: true means exit
func (s *Controller) tryInstallDocker() bool {
	mdl := s.a.GetModel()
	ui := s.a.GetUI()

	model.SwitchState(model.UIStateInstallNeeded)
	if !ui.WaitDialogueComplete() {
		return true
	}
	model.SwitchState(model.UIStateInstallInProgress)

	model.UpdateProperties(model.UIProps{"CheckVTx": model.StepInProgress})
	featuresOK, err := s.mgr.Features()
	if err != nil {
		log.Println("Failed to query feature:", err)
		model.SwitchState(model.UIStateInstallError)
		model.UpdateProperties(model.UIProps{"CheckVTx": model.StepFailed})
		return true
	}
	if !featuresOK {
		log.Println("Virtualization is not supported !")
		model.SwitchState(model.UIStateInstallError)
		return true
	}
	model.UpdateProperties(model.UIProps{"CheckVTx": model.StepFinished})

	model.UpdateProperties(model.UIProps{"CheckDocker": model.StepInProgress})
	hasDocker, err := utils.HasDocker()
	if err != nil {
		log.Println("Failed to check Docker:", err)
		model.SwitchState(model.UIStateInstallError)
		model.UpdateProperties(model.UIProps{"CheckDocker": model.StepFailed})
		return true
	}
	model.UpdateProperties(model.UIProps{"CheckDocker": model.StepFinished})

	if hasDocker {
		model.SwitchState(model.UIStateInstallFinished)
		ok := ui.WaitDialogueComplete()
		if !ok {
			return true
		}
		model.SwitchState(model.UIStateInitial)
		return false
	}

	model.UpdateProperties(model.UIProps{"DownloadFiles": model.StepInProgress})
	name := "Docker.dmg"
	url, err := utils.GetDockerDesktopLink()
	if err != nil {
		log.Println("Couldn't get Docker Desktop link")
		model.SwitchState(model.UIStateInstallError)
		return true
	}
	log.Println("Downloading Docker desktop: ", url)
	err = utils.DownloadFile(utils.GetTmpDir()+name, url, func(progress int) {
		if progress%10 == 0 {
			log.Println(fmt.Sprintf("%s - %d%%", name, progress))
		}
	})
	if err != nil {
		log.Println("Couldn't get Docker Desktop")
		model.SwitchState(model.UIStateInstallError)
		model.UpdateProperties(model.UIProps{"DownloadFiles": model.StepFailed})
		return true
	}
	model.UpdateProperties(model.UIProps{"DownloadFiles": model.StepFinished})

	model.UpdateProperties(model.UIProps{"InstallDocker": model.StepInProgress})
	var buf bytes.Buffer
	_, err = utils.CmdRun(&buf, "/usr/sbin/diskutil", "unmount", "/Volumes/Docker")
	if err != nil {
		log.Println("Failed to run command:", err)
		model.SwitchState(model.UIStateInstallError)
		model.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	_, err = utils.CmdRun(&buf, "/usr/bin/hdiutil", "attach", utils.GetTmpDir()+name)
	if err != nil {
		log.Println("Failed to run command:", err)
		model.SwitchState(model.UIStateInstallError)
		model.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	// cp -R /Volumes/Docker/Docker.app /Applications
	_, err = utils.CmdRun(&buf, "/bin/cp", "-pR", "/Volumes/Docker/Docker.app", "/Applications")
	if err != nil {
		log.Println("Failed to run command:", err)
		model.SwitchState(model.UIStateInstallError)
		model.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	//  xattr -d -r com.apple.quarantine /Applications/Docker.app
	_, err = utils.CmdRun(&buf, "/usr/bin/xattr", "-d", "-r", "com.apple.quarantine", "/Applications/Docker.app")
	if err != nil {
		log.Println("Failed to run command:", err)
		model.SwitchState(model.UIStateInstallError)
		model.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	_, err = utils.CmdRun(&buf, "/usr/sbin/diskutil", "unmount", "/Volumes/Docker")
	if err != nil {
		log.Println("Failed to run command:", err)
		model.SwitchState(model.UIStateInstallError)
		model.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	// initialize docker desktop
	_, err = utils.CmdRun(&buf, "/usr/bin/open", "/Applications/Docker.app")
	if err != nil {
		log.Println("Failed to run command:", err)
		model.SwitchState(model.UIStateInstallError)
		model.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	model.UpdateProperties(model.UIProps{"InstallDocker": model.StepFinished})
	log.Println("Installation succeeded")

	model.SwitchState(model.UIStateInstallFinished)
	if !ui.WaitDialogueComplete() {
		return true
	}
	model.SwitchState(model.UIStateInitial)
	return false
}
