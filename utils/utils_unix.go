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
