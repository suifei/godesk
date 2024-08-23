package server

import (
	"time"

	"github.com/suifei/godesk/pkg/log"

	"github.com/suifei/godesk/internal/protocol"
	"github.com/suifei/godesk/pkg/network"
)

type ClientHandler struct {
	conn     *network.TCPConnection
	capturer *Capturer
}

func NewClientHandler(conn *network.TCPConnection) *ClientHandler {
	return &ClientHandler{
		conn:     conn,
		capturer: NewCapturer(100 * time.Millisecond), // 每100ms捕获一次屏幕
	}
}

func (h *ClientHandler) Handle() {
	log.Infoln("New client connected, starting handler")

	// 开始屏幕捕获
	go h.capturer.Start()
	defer h.capturer.Stop()

	// 处理来自客户端的消息
	go h.handleIncomingMessages()

	// 发送屏幕更新到客户端
	h.sendScreenUpdates()
}

func (h *ClientHandler) handleIncomingMessages() {
	log.Infoln("Starting to handle incoming messages")
	for {
		msg, err := h.conn.Receive()
		if err != nil {
			log.Errorf("Error receiving message: %v", err)
			return
		}

		switch payload := msg.Payload.(type) {
		case *protocol.Message_InputEvent:
			log.Debugf("Received input event: %v", payload.InputEvent)
			HandleInputEvent(payload.InputEvent)
		default:
			log.Warnf("Received unknown message type: %T", payload)
		}
	}
}

func (h *ClientHandler) sendScreenUpdates() {
	log.Infoln("Starting to send screen updates")
	for update := range h.capturer.Updates() {
		log.Debugf("Sending screen update: %dx%d, %d bytes",
			update.Width, update.Height, len(update.ImageData))
		msg := &protocol.Message{
			Payload: &protocol.Message_ScreenUpdate{
				ScreenUpdate: update,
			},
		}
		if err := h.conn.Send(msg); err != nil {
			log.Errorf("Error sending screen update: %v", err)
			return
		}
	}
}
