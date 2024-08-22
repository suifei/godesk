package server

import (
	"log"

	"github.com/suifei/godesk/internal/protocol"
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

// 模拟键盘按下事件
func keyDown(vkCode byte) {
	procKeyboardEvent.Call(uintptr(vkCode), 0, uintptr(inputKeyboardWVKKeyDown), 0)
}

// 模拟键盘松开事件
func keyUp(vkCode byte) {
	procKeyboardEvent.Call(uintptr(vkCode), 0, uintptr(inputKeyboardWVKKeyUp), 0)
}

// 处理键盘事件
func HandleKeyEvent(event *protocol.KeyEvent) {
	if event.EventType == protocol.KeyEvent_KEY_DOWN {
		keyDown(byte(event.KeyCode))
	} else if event.EventType == protocol.KeyEvent_KEY_UP {
		keyUp(byte(event.KeyCode))
	}
	log.Printf("Handled key event: type=%v, keyCode=%d", event.EventType, event.KeyCode)
}

// 处理鼠标事件
func HandleMouseEvent(event *protocol.MouseEvent) {
	switch event.EventType {
	case protocol.MouseEvent_MOVE:
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
		MouseEvent(MOUSEEVENTF_WHEEL, int32(event.ScrollDelta)*WHEEL_DELTA)
	default:
		log.Printf("Unhandled mouse event type: %v", event.EventType)
	}
	log.Printf("Handled mouse event: type=%v, x=%d, y=%d", event.EventType, event.X, event.Y)
}

// 模拟鼠标事件
func MouseEvent(dwFlags uint32, mouseData ...int32) {
	var md int32
	if len(mouseData) > 0 {
		md = mouseData[0]
	}
	procMouseEvent.Call(uintptr(dwFlags), 0, 0, uintptr(md), 0)
}

// 设置鼠标位置
func SetCursorPos(x, y int) {
	screenWidth, _ := GetSystemMetrics(SM_CXSCREEN)
	screenHeight, _ := GetSystemMetrics(SM_CYSCREEN)

	// 将相对坐标转换为绝对坐标
	absX := int(float64(x) / float64(screenWidth) * 65535)
	absY := int(float64(y) / float64(screenHeight) * 65535)

	procSetCursorPos.Call(uintptr(absX), uintptr(absY))
}

// 获取系统指标
func GetSystemMetrics(nIndex int) (int, error) {
	ret, _, err := procGetSystemMetrics.Call(uintptr(nIndex))
	if ret == 0 {
		return 0, err
	}
	return int(ret), nil
}
