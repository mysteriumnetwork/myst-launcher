/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package app

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/mysteriumnetwork/myst-launcher/myst"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type DockerRunner struct {
	tryStartCount int
	m             *myst.Manager
}

func NewDockerMonitor(m *myst.Manager) *DockerRunner {
	return &DockerRunner{
		tryStartCount: 0,
		m:             m,
	}
}

func (r *DockerRunner) IsRunningShort() bool {
	canPingDocker := r.m.CanPingDocker()
	if canPingDocker {
		r.tryStartCount = 0
	}
	return canPingDocker
}

// return values: isRunning, couldNotStart
func (r *DockerRunner) IsRunning() (bool, bool) {
	canPingDocker := r.m.CanPingDocker()

	if !canPingDocker {
		r.tryStartCount++

		if !r.tryStartDockerDesktop() || r.tryStartCount >= 20 {
			r.tryStartCount = 0
			return false, true
		}
		return false, false
	}
	r.tryStartCount = 0
	return true, false
}

func (r *DockerRunner) tryStartDockerDesktop() bool {
	exe := "Docker Desktop.exe"
	if runtime.GOOS == "darwin" {
		exe = "Docker"
	}

	if utils.IsProcessRunning(exe) {
		return true
	}
	if err := StartDockerDesktop(); err != nil {
		log.Printf("Failed to start cmd: %v", err)
		return false
	}
	return true
}

func StartDockerDesktop() error {
	var cmd *exec.Cmd
	log.Println("StartDockerDesktop >")
	switch runtime.GOOS {
	case "windows":
		dd := os.Getenv("ProgramFiles") + "\\Docker\\Docker\\Docker Desktop.exe"
		cmd = exec.Command(dd, "-Autostart")
	case "darwin":
		cmd = exec.Command("open", "/Applications/Docker.app/")
	default:
		return errors.New("unsupported OS: " + runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		log.Println("err>", err)
		return err
	}
	return nil
}
