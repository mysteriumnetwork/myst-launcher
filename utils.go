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
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	modkernel32   = syscall.NewLazyDLL("kernel32.dll")
	procCopyFileW = modkernel32.NewProc("CopyFileW")

	user32DLL          = windows.NewLazyDLL("user32.dll")
	switchToThisWindow = user32DLL.NewProc("SwitchToThisWindow")
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

func SwitchToThisWindow(hwnd win.HWND, f bool) int32 {
	ret, _, e := syscall.Syscall(switchToThisWindow.Addr(), 2,
		uintptr(hwnd),
		uintptr(win.BoolToBOOL(f)),
		0,
	)
	fmt.Println(e)
	return int32(ret)
}

func CreateShortcut(dst, target string) error {
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

func getExeName() (string, string) {
	fullExe, _ := os.Executable()
	exe := filepath.Clean(fullExe)
	fmt.Println(exe, filepath.Dir(exe))
	exe = exe[len(filepath.Dir(exe))+1:]
	fmt.Println(exe)

	return fullExe, exe
}

func checkExe() bool {
	dst := os.Getenv("ProgramFiles") + "\\MystNodeLauncher"

	_, exe := getExeName()
	_, err := os.Stat(dst + "\\" + exe)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func installExe() {
	dst := os.Getenv("ProgramFiles") + "\\MystNodeLauncher"
	os.Mkdir(dst, os.ModePerm)

	fullExe, exe := getExeName()
	CopyFile(fullExe, dst+"\\"+exe, false)

	shcDst := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Startup", "mysterium node launcher.lnk")
	CreateShortcut(shcDst, dst+"\\"+exe)
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

	v := windows.RtlGetVersion()
	fmt.Println(v.MajorVersion)
	if v.MajorVersion == 10 && releaseId >= 1906 {
		return true
	} else if v.MajorVersion > 10 {
		return true
	}
	return false
}
