// +build !windows

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

func Win32Initialize() {}

func LauncherUpgradeAvailable() bool {
	log.Println("LauncherUpgradeAvailable: not implemented")
	return false
}

func UpdateExe() {
	log.Println("UpdateExe: not implemented")
}

func SystemUnderVm() (bool, error) {
	log.Println("SystemUnderVm: not implemented")
	return false, nil
}

func HasDocker() (bool, error) {
	res, err := CmdRun(nil, "/usr/local/bin/docker", "version")
	if err != nil {
		log.Println("HasDocker error:", err)
		return false, err
	}
	return res == 0 || res == 1, nil
}
