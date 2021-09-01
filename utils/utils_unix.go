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
	res, err := CmdRun("pgrep", "-xq", "--", `^`+name)
	if err == nil {
		return res == 0
	}
	log.Println("CmdRun error:", err)
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

// query if there are features to be enabled
func QueryFeatures() ([]int, error) {
	f := make([]int, 0)
	return f, nil
}

func InstallFeatures(features []int, onFeatureReady func(int, string)) error {
	return nil
}

func SystemUnderVm() (bool, error) {
	log.Println("SystemUnderVm: not implemented")
	return false, nil
}
