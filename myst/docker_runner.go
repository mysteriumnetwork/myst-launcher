package myst

import (
	"errors"
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

// return values: isRunning, couldNotStart
func (r *DockerMonitor) IsRunning() (bool, bool) {
	canPingDocker := r.m.CanPingDocker()
	if !canPingDocker {
		r.tryStartCount++

		if !r.tryStartDockerDesktop() || r.tryStartCount == 100 {
			r.tryStartCount = 0
			return false, true
		}
		return false, false
	}
	return true, false
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
