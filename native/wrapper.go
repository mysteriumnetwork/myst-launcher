package native

import (
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"syscall"
)

var (
	user32DLL          = windows.NewLazyDLL("user32.dll")
	switchToThisWindow = user32DLL.NewProc("SwitchToThisWindow")
)

func SwitchToThisWindow(hwnd win.HWND, f bool) int32 {
	ret, _, _ := syscall.Syscall(switchToThisWindow.Addr(), 2,
		uintptr(hwnd),
		uintptr(win.BoolToBOOL(f)),
		0,
	)
	return int32(ret)
}
