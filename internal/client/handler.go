package client

import (
	"log"
	"net"
	"time"

	"github.com/suifei/godesk/internal/protocol"
	"github.com/suifei/godesk/pkg/network"
)

type ClientHandler struct {
	conn    *network.TCPConnection
	display *Display
	input   *InputHandler
	running bool
}

func NewClientHandler(serverAddr string) (*ClientHandler, error) {
	log.Printf("Attempting to connect to %s", serverAddr)
	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)
		return nil, err
	}
	log.Printf("Successfully connected to %s", serverAddr)

	tcpConn := network.NewTCPConnection(conn)
	display := NewDisplay()
	input := NewInputHandler(display.Window, display)

	return &ClientHandler{
		conn:    tcpConn,
		display: display,
		input:   input,
		running: true,
	}, nil
}

func (h *ClientHandler) Handle() {
	defer h.conn.Close()

	h.input.Start()
	go h.handleServerMessages()
	go h.handleUserInput()

	log.Println("Starting main display loop")
	for h.running && !h.display.ShouldClose() {
		h.display.Update()
		h.display.Window.UpdateInput()
		time.Sleep(time.Millisecond * 16) // 约60 FPS
	}
	log.Println("Display loop ended")
}

func (h *ClientHandler) handleServerMessages() {
	log.Println("Started handling server messages")
	for h.running {
		msg, err := h.conn.Receive()
		if err != nil {
			log.Printf("Error receiving message from server: %v", err)
			h.running = false
			return
		}

		log.Printf("Received message type: %T", msg.Payload)

		switch payload := msg.Payload.(type) {
		case *protocol.Message_ScreenUpdate:
			log.Printf("Received screen update: %dx%d, %d bytes",
				payload.ScreenUpdate.Width,
				payload.ScreenUpdate.Height,
				len(payload.ScreenUpdate.ImageData))
			h.display.UpdateScreen(payload.ScreenUpdate)
		default:
			log.Printf("Unhandled message type: %T", payload)
		}
	}
}

func (h *ClientHandler) handleUserInput() {
	for h.running {
		event := h.input.NextEvent()
		if event == nil {
			time.Sleep(time.Millisecond * 16)
			continue
		}

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
				h.running = false
				return
			}
		}
	}
}
