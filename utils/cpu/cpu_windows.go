// +build windows

package cpu

import (
	"errors"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

type fileTime struct {
	lowDateTime  uint32
	highDateTime uint32
}

var getSystemTimes *syscall.LazyProc

func init() {
	if runtime.GOOS == "windows" {
		kernel := syscall.NewLazyDLL("kernel32.dll")
		if kernel != nil {
			getSystemTimes = kernel.NewProc("GetSystemTimes")
		}
	}
}
func systemTimes() (uint64, uint64, uint64, error) {
	if getSystemTimes == nil {
		return 0, 0, 0, errors.New("获取系统时间函数为空")
	}
	var idleTime, kernelTime, userTime fileTime
	ret, _, _ := getSystemTimes.Call(uintptr(unsafe.Pointer(&idleTime)), uintptr(unsafe.Pointer(&kernelTime)), uintptr(unsafe.Pointer(&userTime)))
	if ret == 0 {
		return 0, 0, 0, errors.New("获取系统时间失败")
	}
	idleTime64 := uint64(idleTime.highDateTime)<<32 | uint64(idleTime.lowDateTime)
	kernelTime64 := uint64(kernelTime.highDateTime)<<32 | uint64(kernelTime.lowDateTime)
	userTime64 := uint64(userTime.highDateTime)<<32 | uint64(userTime.lowDateTime)
	return idleTime64, kernelTime64, userTime64, nil
}

func GetCpuUseRate() (int, error) {
	oldIdleTime, oldKernelTime, oldUserTime, err := systemTimes()
	if err != nil {
		return 0, err
	}
	time.Sleep(time.Second)
	idleTime, kernelTime, userTime, err := systemTimes()
	if err != nil {
		return 0, err
	}
	idleTimeDelta := idleTime - oldIdleTime
	kernelTimeDelta := kernelTime - oldKernelTime
	userTimeDelta := userTime - oldUserTime
	useRate := 100 - (idleTimeDelta * 100 / (kernelTimeDelta + userTimeDelta))
	return int(useRate), nil
}
