package utils

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
)

func GetDockerDesktopLink() (string, error) {
	os := ""
	switch runtime.GOOS {
	case "windows":
		os = "win"
	case "darwin":
		os = "mac"
	default:
		return "", errors.New("unknown system")
	}

	res := fmt.Sprintf("https://desktop.docker.com/%s/stable/amd64/appcast.xml", os)
	return res, nil
}
