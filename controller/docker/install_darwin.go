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

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

// returns exit mode: true means exit
func (c *Controller) tryInstallDocker() bool {
	mdl := c.a.GetModel()
	ui := c.a.GetUI()

	mdl.SwitchState(model.UIStateInstallNeeded)
	if !ui.WaitDialogueComplete() {
		return true
	}
	mdl.SwitchState(model.UIStateInstallInProgress)

	mdl.UpdateProperties(model.UIProps{"CheckVTx": model.StepInProgress})
	featuresOK, err := c.mgr.Features()
	if err != nil {
		c.lg.Println("Failed to query feature:", err)
		mdl.SwitchState(model.UIStateInstallError)
		mdl.UpdateProperties(model.UIProps{"CheckVTx": model.StepFailed})
		return true
	}
	if !featuresOK {
		c.lg.Println("Virtualization is not supported !")
		mdl.SwitchState(model.UIStateInstallError)
		return true
	}
	mdl.UpdateProperties(model.UIProps{"CheckVTx": model.StepFinished})

	mdl.UpdateProperties(model.UIProps{"CheckDocker": model.StepInProgress})
	hasDocker, err := utils.HasDocker()
	if err != nil {
		c.lg.Println("Failed to check Docker:", err)
		mdl.SwitchState(model.UIStateInstallError)
		mdl.UpdateProperties(model.UIProps{"CheckDocker": model.StepFailed})
		return true
	}
	mdl.UpdateProperties(model.UIProps{"CheckDocker": model.StepFinished})

	if hasDocker {
		mdl.SwitchState(model.UIStateInstallFinished)
		ok := ui.WaitDialogueComplete()
		if !ok {
			return true
		}
		mdl.SwitchState(model.UIStateInitial)
		return false
	}

	mdl.UpdateProperties(model.UIProps{"DownloadFiles": model.StepInProgress})
	name := "Docker.dmg"
	url, err := utils.GetDockerDesktopLink()
	if err != nil {
		c.lg.Println("Couldn't get Docker Desktop link")
		mdl.SwitchState(model.UIStateInstallError)
		return true
	}
	c.lg.Println("Downloading Docker desktop: ", url)
	err = utils.DownloadFile(utils.GetTmpDir()+name, url, func(progress int) {
		if progress%10 == 0 {
			c.lg.Println(fmt.Sprintf("%s - %d%%", name, progress))
		}
	})
	if err != nil {
		c.lg.Println("Couldn't get Docker Desktop")
		mdl.SwitchState(model.UIStateInstallError)
		mdl.UpdateProperties(model.UIProps{"DownloadFiles": model.StepFailed})
		return true
	}
	mdl.UpdateProperties(model.UIProps{"DownloadFiles": model.StepFinished})

	mdl.UpdateProperties(model.UIProps{"InstallDocker": model.StepInProgress})
	var buf bytes.Buffer
	_, err = utils.CmdRun(&buf, "/usr/sbin/diskutil", "unmount", "/Volumes/Docker")
	if err != nil {
		c.lg.Println("Failed to run command:", err)
		mdl.SwitchState(model.UIStateInstallError)
		mdl.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	_, err = utils.CmdRun(&buf, "/usr/bin/hdiutil", "attach", utils.GetTmpDir()+name)
	if err != nil {
		c.lg.Println("Failed to run command:", err)
		mdl.SwitchState(model.UIStateInstallError)
		mdl.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	// cp -R /Volumes/Docker/Docker.app /Applications
	_, err = utils.CmdRun(&buf, "/bin/cp", "-pR", "/Volumes/Docker/Docker.app", "/Applications")
	if err != nil {
		c.lg.Println("Failed to run command:", err)
		mdl.SwitchState(model.UIStateInstallError)
		mdl.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	//  xattr -d -r com.apple.quarantine /Applications/Docker.app
	_, err = utils.CmdRun(&buf, "/usr/bin/xattr", "-d", "-r", "com.apple.quarantine", "/Applications/Docker.app")
	if err != nil {
		c.lg.Println("Failed to run command:", err)
		mdl.SwitchState(model.UIStateInstallError)
		mdl.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	_, err = utils.CmdRun(&buf, "/usr/sbin/diskutil", "unmount", "/Volumes/Docker")
	if err != nil {
		c.lg.Println("Failed to run command:", err)
		mdl.SwitchState(model.UIStateInstallError)
		mdl.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	// initialize docker desktop
	_, err = utils.CmdRun(&buf, "/usr/bin/open", "/Applications/Docker.app")
	if err != nil {
		c.lg.Println("Failed to run command:", err)
		mdl.SwitchState(model.UIStateInstallError)
		mdl.UpdateProperties(model.UIProps{"InstallDocker": model.StepFailed})
		return true
	}
	buf.Reset()

	mdl.UpdateProperties(model.UIProps{"InstallDocker": model.StepFinished})
	c.lg.Println("Installation succeeded")

	mdl.SwitchState(model.UIStateInstallFinished)
	if !ui.WaitDialogueComplete() {
		return true
	}
	mdl.SwitchState(model.UIStateInitial)
	return false
}
