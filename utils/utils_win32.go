//go:build windows
// +build windows

/**
 * Copyright (c) 2021 BlockDev AG
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package utils

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/blang/semver/v4"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/gonutz/w32"
	"github.com/lxn/walk"
	"github.com/mysteriumnetwork/go-fileversion"
	"github.com/pkg/errors"
	"github.com/winlabs/gowin32"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
	"github.com/mysteriumnetwork/myst-launcher/native"
)

const launcherLnk = "Mysterium Node Launcher.lnk"

var (
	errProductNotFound = errors.New("Package not found")
)

func getSysProcAttrs() syscall.SysProcAttr {
	return syscall.SysProcAttr{
		HideWindow: true,
	}
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

func UpdateExe() {
	exePath, _ := os.Executable()
	err := native.ShellExecuteAndWait(0, "runas", exePath, _const.FlagInstall, "", syscall.SW_NORMAL)
	if err != nil {
		log.Println("Failed to install exe", err)
	}
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

	return strings.Compare(verToStr(ver.FixedInfo().ProductVersion), verToStr(verDst.FixedInfo().ProductVersion)) > 0
}

// install exe if n/e
func CheckAndInstallExe() error {
	if !checkExe() {
		fullExe, _ := os.Executable()
		cmdArgs := _const.FlagInstall
		err := native.ShellExecuteAndWait(0, "runas", fullExe, cmdArgs, "", syscall.SW_NORMAL)
		if err != nil {
			return err
		}
	}
	return nil
}

func RunasWithArgsAndWait(cmdArgs string) error {
	fullExe, _ := os.Executable()
	err := native.ShellExecuteAndWait(0, "runas", fullExe, cmdArgs, "", syscall.SW_NORMAL)
	return err
}

func RunasWithArgsNoWait(cmdArgs string) error {
	fullExe, _ := os.Executable()
	err := native.ShellExecuteNowait(0, "runas", fullExe, cmdArgs, "", syscall.SW_NORMAL)
	return err
}

func RunWithArgsNoWait(cmdArgs string) error {
	fullExe, _ := os.Executable()
	err := native.ShellExecuteNowait(0, "", fullExe, cmdArgs, "", syscall.SW_NORMAL)
	return err
}

// should be executed with admin's privileges
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

// should be executed with admin privs
func UninstallExe() error {
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
	_ = releaseId

	// https://docs.docker.com/docker-for-windows/install/#wsl-2-backend
	//releaseId >= 2004 - home & professional
	//releaseId >= 1909 - Enterprise or Education
	v := windows.RtlGetVersion()
	if v.MajorVersion >= 10 {
		return true
	}
	return false
}

const processEntrySize = uint32(unsafe.Sizeof(windows.ProcessEntry32{}))

func IsProcessRunning(name string) bool {
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
			log.Println("IsProcessRunning >", s)

			return true
		}
	}
	return false
}

func getExePath(id uint32) (string, error) {
	h, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE, id)
	if e != nil {
		return "", e
	}
	defer windows.CloseHandle(h)

	p := windows.ModuleEntry32{Size: uint32(windows.SizeofModuleEntry32)}
	for {
		e := windows.Module32Next(h, &p)
		if e != nil {
			return "", e
		}

		s := windows.UTF16ToString(p.ExePath[:])
		if len(s) > 4 && s[len(s)-3:] == "exe" {
			return s, nil
		}
	}
}

func TerminateProcess(pid uint32, exitcode int) error {
	h, e := windows.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
	if e != nil {
		return e
	}
	defer windows.CloseHandle(h)
	e = windows.TerminateProcess(h, uint32(exitcode))
	return e
}

func FindProcess(exeName, fullpath string) (uint32, error) {
	h, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if e != nil {
		return 0, e
	}
	defer windows.CloseHandle(h)

	p := windows.ProcessEntry32{Size: processEntrySize}
	for {
		if e := windows.Process32Next(h, &p); e != nil {
			return 0, e
		}

		if s := windows.UTF16ToString(p.ExeFile[:]); s == exeName {
			return p.ProcessID, e
		}
	}
	// return 0, nil
}

func IsProcessRunningExt(exeName, fullpath string) (uint32, error) {
	h, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if e != nil {
		return 0, e
	}
	defer windows.CloseHandle(h)

	p := windows.ProcessEntry32{Size: processEntrySize}
	for {
		if e := windows.Process32Next(h, &p); e != nil {
			return 0, e
		}

		if s := windows.UTF16ToString(p.ExeFile[:]); s == exeName {
			pp, e := getExePath(p.ProcessID)
			if e != nil {
				return 0, e
			}

			if pp == fullpath {
				return uint32(p.ProcessID), nil
			}
		}
	}
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

func HasDocker() (bool, error) {
	res, err := CmdRun(nil, "docker", "version")
	if err != nil {
		log.Println("HasDocker", err)
		return false, err
	}
	return res == 0 || res == 1, nil
}

func GetProductVersion() (string, error) {
	fullExe_, err := os.Executable()
	if err != nil {
		return "", err
	}
	fv, err := fileversion.New(fullExe_)
	if err != nil {
		return "", err
	}
	return fv.ProductVersion(), nil
}

func ErrorModal(title, message string) int {
	return walk.MsgBox(nil, title, message, walk.MsgBoxTopMost|walk.MsgBoxOK|walk.MsgBoxIconError)
}

func DiscoverDockerPathAndPatchEnv(wait bool) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	sfx := ""
	for i := 0; i <= 10; i++ {
		pathValue, _, err := k.GetStringValue("PATH")
		if err != nil {
			log.Fatal(err)
		}

		pp := strings.Split(pathValue, ";")
		for _, v := range pp {
			if strings.Contains(strings.ToLower(v), "docker") {
				sfx = sfx + ";" + v
			}
		}
		if sfx != "" {
			break
		}
		time.Sleep(5 * time.Second)
	}

	if sfx != "" {
		fmt.Println(sfx)
		w32.SetEnvironmentVariable("PATH", os.Getenv("PATH")+sfx)
	}
}

func getMSIProductCodeByName(productName string) (string, error) {
	l, err := gowin32.GetInstalledProducts()
	if err != nil {
		return "", err
	}

	for _, v := range l {
		n, err := gowin32.GetInstalledProductProperty(v, gowin32.InstallPropertyProductName)
		if err != nil {
			return "", err
		}

		if strings.HasPrefix(n, productName) {
			return v, nil
		}
	}
	return "", errProductNotFound
}

// trunk excessive number (build), so that semver could parse it
// example: "11.22.33.44" -> "11.22.33"
func normalizeVersion(v string) string {
	p := strings.Split(v, ".")
	if len(p) > 3 {
		r := ""
		p := p[:3]
		for i, e := range p {
			r += e
			if i < 2 {
				r += "."
			}
		}
		return r
	}
	return v
}

func IsWSLUpdated() (bool, error) {
	wslUpdateProductCode, err := getMSIProductCodeByName("Windows Subsystem for Linux Update")
	if errors.Is(err, errProductNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	log.Println("wslUpdateProductCode", wslUpdateProductCode)

	state := gowin32.GetInstalledProductState(wslUpdateProductCode)
	if state != gowin32.InstallStateDefault {
		return false, nil
	}

	installedVer, err := gowin32.GetInstalledProductProperty(wslUpdateProductCode, gowin32.InstallPropertyVersionString)
	if err != nil {
		return false, errors.Wrap(err, "gowin32.GetInstalledProductProperty")
	}
	log.Println("IsWSLUpdated > installedVer", installedVer)

	pkg, err := gowin32.OpenInstallerPackage(GetTmpDir() + "\\wsl_update_x64.msi")
	if err != nil {
		return false, errors.Wrap(err, "gowin32.OpenInstallerPackage")
	}
	defer pkg.Close()

	fileVer, err := pkg.GetProductProperty("ProductVersion")
	if err != nil {
		return false, errors.Wrap(err, "gowin32.GetProductProperty")
	}
	log.Println("IsWSLUpdated > fileVer", fileVer)

	semverFileVer, err := semver.Parse(normalizeVersion(fileVer))
	if err != nil {
		return false, errors.Wrap(err, "semver.Parse")
	}
	semverInstalledVer, err := semver.Parse(normalizeVersion(installedVer))
	if err != nil {
		return false, errors.Wrap(err, "semver.Parse")
	}
	log.Println("IsWSLUpdated > fileVer, installedVer >", semverFileVer, semverInstalledVer)

	// semverInstalledVer >= semverFileVer
	return semverInstalledVer.Compare(semverFileVer) >= 0, nil
}

func OpenUrlInBrowser(url string) {
	native.ShellExecuteAndWait(
		0,
		"",
		"rundll32",
		"url.dll,FileProtocolHandler "+url,
		"",
		syscall.SW_NORMAL)
}
