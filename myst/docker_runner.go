package myst

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/mysteriumnetwork/myst-launcher/utils"
)

type DockerMonitor struct {
	tryStartCount int
	m             *Manager
	couldNotStart bool
}

func NewDockerMonitor(m *Manager) *DockerMonitor {
	return &DockerMonitor{
		tryStartCount: 0,
		m:             m,
	}
}

func (r *DockerMonitor) IsRunning() bool {
	canPingDocker := r.m.CanPingDocker()
	if !canPingDocker {
		r.tryStartCount++

		// try starting docker for 10 times, else try install
		if !r.tryStartDockerDesktop() || r.tryStartCount == 10 {
			r.tryStartCount = 0
			r.couldNotStart = true
		}

		return false // not running or still starting
	}

	return true
}

func (r *DockerMonitor) CouldNotStart() bool {
	val := r.couldNotStart

	if val {
		r.couldNotStart = false
		r.tryStartCount = 0
	}
	return val
}

func (r *DockerMonitor) tryStartDockerDesktop() bool {
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
	fmt.Println("StartDockerDesktop>", runtime.GOOS)
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
		fmt.Println("err>", err)
		return err
	}
	return nil
}
