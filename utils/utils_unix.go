//go:build !windows
// +build !windows

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package utils

import (
	"log"
	"syscall"
)

func getSysProcAttrs() syscall.SysProcAttr {
	return syscall.SysProcAttr{}
}

func IsProcessRunning(name string) bool {
	res, err := CmdRun(nil, "pgrep", "-xq", "--", `^`+name)
	if err == nil {
		return res == 0
	}
	log.Println("IsProcessRunning error:", err)
	return false
}

func LauncherUpgradeAvailable() bool {
	return false
}

func UpdateExe() {}

func HasDocker() (bool, error) {
	// Don't try running docker binary directly
	// b/c command.Start() may hang on darwin
	res, err := CmdRun(nil, "/bin/sh", "-c", "/usr/local/bin/docker version")
	if err != nil {
		log.Println("HasDocker error:", err)
		return false, err
	}
	return res == 0 || res == 1, nil
}

func GetProductVersion() (string, error) {
	return "", nil
}

// install exe if n/e
func CheckAndInstallExe() error {
	return nil
}
