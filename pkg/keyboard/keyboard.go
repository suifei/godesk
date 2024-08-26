package keyboard

import (
	"fmt"
	"syscall"

	"github.com/suifei/godesk/internal/protocol"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procKeyBdEvent    = user32.NewProc("keybd_event")
	procMapVirtualKey = user32.NewProc("MapVirtualKeyW")
)

const (
	KEYEVENTF_KEYUP = 0x0002
	MAPVK_VK_TO_VSC = 0x0
)

type Keyboard struct {
	// 可以添加一些状态，如果需要的话
}

func New() *Keyboard {
	return &Keyboard{}
}

func (k *Keyboard) KeyDown(keyCode int) error {
	scanCode := k.MapVirtualKey(uint32(keyCode), MAPVK_VK_TO_VSC)
	ret, _, err := procKeyBdEvent.Call(
		uintptr(keyCode),
		uintptr(scanCode),
		0,
		0)
	if ret == 0 {
		return fmt.Errorf("keyDown failed for key %d: %v", keyCode, err)
	}
	return nil
}

func (k *Keyboard) KeyUp(keyCode int) error {
	scanCode := k.MapVirtualKey(uint32(keyCode), MAPVK_VK_TO_VSC)
	ret, _, err := procKeyBdEvent.Call(
		uintptr(keyCode),
		uintptr(scanCode),
		uintptr(KEYEVENTF_KEYUP),
		0)
	if ret == 0 {
		return fmt.Errorf("keyUp failed for key %d: %v", keyCode, err)
	}
	return nil
}

func (k *Keyboard) MapVirtualKey(uCode uint32, uMapType uint32) uint32 {
	ret, _, _ := procMapVirtualKey.Call(
		uintptr(uCode),
		uintptr(uMapType))
	return uint32(ret)
}

func (k *Keyboard) HandleKeyEvent(event *protocol.KeyEvent) error {
	keyCode := int(event.KeyCode)

	// 处理修饰键
	modifiers := []struct {
		isPressed bool
		keyCode   int
	}{
		{event.Shift, 0x10}, // VK_SHIFT
		{event.Ctrl, 0x11},  // VK_CONTROL
		{event.Alt, 0x12},   // VK_MENU (ALT)
		{event.Meta, 0x5B},  // VK_LWIN
	}

	// 按下修饰键
	for _, mod := range modifiers {
		if mod.isPressed {
			if err := k.KeyDown(mod.keyCode); err != nil {
				return err
			}
		}
	}

	// 处理主键
	if event.EventType == protocol.KeyEvent_KEY_DOWN {
		if err := k.KeyDown(keyCode); err != nil {
			return err
		}
	} else if event.EventType == protocol.KeyEvent_KEY_UP {
		if err := k.KeyUp(keyCode); err != nil {
			return err
		}
	}

	// 释放修饰键（如果是 KEY_UP 事件）
	if event.EventType == protocol.KeyEvent_KEY_UP {
		for i := len(modifiers) - 1; i >= 0; i-- {
			if modifiers[i].isPressed {
				if err := k.KeyUp(modifiers[i].keyCode); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
