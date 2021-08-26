// +build !windows

package utils

import (
	"fmt"
	"syscall"
)

func getSysProcAttrs() syscall.SysProcAttr {
	return syscall.SysProcAttr{}
}

func IsProcessRunning(name string) bool {
	res := CmdRun("pgrep", "-xq", "--", `^`+name)
	fmt.Println("IsProcessRunning>", res)
	return false
}
