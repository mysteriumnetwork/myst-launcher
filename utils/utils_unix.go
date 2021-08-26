// +build !windows

package utils

import (
	"os/exec"
)

func IsProcessRunning(name string) bool {
	cmd := exec.Command("pgrep", "-xq", "--", `"^Docker"`)
	_, err := cmd.Output()
	if werr, ok := err.(*exec.ExitError); ok {
		if s := werr.Error(); s != "0" {
			return true
		}
	}
	return false
}
