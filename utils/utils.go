package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var a = getSysProcAttrs()

func CmdRun(name string, args ...string) int {
	log.Print(fmt.Sprintf("Run %v %v \r\n", name, strings.Join(args, " ")))

	cmd := exec.Command(name, args...)
	//cmd.SysProcAttr = &a

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return -1
	}
	defer stdout.Close()

	if err = cmd.Start(); err != nil {
		log.Println(err)
		return -1
	}
	r := bufio.NewReader(stdout)
	for {
		var line []byte
		line, _, e := r.ReadLine()
		if e == io.EOF {
			break
		}
		_ = line

		log.Println(string(line))
	}

	var exitCode int
	const defaultFailedCode = 1

	if err := cmd.Wait(); err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			log.Printf("Could not get exit code for failed program: %v, %v \r\n", name, args)
			exitCode = defaultFailedCode

			log.Printf(">>>> %+v, \r\n", err)
		}
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	log.Printf("command exitCode: %v \r\n", exitCode)
	return exitCode
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
