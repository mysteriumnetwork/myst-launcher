package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func runProc(lv *LogView, name string, args []string) int {
	lv.PostAppendText(fmt.Sprintf("Run %v %v \r\n", name, args))

	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		//CreationFlags: 0x08000000, //CREATE_NO_WINDOW
	}

	//output, _ := cmd.CombinedOutput()
	//fmt.Println(string(output))
	stdout, err := cmd.StdoutPipe()
	//cmd.Stderr = cmd.Stdout
	if err != nil {
		log.Println(err)
		return -1
	}
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

		//fmt.Println(string(line))
		//lv.PostAppendText(string(line)+"\r\n")
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

			//if stderr == "" {
			//	stderr = err.Error()
			//}
		}
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	//log.Printf("command exitCode: %v \r\n", exitCode)
	return exitCode
}

func env() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		j := strings.Index(e, "=")
		if j == 0 {
			continue
		}
		name := e[0:j]
		value := strings.Replace(e[j+1:], ";", "\r\n", -1)

		//fmt.Println("nv", name, value)
		env[name] = value
	}
	return env
}

func runMeElevated(exe string, args string, cwd string) error {
	verb := "runas"
	//cwd, _ := os.Getwd()
	//args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	return err
}
