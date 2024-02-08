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

	"github.com/bi-zone/go-fileversion"
	"github.com/blang/semver/v4"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/gonutz/w32"
	"github.com/mysteriumnetwork/myst-launcher/native"
	"github.com/pkg/errors"
	"github.com/scjalliance/comshim"
	"github.com/winlabs/gowin32"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	_const "github.com/mysteriumnetwork/myst-launcher/const"
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
	comshim.Add(1)
	defer comshim.Done()

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
	_, err := os.Stat(GetMystNodeLauncherExeLegacyPath())
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

	verDst, err := fileversion.New(GetMystNodeLauncherExeLegacyPath())
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

func EnableAutorun(en bool) error {
	return nil
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
	_ = os.RemoveAll(GetMystNodeLauncherLegacyPath())

	return nil
}

func GetMystNodeLauncherLegacyPath() string {
	return os.Getenv("ProgramFiles") + "\\MystNodeLauncher"
}

func GetMystNodeLauncherExeLegacyPath() string {
	return os.Getenv("ProgramFiles") + "\\MystNodeLauncher" + "\\" + launcherExe
}

func CreateAutostartShortcut(args string) error {
	shcDst := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Startup", launcherLnk)
	return CreateShortcut(shcDst, GetMystNodeLauncherExeLegacyPath(), args)
}

func CreateDesktopShortcut(args string) error {
	shcDst := path.Join(os.Getenv("USERPROFILE"), "Desktop", launcherLnk)
	return CreateShortcut(shcDst, GetMystNodeLauncherExeLegacyPath(), args)
}

func CreateStartMenuShortcut(args string) error {
	dir := path.Join(os.Getenv("APPDATA"), "Microsoft\\Windows\\Start Menu\\Programs\\Mysterium Network")
	os.Mkdir(dir, os.ModePerm)
	shcDst := path.Join(dir, launcherLnk)
	return CreateShortcut(shcDst, GetMystNodeLauncherExeLegacyPath(), args)
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

// kill process by name except current process
func KillProcessByName(exeName string) error {
	pid := windows.GetCurrentProcessId()

	h, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)

	p := windows.ProcessEntry32{Size: processEntrySize}
	for {
		if err := windows.Process32Next(h, &p); err != nil {
			return err
		}

		if s := windows.UTF16ToString(p.ExeFile[:]); s == exeName {
			if pid != p.ProcessID {
				if err := TerminateProcess(p.ProcessID, 0); err != nil {
					return err
				}
			}
		}
	}

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

func getMSIProductCodeByName(productName string) (string, string, error) {
	l, err := gowin32.GetInstalledProducts()
	if err != nil {
		return "", "", err
	}

	for _, v := range l {
		n, err := gowin32.GetInstalledProductProperty(v, gowin32.InstallPropertyProductName)
		if err != nil {
			return "", "", err
		}
		// fmt.Println(">", productName, n, v, strings.HasPrefix(n, productName))

		if strings.HasPrefix(n, productName) {
			ver, err := gowin32.GetInstalledProductProperty(v, gowin32.InstallPropertyVersionString)
			if err != nil {
				return "", "", err
			}

			return v, normalizeVersion(ver), nil
		}
	}
	return "", "", errProductNotFound
}

// trunk excessive number (build), so that semver could parse it
// example: "11.22.33.44" -> "11.22.33"
func normalizeVersion(v string) string {
	// trim -xxx postfix
	p := strings.Split(v, "-")
	v = p[0]

	p = strings.Split(v, ".")
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

func GetInstalledPackageVersion() (string, error) {
	mystProductCode, ver, err := getMSIProductCodeByName("Mysterium Launcher x64")
	if errors.Is(err, errProductNotFound) {
		return "", nil
	}
	if err != nil {
		return "", nil
	}

	ver = normalizeVersion(ver)
	fmt.Println("mystProductCode", mystProductCode, ver)
	return ver, nil
}

func LauncherMSIHasUpdateOrPkgNI(latest string, currentVer *string) (bool, error) {
	mystProductCode, ver, err := getMSIProductCodeByName("Mysterium Launcher x64")
	if errors.Is(err, errProductNotFound) {
		return true, nil
	}
	if err != nil {
		return false, nil
	}

	latest = normalizeVersion(latest)
	// fmt.Println("mystProductCode", mystProductCode, ver, latest)
	_ = mystProductCode

	semverLatest, err := semver.Parse(normalizeVersion(latest))
	if err != nil {
		return false, errors.Wrap(err, "semver.Parse")
	}

	// ver = "1.0.35"
	current := normalizeVersion(ver)
	semverCurrent, err := semver.Parse(normalizeVersion(current))
	if err != nil {
		return false, errors.Wrap(err, "semver.Parse")
	}
	log.Println("semverLatest>", semverLatest, semverCurrent, semverLatest.Compare(semverCurrent))
	*currentVer = ver

	return semverLatest.Compare(semverCurrent) > 0, nil
}

func IsWSLUpdated() (bool, error) {
	wslUpdateProductCode, _, err := getMSIProductCodeByName("Windows Subsystem for Linux Update")
	if errors.Is(err, errProductNotFound) {
		return false, nil
	}
	if err != nil {
		return false, nil
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

////////////////////////////////////////////////////////////////////////////////////////////////////
// win32 console utils. Borrowed from https://github.com/yuk7/wsldl/blob/main/src/lib/utils/utils.go

// AllocConsole calls AllocConsole API in Windows kernel32
func AllocConsole(attach bool) {
	kernel32, _ := syscall.LoadDLL("Kernel32.dll")

	if attach {
		attach, _ := kernel32.FindProc("AttachConsole")
		const ATTACH_PARENT_PROCESS = ^uintptr(0)
		attach.Call(ATTACH_PARENT_PROCESS)
	} else {
		alloc, _ := kernel32.FindProc("AllocConsole")
		alloc.Call()
	}

	hout, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	herr, _ := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)
	hin, _ := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)
	os.Stdout = os.NewFile(uintptr(hout), "/dev/stdout")
	os.Stderr = os.NewFile(uintptr(herr), "/dev/stderr")
	os.Stdin = os.NewFile(uintptr(hin), "/dev/stdin")
}

// SetConsoleTitle calls SetConsoleTitleW API in Windows kernel32
func SetConsoleTitle(title string) {
	kernel32, _ := syscall.LoadDLL("Kernel32.dll")
	proc, _ := kernel32.FindProc("SetConsoleTitleW")
	pTitle, _ := syscall.UTF16PtrFromString(title)
	syscall.Syscall(proc.Addr(), 1, uintptr(unsafe.Pointer(pTitle)), 0, 0)
	return
}

// FreeConsole calls FreeConsole API in Windows kernel32
func FreeConsole() error {
	kernel32, _ := syscall.LoadDLL("Kernel32.dll")
	proc, err := kernel32.FindProc("FreeConsole")
	if err != nil {
		return err
	}
	proc.Call()
	return nil
}

func HideFile(path string, hide bool) (string, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	flags := uint32(0)
	if hide {
		flags |= syscall.FILE_ATTRIBUTE_HIDDEN
	}
	err = syscall.SetFileAttributes(p, flags)
	if err != nil {
		return "", err
	}
	return path, nil
}

func RunMsi(msi string) error {
	system32, err := windows.GetSystemDirectory()
	if err != nil {
		return err
	}
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer devNull.Close()

	attr := &os.ProcAttr{
		Sys:   &syscall.SysProcAttr{},
		Files: []*os.File{devNull, devNull, devNull},
		Dir:   filepath.Dir(msi),
	}
	msiexec := filepath.Join(system32, "msiexec.exe")

	_, err = os.StartProcess(msiexec, []string{msiexec, "/qb!-", "/i", filepath.Base(msi), `RUNAFTER=1`}, attr)
	return err
}

func IsAdmin() bool {
	return w32.SHIsUserAnAdmin()
}
