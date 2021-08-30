package native

import (
	"syscall"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

var (
	user32DLL          = windows.NewLazyDLL("user32.dll")
	switchToThisWindow = user32DLL.NewProc("SwitchToThisWindow")

	modkernel32   = syscall.NewLazyDLL("kernel32.dll")
	procCopyFileW = modkernel32.NewProc("CopyFileW")
)

func SwitchToThisWindow(hwnd win.HWND, f bool) int32 {
	ret, _, _ := syscall.Syscall(switchToThisWindow.Addr(), 2,
		uintptr(hwnd),
		uintptr(win.BoolToBOOL(f)),
		0,
	)
	return int32(ret)
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
