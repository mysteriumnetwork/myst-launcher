/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
package app

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

	"github.com/mysteriumnetwork/go-fileversion"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	modkernel32                   = syscall.NewLazyDLL("kernel32.dll")
	procCopyFileW                 = modkernel32.NewProc("CopyFileW")
	procIsProcessorFeaturePresent = modkernel32.NewProc("IsProcessorFeaturePresent")
)

func cmdRun(name string, args ...string) int {
	log.Print(fmt.Sprintf("Run %v %v \r\n", name, strings.Join(args, " ")))

	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

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
	dst := os.Getenv("ProgramFiles") + "\\MystNodeLauncher"

	fullExe, _ := os.Executable()
	exe := getExeNameFromFullPath(fullExe)
	_, err := os.Stat(dst + "\\" + exe)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func InstallExe() error {
	fullExe_, _ := os.Executable()
	f, err := fileversion.New(fullExe_)
	if err != nil {
		return err
	}
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\MysteriumLauncher`, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer k.Close()

	dstPath := os.Getenv("ProgramFiles") + "\\MystNodeLauncher"
	exeName := "\\myst-launcher-amd64.exe"

	k.SetStringValue("DisplayIcon", dstPath+exeName)
	k.SetStringValue("DisplayName", f.ProductName()+" "+f.ProductVersion())
	k.SetStringValue("DisplayVersion", f.ProductVersion())
	k.SetStringValue("InstallLocation", dstPath)
	k.SetStringValue("UninstallString", fmt.Sprintf(`"%s%s" -uninstall`, dstPath, exeName))
	k.SetStringValue("Publisher", f.CompanyName())

	os.Mkdir(dstPath, os.ModePerm)
	fullExe, _ := os.Executable()
	exe := getExeNameFromFullPath(fullExe)
	CopyFile(fullExe, dstPath+`\`+exe, false)
	return nil
}

func UninstallExe() error {
	registry.DeleteKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\MysteriumLauncher`)

	shcDst := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Startup", "mysterium node launcher.lnk")
	_ = os.Remove(shcDst)

	shcDst = path.Join(os.Getenv("USERPROFILE"), "Desktop", "mysterium node launcher.lnk")
	_ = os.Remove(shcDst)

	dir := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Mysterium Network")
	os.Mkdir(dir, os.ModePerm)
	shcDst = path.Join(dir, "mysterium node launcher.lnk")
	_ = os.Remove(shcDst)
	return nil
}

func MystNodeLauncherExePath() string {
	dst := os.Getenv("ProgramFiles") + "\\MystNodeLauncher"
	fullExe, _ := os.Executable()
	exe := getExeNameFromFullPath(fullExe)
	return dst + "\\" + exe
}

func CreateAutostartShortcut(args string) {
	shcDst := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Startup", "Mysterium node launcher.lnk")
	CreateShortcut(shcDst, MystNodeLauncherExePath(), args)
}

func CreateDesktopShortcut(args string) {
	shcDst := path.Join(os.Getenv("USERPROFILE"), "Desktop", "Mysterium node launcher.lnk")
	CreateShortcut(shcDst, MystNodeLauncherExePath(), args)
}

func CreateStartMenuShortcut(args string) {
	dir := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Mysterium Network")
	os.Mkdir(dir, os.ModePerm)
	shcDst := path.Join(dir, "Mysterium node launcher.lnk")
	CreateShortcut(shcDst, MystNodeLauncherExePath(), args)
}

func IsWindowsVersionCompatible() bool {
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

// does not matter in self-virtualized environment
// see https://devblogs.microsoft.com/oldnewthing/20201216-00/?p=104550
func hasVTx_() bool {
	const PF_VIRT_FIRMWARE_ENABLED = 21
	r1, r2, e1 := syscall.Syscall(procIsProcessorFeaturePresent.Addr(), 1, uintptr(PF_VIRT_FIRMWARE_ENABLED), 0, 0)
	if e1 != 0 {
		fmt.Printf("Err: %s \n", syscall.Errno(e1))
	}
	log.Println("hasVTx", r1, r2)
	return r1 != 0
}

func hasVTx() bool {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, _ := oleutil.CreateObject("WbemScripting.SWbemLocator")
	defer unknown.Release()

	wmi, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, _ := oleutil.CallMethod(wmi, "ConnectServer", nil, "root\\cimv2")
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	// result is a SWBemObjectSet
	resultRaw, _ := oleutil.CallMethod(service, "ExecQuery", "SELECT * FROM Win32_ComputerSystem")
	result := resultRaw.ToIDispatch()
	defer result.Release()

	countVar, _ := oleutil.GetProperty(result, "Count")
	count := int(countVar.Val)

	for i := 0; i < count; i++ {
		itemRaw, _ := oleutil.CallMethod(result, "ItemIndex", i)
		item := itemRaw.ToIDispatch()
		defer item.Release()

		variantHypervisorPresent, err := oleutil.GetProperty(item, "HypervisorPresent")
		if err == nil {
			return variantHypervisorPresent.Value().(bool)
		}
	}
	return false
}

func isWLSEnabled() bool {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, _ := oleutil.CreateObject("WbemScripting.SWbemLocator")
	defer unknown.Release()

	wmi, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, _ := oleutil.CallMethod(wmi, "ConnectServer", nil, "root\\cimv2")
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	// result is a SWBemObjectSet
	resultRaw, _ := oleutil.CallMethod(service, "ExecQuery", "SELECT * FROM Win32_OptionalFeature Where Name='Microsoft-Windows-Subsystem-Linux' and InstallState=1")
	result := resultRaw.ToIDispatch()
	defer result.Release()

	countVar, _ := oleutil.GetProperty(result, "Count")
	count := int(countVar.Val)
	return count > 0
}
