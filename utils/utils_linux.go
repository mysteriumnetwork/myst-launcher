//go:build linux
// +build linux

/*
 * copyright
 */

package utils

/*
#include <unistd.h>
#include <signal.h>
#include <stdio.h>
*/
import "C"

import (
	"errors"
	"os"

	"github.com/shirou/gopsutil/v3/process"
)

// kill process by name except current process
func KillProcessByName(exeName string) error {
	pid := os.Getpid()
	processes, err := process.Processes()
	if err != nil {
		return err
	}

	for _, p := range processes {
		n, _ := p.Name()
		if n == exeName && int(p.Pid) != pid {
			if err := TerminateProcess(uint32(p.Pid), 0); err != nil {
				return err
			}
		}
	}
	return nil
}

func TerminateProcess(pid uint32, exitcode int) error {

	p, err := process.NewProcess(int32(pid))
	err = p.Kill()
	if err != nil {
		return errors.New("Process not found")
	}
	return nil
}

func IsProcessRunning(name string) bool {
	processes, err := process.Processes()
	if err != nil {
		return false
	}
	for _, p := range processes {
		n, _ := p.Name()
		if n == name {
			return true
		}
	}
	return false
}

func IsProcessRunningExt(exeName, fullpath string) (uint32, error) {
	processes, err := process.Processes()
	if err != nil {
		return 0, err
	}
	for _, p := range processes {
		n, _ := p.Name()
		exepath, _ := p.Exe()
		if n == exeName && exepath == fullpath {
			return uint32(p.Pid), nil
		}
	}
	return 0, nil
}
