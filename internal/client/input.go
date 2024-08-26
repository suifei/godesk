package client

import (
	"time"

	"github.com/suifei/godesk/internal/protocol"
	"github.com/suifei/godesk/pkg/log"
	"github.com/suifei/godesk/pkg/network"
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
	LeftButtonDbClick
	RightButton
	RightButtonDbClick
	MiddleButton
	MiddleButtonDbClick
	Scroll
)

type InputEvent struct {
	Type        InputEventType
	KeyCode     int
	X, Y        int
	Button      MouseButton
	Down        bool
	ScrollDelta int
	Shift       bool
	Ctrl        bool
	Alt         bool
	Meta        bool
}

type InputHandler struct {
	display *Display
	conn    *network.TCPConnection
}

func NewInputHandler(display *Display, conn *network.TCPConnection) *InputHandler {
	return &InputHandler{
		display: display,
		conn:    conn,
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

func (h *InputHandler) sendEventToServer(event *protocol.InputEvent) {
	msg := &protocol.Message{
		Payload: &protocol.Message_InputEvent{
			InputEvent: event,
		},
	}

	if err := h.conn.Send(msg); err != nil {
		log.Errorf("Error sending input event to server: %v", err)
		h.display.Close()
		return
	}
}
func (h *InputHandler) handleKeyEvent(event InputEvent) {
	log.Debugf("Key event: code=%d, down=%v, shift=%v, ctrl=%v, alt=%v, meta=%v", event.KeyCode, event.Down, event.Shift, event.Ctrl, event.Alt, event.Meta)
	msg := &protocol.InputEvent{
		Event: &protocol.InputEvent_KeyEvent{
			KeyEvent: &protocol.KeyEvent{
				EventType: h.getKeyEventType(event.Down),
				KeyCode:   int32(event.KeyCode),
			},
		},
		Timestamp: time.Now().UnixNano(),
	}

	// Set Shift, Ctrl, Alt, Meta as needed
	msg.Event.(*protocol.InputEvent_KeyEvent).KeyEvent.Shift = event.Shift
	msg.Event.(*protocol.InputEvent_KeyEvent).KeyEvent.Ctrl = event.Ctrl
	msg.Event.(*protocol.InputEvent_KeyEvent).KeyEvent.Alt = event.Alt
	msg.Event.(*protocol.InputEvent_KeyEvent).KeyEvent.Meta = event.Meta

	h.sendEventToServer(msg)
}

func (h *InputHandler) handleMouseEvent(event InputEvent) {
	// log.Debugf("Mouse event: x=%d, y=%d, button=%v, down=%v", event.X, event.Y, event.Button, event.Down)
	msg := &protocol.InputEvent{
		Event: &protocol.InputEvent_MouseEvent{
			MouseEvent: &protocol.MouseEvent{
				EventType:   h.getMouseEventType(event),
				X:           int32(event.X),
				Y:           int32(event.Y),
				ScrollDelta: int32(event.ScrollDelta),
			},
		},
		Timestamp: time.Now().UnixNano(),
	}
	h.sendEventToServer(msg)
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
	case MiddleButton:
		if event.Down {
			return protocol.MouseEvent_MIDDLE_DOWN
		}
		return protocol.MouseEvent_MIDDLE_UP
	case Scroll:
		return protocol.MouseEvent_SCROLL
	case LeftButtonDbClick:
		return protocol.MouseEvent_LEFT_DBLCLICK
	case RightButtonDbClick:
		return protocol.MouseEvent_RIGHT_DBLCLICK
	case MiddleButtonDbClick:
		return protocol.MouseEvent_MIDDLE_DBLCLICK
	default:
		return protocol.MouseEvent_MOVE
	}
}
