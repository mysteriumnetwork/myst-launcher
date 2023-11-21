/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package docker

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type DockerRunner struct {
	tryStartCount int
	myst          model.DockerManager
}

func NewDockerRunner(myst model.DockerManager) *DockerRunner {
	return &DockerRunner{
		tryStartCount: 0,
		myst:          myst,
	}
}

func (r *DockerRunner) IsRunning() bool {
	canPingDocker := r.myst.PingDocker()
	// defer log.Println("IsRunning >", canPingDocker)

	if canPingDocker {
		r.tryStartCount = 0
	}
	return canPingDocker
}

// return values: isRunning, couldNotStart
func (r *DockerRunner) IsRunningOrTryStart() (bool, bool) {
	// fmt.Println("IsRunningOrTryStart >")
	// defer fmt.Println("IsRunningOrTryStart >>>")

	if !r.myst.PingDocker() {
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

func getProcessName() string {
	exe := "Docker Desktop.exe"
	if runtime.GOOS == "darwin" {
		exe = "Docker"
	}
	return exe
}

func (r *DockerRunner) tryStartDockerDesktop() bool {
	exe := getProcessName()

	if utils.IsProcessRunning(exe) {
		return true
	}
	// fmt.Println("Start Docker Desktop 1>>>")
	err := startDockerDesktop()
	if err != nil {
		fmt.Println("Failed to start cmd:", err)
		return false
	}
	// fmt.Println("Start Docker Desktop 2>")
	return true
}

func startDockerDesktop() error {
	var cmd *exec.Cmd
	fmt.Println("Start Docker Desktop>", runtime.GOOS)

	switch runtime.GOOS {
	case "windows":
		dd := os.Getenv("ProgramFiles") + "\\Docker\\Docker\\Docker Desktop.exe"
		cmd = exec.Command(dd, "-Autostart")
		// fmt.Println("Start Docker Desktop>", cmd)
	case "darwin":
		cmd = exec.Command("open", "/Applications/Docker.app/")
	default:
		return errors.New("unsupported OS: " + runtime.GOOS)
	}
	fmt.Println("Start Docker Desktop>", cmd)

	if err := cmd.Start(); err != nil {
		fmt.Println("err>", err)
		return err
	}
	return nil
}
