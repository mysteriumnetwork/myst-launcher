package utils

import (
	"errors"
	"runtime"
)

func GetDockerDesktopLink() (string, error) {
	switch runtime.GOOS {

	case "windows":
		return "https://desktop.docker.com/mac/stable/amd64/Docker Desktop Installer.exe", nil
		
	case "darwin":
		return "https://desktop.docker.com/mac/stable/amd64/Docker.dmg", nil

	default:
		return "", errors.New("unknown system")
	}
}
