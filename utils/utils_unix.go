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
	"os"
	"syscall"
)

func getSysProcAttrs() syscall.SysProcAttr {
	return syscall.SysProcAttr{}
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

var productVersion string

func SetProductVersion(ver string) {
	productVersion = ver
}

func GetProductVersion() (string, error) {
	return productVersion, nil
}

// install exe if n/e
func CheckAndInstallExe() error {
	return nil
}

func RunasWithArgsAndWait(cmdArgs string) error {
	panic("not implemented")
}

func EnableAutorun(en bool) error {
    return nil
}

func IsAdmin() bool {
    return os.Getuid() == 0
}