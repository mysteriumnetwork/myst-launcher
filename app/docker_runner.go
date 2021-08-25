package app

import (
	"log"
	"os"
	"os/exec"

	"github.com/mysteriumnetwork/myst-launcher/gui"
	"github.com/mysteriumnetwork/myst-launcher/myst"
)

type DockerMonitor struct {
	tryStartCount int
	m             *myst.Manager
	couldNotStart bool
}

func NewDockerMonitor(m *myst.Manager) *DockerMonitor {
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
	gui.UI.SetStateContainer(gui.RunnableStateUnknown)
	gui.UI.SetStateDocker(gui.RunnableStateStarting)

	if IsProcessRunning("Docker Desktop.exe") {
		return true
	}
	if err := startDocker(); err != nil {
		log.Printf("Failed to start cmd: %v", err)
		return false
	}
	return true
}

func startDocker() error {
	dd := os.Getenv("ProgramFiles") + "\\Docker\\Docker\\Docker Desktop.exe"
	cmd := exec.Command(dd, "-Autostart")
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
