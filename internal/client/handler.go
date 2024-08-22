package client

import (
	"log"

	"github.com/suifei/godesk/internal/protocol"
	"github.com/suifei/godesk/pkg/network"
)

type ClientHandler struct {
	conn    *network.TCPConnection
	display *Display
	input   *InputHandler
}

func NewClientHandler(conn *network.TCPConnection) *ClientHandler {
	display := NewDisplay()
	input := NewInputHandler(display.Window)
	return &ClientHandler{
		conn:    conn,
		display: display,
		input:   input,
	}
}

func (h *ClientHandler) Handle() {
	h.input.Start()
	go h.handleServerMessages()
	h.handleUserInput()
}

func (h *ClientHandler) handleServerMessages() {
	for {
		msg, err := h.conn.Receive()
		if err != nil {
			log.Printf("Error receiving message from server: %v", err)
			return
		}

		switch payload := msg.Payload.(type) {
		case *protocol.Message_ScreenUpdate:
			h.display.Update(payload.ScreenUpdate)
		default:
			log.Printf("Unhandled message type: %T", payload)
		}
	}
}

func (h *ClientHandler) handleUserInput() {
	for event := range h.input.Events() {
		var msg *protocol.Message
		switch e := event.(type) {
		case MouseEvent:
			msg = &protocol.Message{
				Payload: &protocol.Message_MouseEvent{
					MouseEvent: &protocol.MouseEvent{
						EventType: protocol.MouseEvent_EventType(e.EventType),
						X:         int32(e.X),
						Y:         int32(e.Y),
					},
				},
			}
		case KeyEvent:
			msg = &protocol.Message{
				Payload: &protocol.Message_KeyEvent{
					KeyEvent: &protocol.KeyEvent{
						EventType: protocol.KeyEvent_EventType(e.EventType),
						KeyCode:   int32(e.KeyCode),
					},
				},
			}
		}

		if msg != nil {
			if err := h.conn.Send(msg); err != nil {
				log.Printf("Error sending input event to server: %v", err)
			}
		}
	}
}
