package server

import (
	"log"
	"time"

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
	log.Println("New client connected")

	// 开始屏幕捕获
	go h.capturer.Start()
	defer h.capturer.Stop()

	// 处理来自客户端的消息
	go h.handleIncomingMessages()

	// 发送屏幕更新到客户端
	h.sendScreenUpdates()
}

func (h *ClientHandler) handleIncomingMessages() {
	for {
		msg, err := h.conn.Receive()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return
		}

		switch payload := msg.Payload.(type) {
		case *protocol.Message_MouseEvent:
			h.handleMouseEvent(payload.MouseEvent)
		case *protocol.Message_KeyEvent:
			h.handleKeyEvent(payload.KeyEvent)
		case *protocol.Message_FileTransferRequest:
			h.handleFileTransferRequest(payload.FileTransferRequest)
		default:
			log.Printf("Unhandled message type: %T", payload)
		}
	}
}

func (h *ClientHandler) sendScreenUpdates() {
	for update := range h.capturer.Updates() {
		msg := &protocol.Message{
			Payload: &protocol.Message_ScreenUpdate{
				ScreenUpdate: update,
			},
		}
		if err := h.conn.Send(msg); err != nil {
			log.Printf("Error sending screen update: %v", err)
			return
		}
	}
}

func (h *ClientHandler) handleMouseEvent(event *protocol.MouseEvent) {
	// 实现鼠标事件处理
	log.Printf("Received mouse event: %v", event)
}

func (h *ClientHandler) handleKeyEvent(event *protocol.KeyEvent) {
	// 实现键盘事件处理
	log.Printf("Received key event: %v", event)
}

func (h *ClientHandler) handleFileTransferRequest(request *protocol.FileTransferRequest) {
	// 实现文件传输请求处理
	log.Printf("Received file transfer request: %v", request)
}
