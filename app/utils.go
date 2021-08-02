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
	"time"

	"github.com/mysteriumnetwork/go-fileversion"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"github.com/mysteriumnetwork/myst-launcher/utils"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const launcherLnk = "Mysterium Node Launcher.lnk"

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

func getExeNameFromFullPath(fullExe string) string {
	exe := filepath.Clean(fullExe)
	return exe[len(filepath.Dir(exe))+1:]
}

const launcherExe = "myst-launcher-amd64.exe"

func checkExe() bool {
	dst := os.Getenv("ProgramFiles") + "\\MystNodeLauncher"
	_, err := os.Stat(dst + "\\" + launcherExe)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func LauncherUpgradeAvailable() bool {
	verToStr := func(f fileversion.FileVersion) string {
		return fmt.Sprintf("%.3d.%.3d.%.3d.%.3d", f.Major, f.Minor, f.Patch, f.Build)
	}

	fullExe_, _ := os.Executable()
	ver, err := fileversion.New(fullExe_)
	if err != nil {
		return false
	}

	verDst, err := fileversion.New(MystNodeLauncherExePath())
	if err != nil {
		return false
	}

	return strings.Compare(verToStr(ver.FixedInfo().FileVersion), verToStr(verDst.FixedInfo().FileVersion)) > 0
}

func InstallExe() error {
	fullExe_, _ := os.Executable()
	ver, err := fileversion.New(fullExe_)
	if err != nil {
		return err
	}
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\MysteriumLauncher`, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer k.Close()

	dstPath := os.Getenv("ProgramFiles") + "\\MystNodeLauncher"

	k.SetStringValue("DisplayIcon", fmt.Sprintf(`%s\%s -uninstall`, dstPath, launcherExe))
	k.SetStringValue("DisplayName", ver.ProductName()+" "+ver.ProductVersion())
	k.SetStringValue("DisplayVersion", ver.ProductVersion())
	k.SetStringValue("InstallLocation", dstPath)
	k.SetStringValue("UninstallString", fmt.Sprintf(`"%s\%s" -uninstall`, dstPath, launcherExe))
	k.SetStringValue("Publisher", ver.CompanyName())

	os.Mkdir(dstPath, os.ModePerm)
	srcPath, _ := os.Executable()

	native.CopyFile(srcPath, dstPath+`\`+launcherExe, false)
	return nil
}

func UninstallExe() error {
	utils.StopApp()
	registry.DeleteKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\MysteriumLauncher`)

	shcDst := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Startup", launcherLnk)
	_ = os.Remove(shcDst)

	shcDst = path.Join(os.Getenv("USERPROFILE"), "Desktop", launcherLnk)
	_ = os.Remove(shcDst)

	dir := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Mysterium Network")
	os.Mkdir(dir, os.ModePerm)
	shcDst = path.Join(dir, launcherLnk)
	_ = os.Remove(shcDst)

	return nil
}

func MystNodeLauncherExePath() string {
	return os.Getenv("ProgramFiles") + "\\MystNodeLauncher" + "\\" + launcherExe
}

func CreateAutostartShortcut(args string) {
	shcDst := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Startup", launcherLnk)
	CreateShortcut(shcDst, MystNodeLauncherExePath(), args)
}

func CreateDesktopShortcut(args string) {
	shcDst := path.Join(os.Getenv("USERPROFILE"), "Desktop", launcherLnk)
	CreateShortcut(shcDst, MystNodeLauncherExePath(), args)
}

func CreateStartMenuShortcut(args string) {
	dir := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Mysterium Network")
	os.Mkdir(dir, os.ModePerm)
	shcDst := path.Join(dir, launcherLnk)
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

func isUnderVm() (bool, error) {
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
	model := ""
	if count > 0 {
		itemRaw, _ := oleutil.CallMethod(result, "ItemIndex", 0)
		item := itemRaw.ToIDispatch()
		defer item.Release()

		variantModel, err := oleutil.GetProperty(item, "Model")
		if err != nil {
			return false, err
		}
		model = variantModel.ToString()
	}
	vmTest := []string{"virtual", "vmware", "kvm", "xen"}
	isVM := false
	for _, v := range vmTest {
		if strings.Contains(strings.ToLower(model), v) {
			isVM = true
			break
		}
	}
	return isVM, nil
}

// We can not use the IsProcessorFeaturePresent approach, as it does not matter in self-virtualized environment
// see https://devblogs.microsoft.com/oldnewthing/20201216-00/?p=104550
func hasVTx() bool {
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

const (
	FeatureWSL    = "Microsoft-Windows-Subsystem-Linux"
	FeatureHyperV = "Microsoft-Hyper-V"
)

// Returns: featureExists, featureEnabled, error
func QueryWindowsFeature(feature string) (bool, bool, error) {
	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return false, false, err
	}
	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return false, false, err
	}
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", nil, "root\\cimv2")
	if err != nil {
		return false, false, err
	}
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	// result is a SWBemObjectSet
	resultRaw, err := oleutil.CallMethod(service, "ExecQuery", fmt.Sprintf("SELECT * FROM Win32_OptionalFeature Where Name='%s'", feature))
	if err != nil {
		return false, false, err
	}
	result := resultRaw.ToIDispatch()
	defer result.Release()
	countVar, err := oleutil.GetProperty(result, "Count")
	if err != nil {
		return false, false, err
	}
	count := int(countVar.Val)
	featureExists := count > 0

	resultRaw, err = oleutil.CallMethod(service, "ExecQuery", fmt.Sprintf("SELECT * FROM Win32_OptionalFeature Where Name='%s' and InstallState=1", feature))
	if err != nil {
		return false, false, err
	}
	result = resultRaw.ToIDispatch()
	defer result.Release()

	countVar, err = oleutil.GetProperty(result, "Count")
	if err != nil {
		return false, false, err
	}
	count = int(countVar.Val)
	featureEnabled := count > 0

	return featureExists, featureEnabled, nil
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
