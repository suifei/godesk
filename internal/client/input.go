// package client

// import (
// 	"time"

// 	"github.com/faiface/pixel"
// 	"github.com/faiface/pixel/pixelgl"
// 	"github.com/suifei/godesk/internal/protocol"
// 	"github.com/suifei/godesk/pkg/log"
// 	"github.com/suifei/godesk/pkg/network"
// )

// type InputHandler struct {
// 	window  *pixelgl.Window
// 	display *Display
// 	conn    *network.TCPConnection
// }

// func NewInputHandler(window *pixelgl.Window, display *Display, conn *network.TCPConnection) *InputHandler {
// 	return &InputHandler{
// 		window:  window,
// 		display: display,
// 		conn:    conn,
// 	}
// }

// func (h *InputHandler) Start() {
// 	go h.pollEvents()
// }

// func (h *InputHandler) pollEvents() {
// 	for !h.window.Closed() {
// 		h.checkMouseEvents()
// 		h.checkKeyEvents()
// 		h.window.UpdateInput()
// 		time.Sleep(0 * time.Millisecond) // 小小的睡眠以减少 CPU 使用
// 	}
// }

// func (h *InputHandler) checkMouseEvents() {
// 	buttons := []pixelgl.Button{
// 		pixelgl.MouseButtonLeft,
// 		pixelgl.MouseButtonRight,
// 		pixelgl.MouseButtonMiddle,
// 	}
// 	eventTypes := []protocol.MouseEvent_EventType{
// 		protocol.MouseEvent_LEFT_DOWN,
// 		protocol.MouseEvent_RIGHT_DOWN,
// 		protocol.MouseEvent_MIDDLE_DOWN,
// 	}

// 	for i, button := range buttons {
// 		if h.window.JustPressed(button) {
// 			h.sendMouseEvent(eventTypes[i])
// 		}
// 		if h.window.JustReleased(button) {
// 			h.sendMouseEvent(eventTypes[i] + 1) // UP event is always next to DOWN event
// 		}
// 	}

// 	currentPos := h.window.MousePosition()
// 	if currentPos != h.window.MousePreviousPosition() {
// 		h.sendMouseEvent(protocol.MouseEvent_MOVE)
// 	}

// 	scroll := h.window.MouseScroll()
// 	if scroll.Y != 0 {
// 		h.sendMouseEvent(protocol.MouseEvent_SCROLL)
// 	}
// }

// func (h *InputHandler) checkKeyEvents() {
// 	for key := pixelgl.KeySpace; key <= pixelgl.KeyLast; key++ {
// 		if h.window.JustPressed(key) {
// 			h.sendKeyEvent(protocol.KeyEvent_KEY_DOWN, key)
// 		}
// 		if h.window.JustReleased(key) {
// 			h.sendKeyEvent(protocol.KeyEvent_KEY_UP, key)
// 		}
// 	}
// }

// func (h *InputHandler) sendMouseEvent(eventType protocol.MouseEvent_EventType) {
// 	pos := h.window.MousePosition()
// 	scaledPos := h.scaleMousePosition(pos)
// 	log.Infof("Mouse event: %v, Scaled position: %v", eventType, scaledPos)

// 	event := &protocol.InputEvent{
// 		Event: &protocol.InputEvent_MouseEvent{
// 			MouseEvent: &protocol.MouseEvent{
// 				EventType:   eventType,
// 				X:           int32(scaledPos.X),
// 				Y:           int32(scaledPos.Y),
// 				ScrollDelta: int32(h.window.MouseScroll().Y),
// 			},
// 		},
// 		Timestamp: time.Now().UnixNano(),
// 	}

// 	h.sendEventToServer(event)
// }

// func (h *InputHandler) sendKeyEvent(eventType protocol.KeyEvent_EventType, key pixelgl.Button) {
// 	event := &protocol.InputEvent{
// 		Event: &protocol.InputEvent_KeyEvent{
// 			KeyEvent: &protocol.KeyEvent{
// 				EventType: eventType,
// 				KeyCode:   int32(key),
// 				Shift:     h.window.Pressed(pixelgl.KeyLeftShift) || h.window.Pressed(pixelgl.KeyRightShift),
// 				Ctrl:      h.window.Pressed(pixelgl.KeyLeftControl) || h.window.Pressed(pixelgl.KeyRightControl),
// 				Alt:       h.window.Pressed(pixelgl.KeyLeftAlt) || h.window.Pressed(pixelgl.KeyRightAlt),
// 				Meta:      h.window.Pressed(pixelgl.KeyLeftSuper) || h.window.Pressed(pixelgl.KeyRightSuper),
// 			},
// 		},
// 		Timestamp: time.Now().UnixNano(),
// 	}

// 	h.sendEventToServer(event)
// }

// func (h *InputHandler) sendEventToServer(event *protocol.InputEvent) {
// 	msg := &protocol.Message{
// 		Payload: &protocol.Message_InputEvent{
// 			InputEvent: event,
// 		},
// 	}

// 	err := h.conn.Send(msg)
// 	if err != nil {
// 		log.Errorf("Failed to send input event: %v", err)
// 	}
// }
// func (h *InputHandler) scaleMousePosition(pos pixel.Vec) pixel.Vec {
// 	scale := min(
// 		h.window.Bounds().W()/h.display.spriteRect.W(),
// 		h.window.Bounds().H()/h.display.spriteRect.H(),
// 	)

// 	imagePos := h.window.Bounds().Center().Sub(h.display.spriteRect.Center().Scaled(scale))

// 	// 计算相对于图像的坐标
// 	imageX := (pos.X - imagePos.X) / scale
// 	imageY := (pos.Y - imagePos.Y) / scale

// 	// 翻转 Y 坐标
// 	flippedY := h.display.spriteRect.H() - imageY

// 	var r pixel.Vec

// 	if imageY < 0 || imageX < 0 {
// 		r = pixel.V(0, 0)
// 	} else if imageX > h.display.spriteRect.W() || imageY > h.display.spriteRect.H() {
// 		r = pixel.V(h.display.spriteRect.W(), h.display.spriteRect.H())
// 	} else {
// 		r = pixel.V(imageX, flippedY)
// 	}

// 	log.Debugf("Mouse position: %v, Scaled position: %v, Flipped Y: %v", pos, r, flippedY)
// 	return r
// }
package client

import (
	"github.com/suifei/godesk/internal/protocol"
	"github.com/suifei/godesk/pkg/log"
	"time"
)

type InputEventType int

const (
	KeyboardEvent InputEventType = iota
	MouseEvent
)

type MouseButton int

const (
	NoButton MouseButton = iota
	LeftButton
	RightButton
)

type InputEvent struct {
	Type    InputEventType
	KeyCode int
	X, Y    int
	Button  MouseButton
	Down    bool
}

type InputHandler struct {
	display *Display
	events  chan *protocol.InputEvent
}

func NewInputHandler(display *Display) *InputHandler {
	return &InputHandler{
		display: display,
		events:  make(chan *protocol.InputEvent, 100),
	}
}

func (h *InputHandler) Start() {
	go h.handleInputEvents()
}

func (h *InputHandler) handleInputEvents() {
	for event := range h.display.InputEvents() {
		switch event.Type {
		case KeyboardEvent:
			h.handleKeyEvent(event)
		case MouseEvent:
			h.handleMouseEvent(event)
		}
	}
}

func (h *InputHandler) handleKeyEvent(event InputEvent) {
	log.Debugf("Key event: code=%d, down=%v", event.KeyCode, event.Down)
	h.events <- &protocol.InputEvent{
		Event: &protocol.InputEvent_KeyEvent{
			KeyEvent: &protocol.KeyEvent{
				EventType: h.getKeyEventType(event.Down),
				KeyCode:   int32(event.KeyCode),
				// Set Shift, Ctrl, Alt, Meta as needed
			},
		},
		Timestamp: time.Now().UnixNano(),
	}
}

func (h *InputHandler) handleMouseEvent(event InputEvent) {
	log.Debugf("Mouse event: x=%d, y=%d, button=%v, down=%v", event.X, event.Y, event.Button, event.Down)
	h.events <- &protocol.InputEvent{
		Event: &protocol.InputEvent_MouseEvent{
			MouseEvent: &protocol.MouseEvent{
				EventType: h.getMouseEventType(event),
				X:         int32(event.X),
				Y:         int32(event.Y),
			},
		},
		Timestamp: time.Now().UnixNano(),
	}
}

func (h *InputHandler) getKeyEventType(down bool) protocol.KeyEvent_EventType {
	if down {
		return protocol.KeyEvent_KEY_DOWN
	}
	return protocol.KeyEvent_KEY_UP
}

func (h *InputHandler) getMouseEventType(event InputEvent) protocol.MouseEvent_EventType {
	switch event.Button {
	case LeftButton:
		if event.Down {
			return protocol.MouseEvent_LEFT_DOWN
		}
		return protocol.MouseEvent_LEFT_UP
	case RightButton:
		if event.Down {
			return protocol.MouseEvent_RIGHT_DOWN
		}
		return protocol.MouseEvent_RIGHT_UP
	default:
		return protocol.MouseEvent_MOVE
	}
}

func (h *InputHandler) NextEvent() *protocol.InputEvent {
	select {
	case event := <-h.events:
		return event
	default:
		return nil
	}
}