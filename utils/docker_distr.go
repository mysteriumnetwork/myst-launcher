/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package utils

import (
	"errors"
	"fmt"
	"runtime"
)

func GetDockerDesktopLink() (string, error) {
	switch runtime.GOOS {

	case "windows":
		return "https://desktop.docker.com/mac/stable/amd64/Docker Desktop Installer.exe", nil

	case "darwin":
		return fmt.Sprintf("https://desktop.docker.com/mac/stable/%s/Docker.dmg", runtime.GOARCH), nil

	default:
		return "", errors.New("unknown system")
	}
}
