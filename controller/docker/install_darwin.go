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

	"github.com/mysteriumnetwork/myst-launcher/ctrl/util"
	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

func (c *Controller) tryInstallDocker() {
	fmt.Println("TryInstallDocker>")

	mdl := c.model
	ui := c.ui

	mdl.ResetProperties()
	mdl.SwitchState(model.UIStateInstallInProgress)

	executor := util.NewStepExecutor(mdl)
	executor.AddStep("CheckVTx", func() bool {
		featuresOK, err := c.mgr.Features()
		if err != nil {
			c.lg.Println("Failed to query feature:", err)
			return false
		}
		if !featuresOK {
			c.lg.Println("Virtualization is not supported !")
			return false
		}
		return true
	})

	executor.AddStep("CheckDocker", func() bool {
		hasDocker, err := utils.HasDocker()
		if err != nil {
			c.lg.Println("Failed to check Docker:", err)
			return false
		}
		if hasDocker {
			mdl.SwitchState(model.UIStateInstallFinished)
			return true
		}
		return true
	})

	executor.AddStep("DownloadFiles", func() bool {
		name := "Docker.dmg"
		url, err := utils.GetDockerDesktopLink()
		if err != nil {
			c.lg.Println("Couldn't get Docker Desktop link")
			return false
		}
		c.lg.Println("Downloading Docker desktop: ", url)
		err = utils.DownloadFile(utils.GetTmpDir()+name, url, func(progress int) {
			if progress%10 == 0 {
				c.lg.Printf("%s - %d%%\n", name, progress)
			}
		})
		if err != nil {
			c.lg.Println("Couldn't get Docker Desktop")
			return false
		}
		return true
	})

	executor.AddStep("InstallDocker", func() bool {
		var buf bytes.Buffer
		_, err = utils.CmdRun(&buf, "/usr/sbin/diskutil", "unmount", "/Volumes/Docker")
		if err != nil {
			c.lg.Println("Failed to run command:", err)
			return false
		}
		buf.Reset()

		_, err = utils.CmdRun(&buf, "/usr/bin/hdiutil", "attach", utils.GetTmpDir()+name)
		if err != nil {
			c.lg.Println("Failed to run command:", err)
			return false
		}
		buf.Reset()

		// cp -R /Volumes/Docker/Docker.app /Applications
		_, err = utils.CmdRun(&buf, "/bin/cp", "-pR", "/Volumes/Docker/Docker.app", "/Applications")
		if err != nil {
			c.lg.Println("Failed to run command:", err)
			return false
		}
		buf.Reset()

		//  xattr -d -r com.apple.quarantine /Applications/Docker.app
		_, err = utils.CmdRun(&buf, "/usr/bin/xattr", "-d", "-r", "com.apple.quarantine", "/Applications/Docker.app")
		if err != nil {
			c.lg.Println("Failed to run command:", err)
			return false
		}
		buf.Reset()

		_, err = utils.CmdRun(&buf, "/usr/sbin/diskutil", "unmount", "/Volumes/Docker")
		if err != nil {
			c.lg.Println("Failed to run command:", err)
			return false
		}
		buf.Reset()

		// initialize docker desktop
		_, err = utils.CmdRun(&buf, "/usr/bin/open", "/Applications/Docker.app")
		if err != nil {
			c.lg.Println("Failed to run command:", err)
			return false
		}
		buf.Reset()

		return true
	})

	if !executor.Run() {
		mdl.SwitchState(model.UIStateInstallError)
		c.lg.Println("Installation have stopped")
		return false
	}
	mdl.SwitchState(model.UIStateInstallFinished)
	c.lg.Println("Installation succeeded")
	return true	
}
