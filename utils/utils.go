/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package utils

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"syscall"
	"time"
)

var a = getSysProcAttrs()

// returns: exit status, error
func CmdRun(out *bytes.Buffer, name string, args ...string) (int, error) {
	log.Print(fmt.Sprintf("Run %v %v \r\n", name, strings.Join(args, " ")))

	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &a
	if out != nil {
		cmd.Stdout = out
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	if err := cmd.Wait(); err != nil {

		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			if waitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return waitStatus.ExitStatus(), nil
			} else {
				return 0, errors.New("Type assertion failed: syscall.WaitStatus")
			}
		} else {
			log.Printf("error> %+v, \r\n", err)
			return 0, errors.New("Type assertion failed: *exec.ExitError")
		}
	} else {
		if waitStatus, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
			return waitStatus.ExitStatus(), nil
		} else {
			return 0, errors.New("Type assertion failed: syscall.WaitStatus")
		}
	}
}

func Retry(attempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return Retry(attempts, sleep, fn)
		}
		return err
	}
	return nil
}

func PanicHandler(threadName string) {
	if err := recover(); err != nil {

		fmt.Printf("Panic: %s\n", err)
		fmt.Printf("Stacktrace %s: %s\n", threadName, debug.Stack())
		fname := fmt.Sprintf("%s/launcher_trace_%d.txt", GetUserProfileDir(), time.Now().Unix())
		f, err := os.Create(fname)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()

		var bu bytes.Buffer

		v, _ := GetProductVersion()
		bu.WriteString(fmt.Sprintf("Version %s: \n", v))
		bu.WriteString(fmt.Sprintf("Stacktrace %s: \n", threadName))
		bu.Write(debug.Stack())
		f.Write(bu.Bytes())

		//OpenExceptionDlg(threadName, bu.String())
		//os.Exit(1)
	}
}
