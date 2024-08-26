package monitor

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	MONITOR_DEFAULTTONULL    = 0x00000000
	MONITOR_DEFAULTTOPRIMARY = 0x00000001
	MONITOR_DEFAULTTONEAREST = 0x00000002
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfo      = user32.NewProc("GetMonitorInfoW")
	procMonitorFromWindow   = user32.NewProc("MonitorFromWindow")
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}

type Monitor struct {
	Handle uintptr
	Info   MONITORINFO
}

func getAllMonitors() ([]Monitor, error) {
	var monitors []Monitor

	callback := syscall.NewCallback(func(hMonitor uintptr, hdcMonitor uintptr, lprcMonitor *RECT, dwData uintptr) uintptr {
		var mi MONITORINFO
		mi.CbSize = uint32(unsafe.Sizeof(mi))
		ret, _, _ := procGetMonitorInfo.Call(hMonitor, uintptr(unsafe.Pointer(&mi)))
		if ret != 0 {
			monitors = append(monitors, Monitor{Handle: hMonitor, Info: mi})
		}
		return 1 // 继续枚举
	})

	ret, _, _ := procEnumDisplayMonitors.Call(0, 0, callback, 0)
	if ret == 0 {
		return nil, fmt.Errorf("EnumDisplayMonitors failed")
	}

	return monitors, nil
}

func getCurrentMonitor(hwnd windows.HWND, monitors []Monitor) (*Monitor, error) {
	hMonitor, _, _ := procMonitorFromWindow.Call(
		uintptr(hwnd),
		MONITOR_DEFAULTTONEAREST,
	)

	if hMonitor == 0 {
		return nil, fmt.Errorf("failed to get monitor for window")
	}

	for i, m := range monitors {
		if m.Handle == hMonitor {
			return &monitors[i], nil
		}
	}

	return nil, fmt.Errorf("monitor not found in list")
}

type WinDisplay struct {
	window         windows.HWND
	allMonitors    []Monitor
	currentMonitor *Monitor
}

func (d *WinDisplay) UpdateDisplayInfo() error {
	var err error
	d.allMonitors, err = getAllMonitors()
	if err != nil {
		return fmt.Errorf("error getting all monitors: %v", err)
	}

	d.currentMonitor, err = getCurrentMonitor(d.window, d.allMonitors)
	if err != nil {
		return fmt.Errorf("error getting current monitor: %v", err)
	}

	return nil
}

func (d *WinDisplay) GetCurrentResolution() (int, int) {
	if d.currentMonitor == nil {
		return 0, 0
	}
	width := d.currentMonitor.Info.RcMonitor.Right - d.currentMonitor.Info.RcMonitor.Left
	height := d.currentMonitor.Info.RcMonitor.Bottom - d.currentMonitor.Info.RcMonitor.Top
	return int(width), int(height)
}

func TestMonitor() {
	// 假设已经创建了窗口并获得了 HWND
	var hwnd windows.HWND // 这里应该是您实际的窗口句柄

	display := &WinDisplay{window: hwnd}
	err := display.UpdateDisplayInfo()
	if err != nil {
		fmt.Printf("Error updating display info: %v\n", err)
		return
	}

	fmt.Printf("Total monitors: %d\n", len(display.allMonitors))
	width, height := display.GetCurrentResolution()
	fmt.Printf("Current monitor resolution: %dx%d\n", width, height)

	// 打印所有显示器的信息
	for i, monitor := range display.allMonitors {
		fmt.Printf("Monitor %d:\n", i+1)
		fmt.Printf("  Resolution: %dx%d\n",
			monitor.Info.RcMonitor.Right-monitor.Info.RcMonitor.Left,
			monitor.Info.RcMonitor.Bottom-monitor.Info.RcMonitor.Top)
		fmt.Printf("  Position: (%d, %d)\n",
			monitor.Info.RcMonitor.Left, monitor.Info.RcMonitor.Top)
		if monitor.Handle == display.currentMonitor.Handle {
			fmt.Println("  (Current Monitor)")
		}
	}
}
