package server

import (
	"github.com/suifei/godesk/internal/protocol"
	"github.com/suifei/godesk/pkg/log"
	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazyDLL("user32.dll")
	procMouseEvent          = user32.NewProc("mouse_event")
	procKeyboardEvent       = user32.NewProc("keybd_event")
	procGetSystemMetrics    = user32.NewProc("GetSystemMetrics")
	procSetCursorPos        = user32.NewProc("SetCursorPos")
	inputKeyboardWVKKeyDown = uint32(0x0000)
	inputKeyboardWVKKeyUp   = uint32(0x0002)
)

const (
	MOUSEEVENTF_MOVE       = 0x0001
	MOUSEEVENTF_LEFTDOWN   = 0x0002
	MOUSEEVENTF_LEFTUP     = 0x0004
	MOUSEEVENTF_RIGHTDOWN  = 0x0008
	MOUSEEVENTF_RIGHTUP    = 0x0010
	MOUSEEVENTF_MIDDLEDOWN = 0x0020
	MOUSEEVENTF_MIDDLEUP   = 0x0040
	MOUSEEVENTF_WHEEL      = 0x0800
	WHEEL_DELTA            = 120
	SM_CXSCREEN            = 0
	SM_CYSCREEN            = 1
)

func HandleInputEvent(event *protocol.InputEvent) {
	switch e := event.Event.(type) {
	case *protocol.InputEvent_MouseEvent:
		handleMouseEvent(e.MouseEvent)
	case *protocol.InputEvent_KeyEvent:
		handleKeyEvent(e.KeyEvent)
	default:
		log.Warnf("Unknown input event type: %T", e)
	}
}

func handleMouseEvent(event *protocol.MouseEvent) {
	switch event.EventType {
	case protocol.MouseEvent_MOVE:
		// 只有在移动事件时才设置光标位置
		SetCursorPos(int(event.X), int(event.Y))
	case protocol.MouseEvent_LEFT_DOWN:
		MouseEvent(MOUSEEVENTF_LEFTDOWN)
	case protocol.MouseEvent_LEFT_UP:
		MouseEvent(MOUSEEVENTF_LEFTUP)
	case protocol.MouseEvent_RIGHT_DOWN:
		MouseEvent(MOUSEEVENTF_RIGHTDOWN)
	case protocol.MouseEvent_RIGHT_UP:
		MouseEvent(MOUSEEVENTF_RIGHTUP)
	case protocol.MouseEvent_MIDDLE_DOWN:
		MouseEvent(MOUSEEVENTF_MIDDLEDOWN)
	case protocol.MouseEvent_MIDDLE_UP:
		MouseEvent(MOUSEEVENTF_MIDDLEUP)
	case protocol.MouseEvent_SCROLL:
		// 对于滚轮事件，我们可能需要先设置光标位置，然后再触发滚轮事件
		SetCursorPos(int(event.X), int(event.Y))
		MouseEvent(MOUSEEVENTF_WHEEL, int32(event.ScrollDelta)*WHEEL_DELTA)
	default:
		log.Warnf("Unhandled mouse event type: %v", event.EventType)
	}
	log.Debugf("Handled mouse event: type=%v, x=%d, y=%d", event.EventType, event.X, event.Y)
}

func MouseEvent(dwFlags uint32, mouseData ...int32) {
	var md int32
	if len(mouseData) > 0 {
		md = mouseData[0]
	}
	procMouseEvent.Call(uintptr(dwFlags), 0, 0, uintptr(md), 0)
}

func handleKeyEvent(event *protocol.KeyEvent) {
	vkCode := byte(event.KeyCode)
	if event.EventType == protocol.KeyEvent_KEY_DOWN {
		keyDown(vkCode)
	} else if event.EventType == protocol.KeyEvent_KEY_UP {
		keyUp(vkCode)
	}
	log.Debugf("Handled key event: type=%v, keyCode=%d, shift=%v, ctrl=%v, alt=%v, meta=%v",
		event.EventType, event.KeyCode, event.Shift, event.Ctrl, event.Alt, event.Meta)
}

func SetCursorPos(x, y int) {
	screenWidth, err := GetSystemMetrics(SM_CXSCREEN)
	if err != nil {
		log.Errorf("Failed to get screen width: %v", err)
		return
	}
	screenHeight, err := GetSystemMetrics(SM_CYSCREEN)
	if err != nil {
		log.Errorf("Failed to get screen height: %v", err)
		return
	}

	// 调整坐标
	x = clamp(x, 0, screenWidth-1)
	y = clamp(y, 0, screenHeight-1)

	log.Debugf("Setting cursor position: x=%d, y=%d (screen: %dx%d)", x, y, screenWidth, screenHeight)
	_, _, err = procSetCursorPos.Call(uintptr(x), uintptr(y))
	if err != nil && err != windows.ERROR_SUCCESS {
		log.Errorf("Failed to set cursor position: %v", err)
	}
}

// clamp 函数用于将值限制在指定范围内
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func GetSystemMetrics(nIndex int) (int, error) {
	ret, _, err := procGetSystemMetrics.Call(uintptr(nIndex))
	if ret == 0 {
		return 0, err
	}
	return int(ret), nil
}

func keyDown(vkCode byte) {
	log.Infof("Sending key down event: vkCode=%d", vkCode)
	procKeyboardEvent.Call(uintptr(vkCode), 0, uintptr(inputKeyboardWVKKeyDown), 0)
}

func keyUp(vkCode byte) {
	log.Infof("Sending key down event: vkCode=%d", vkCode)
	procKeyboardEvent.Call(uintptr(vkCode), 0, uintptr(inputKeyboardWVKKeyUp), 0)
}
