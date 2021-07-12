/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32DLL   = windows.NewLazyDLL("user32.dll")

	procCopyFileW      = modkernel32.NewProc("CopyFileW")
	switchToThisWindow = user32DLL.NewProc("SwitchToThisWindow")
)

func cmdRun(name string, args ...string) int {
	log.Print(fmt.Sprintf("Run %v %v \r\n", name, strings.Join(args, " ")))

	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		//HideWindow: true,
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
			//if stderr == "" {
			//	stderr = err.Error()
			//}
		}
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	log.Printf("command exitCode: %v \r\n", exitCode)
	return exitCode
}

func SwitchToThisWindow(hwnd win.HWND, f bool) int32 {
	ret, _, _ := syscall.Syscall(switchToThisWindow.Addr(), 2,
		uintptr(hwnd),
		uintptr(win.BoolToBOOL(f)),
		0,
	)
	return int32(ret)
}

func CreateShortcut(dst, target, args string) error {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer oleShellObject.Release()
	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wshell.Release()

	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", dst)
	if err != nil {
		return err
	}
	idispatch := cs.ToIDispatch()
	oleutil.PutProperty(idispatch, "TargetPath", target)
	oleutil.PutProperty(idispatch, "Arguments", args)
	oleutil.CallMethod(idispatch, "Save")
	return nil
}

func CurrentGroupMembership(group string) bool {
	t := windows.GetCurrentProcessToken()
	sid, _, _, _ := windows.LookupSID("", group)
	is, _ := t.IsMember(sid)
	return is
}

// CopyFile wraps windows function CopyFileW
func CopyFile(src, dst string, failIfExists bool) error {
	lpExistingFileName, err := syscall.UTF16PtrFromString(src)
	if err != nil {
		return err
	}

	lpNewFileName, err := syscall.UTF16PtrFromString(dst)
	if err != nil {
		return err
	}

	var bFailIfExists uint32
	if failIfExists {
		bFailIfExists = 1
	} else {
		bFailIfExists = 0
	}

	r1, _, err := syscall.Syscall(
		procCopyFileW.Addr(),
		3,
		uintptr(unsafe.Pointer(lpExistingFileName)),
		uintptr(unsafe.Pointer(lpNewFileName)),
		uintptr(bFailIfExists))

	if r1 == 0 {
		return err
	}
	return nil
}

func getExeNameFromFullPath(fullExe string) string {
	exe := filepath.Clean(fullExe)
	return exe[len(filepath.Dir(exe))+1:]
}

func checkExe() bool {
	return false
	dst := os.Getenv("ProgramFiles") + "\\MystNodeLauncher"

	fullExe, _ := os.Executable()
	exe := getExeNameFromFullPath(fullExe)
	_, err := os.Stat(dst + "\\" + exe)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func installExe() {
	dst := os.Getenv("ProgramFiles") + "\\MystNodeLauncher"
	os.Mkdir(dst, os.ModePerm)

	fullExe, _ := os.Executable()
	exe := getExeNameFromFullPath(fullExe)
	CopyFile(fullExe, dst+"\\"+exe, false)
}

func CreateAutostartShortcut(args string) {
	dst := os.Getenv("ProgramFiles") + "\\MystNodeLauncher"
	fullExe, _ := os.Executable()
	exe := getExeNameFromFullPath(fullExe)

	shcDst := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Startup", "mysterium node launcher.lnk")
	CreateShortcut(shcDst, dst+"\\"+exe, args)
}

func CheckWindowsVersion() bool {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	releaseIdStr, _, err := k.GetStringValue("ReleaseId")
	if err != nil {
		log.Fatal(err)
	}
	releaseId, _ := strconv.Atoi(releaseIdStr)

	// https://docs.docker.com/docker-for-windows/install/#wsl-2-backend
	v := windows.RtlGetVersion()
	if v.MajorVersion == 10 && releaseId >= 2004 {
		return true
	} else if v.MajorVersion > 10 {
		return true
	}
	return false
}

// unsafe.Sizeof(windows.ProcessEntry32{})
const processEntrySize = 568

func isProcessRunning(name string) bool {
	h, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if e != nil {
		return false
	}
	defer windows.CloseHandle(h)

	p := windows.ProcessEntry32{Size: processEntrySize}
	for {
		e := windows.Process32Next(h, &p)
		if e != nil {
			break
		}
		s := windows.UTF16ToString(p.ExeFile[:])

		if s == name {
			println(s)
			return true
		}
	}
	return false
}

func isWindowsUpdateEnabled() bool {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	disableWUfBSafeguards, _, err := k.GetIntegerValue("DisableWUfBSafeguards")
	if err != nil {
		return false
	}
	return disableWUfBSafeguards == 1
}
