//go:build darwin
// +build darwin

/*
 * copyright
 */

package utils

/*
#include <stdlib.h>
#include "libproc.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

func TerminateProcess(pid uint32, exitcode int) error {
  sig := C.int(2) //9-kill
  ret := C.proc_terminate(17234, &sig)
  if ret != 0 {
    return errors.New("Process not found")
  }
  return nil
}

func IsProcessRunning(name string) bool {
	pids := ps()
	for _, p := range pids {
		if procName(p) == name {
			return true
		}
	}
	return false
}

func IsProcessRunningExt(exeName, fullpath string) (uint32, error) {
	pids := ps()
	for _, p := range pids {
		if procName(p) == exeName && pathFromPID(p) == fullpath {
			return uint32(p), nil
		}
	}
	return 0, nil
}

func procName(pid C.pid_t) string {
	buffer := make([]C.char, 8192)
	size := C.proc_name(C.int(pid), unsafe.Pointer(&buffer[0]), C.uint32_t(len(buffer))*C.uint32_t(unsafe.Sizeof(buffer[0])))
	return C.GoStringN(&buffer[0], size)
}

func pathFromPID(pid C.pid_t) string {
	buffer := make([]C.char, 1024)
	size := C.proc_pidpath(C.int(pid), unsafe.Pointer(&buffer[0]), C.uint32_t(len(buffer))*C.uint32_t(unsafe.Sizeof(buffer[0])))
	return C.GoStringN(&buffer[0], size)
}

func ps() []C.pid_t {
	num := C.proc_listallpids(nil, 0)
	pids := make([]C.int, num*2)
	num = C.proc_listallpids(unsafe.Pointer(&pids[0]), C.int(len(pids))*C.int(unsafe.Sizeof(pids[0])))
	result := make([]C.pid_t, 0)
	for i := 0; i < int(num); i++ {
		result = append(result, C.pid_t(pids[i]))
	}
	return result
}
