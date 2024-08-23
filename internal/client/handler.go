// package client

// import (
// 	"net"
// 	"time"

// 	"github.com/suifei/godesk/pkg/log"

// 	"github.com/suifei/godesk/internal/protocol"
// 	"github.com/suifei/godesk/pkg/network"
// )

// type ClientHandler struct {
// 	conn    *network.TCPConnection
// 	display *Display
// 	input   *InputHandler
// 	running bool
// }

// func NewClientHandler(serverAddr string) (*ClientHandler, error) {
// 	log.Debugf("Attempting to connect to %s", serverAddr)
// 	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
// 	if err != nil {
// 		log.Errorf("Failed to connect to server: %v", err)
// 		return nil, err
// 	}
// 	log.Debugf("Successfully connected to %s", serverAddr)

// 	tcpConn := network.NewTCPConnection(conn)
// 	display := NewDisplay()

// 	input := NewInputHandler(display.Window, display, tcpConn)

// 	return &ClientHandler{
// 		conn:    tcpConn,
// 		display: display,
// 		input:   input,
// 		running: true,
// 	}, nil
// }

// func (h *ClientHandler) Handle() {
// 	defer h.conn.Close()

// 	h.input.Start()
// 	go h.handleServerMessages()

// 	log.Debugf("Starting main display loop")
// 	for h.running && !h.display.ShouldClose() {
// 		h.display.Update()
// 		h.display.Window.UpdateInput()
// 		time.Sleep(time.Millisecond * 16) // 约60 FPS
// 	}
// 	log.Debugf("Display loop ended")
// }

// func (h *ClientHandler) handleServerMessages() {
// 	log.Debugf("Started handling server messages")
// 	for h.running {
// 		msg, err := h.conn.Receive()
// 		if err != nil {
// 			log.Errorf("Error receiving message from server: %v", err)
// 			h.running = false
// 			return
// 		}

// 		log.Debugf("Received message type: %T", msg.Payload)

//			switch payload := msg.Payload.(type) {
//			case *protocol.Message_ScreenUpdate:
//				log.Debugf("Received screen update: %dx%d, %d bytes",
//					payload.ScreenUpdate.Width,
//					payload.ScreenUpdate.Height,
//					len(payload.ScreenUpdate.ImageData))
//				h.display.UpdateScreen(payload.ScreenUpdate)
//			default:
//				log.Debugf("Unhandled message type: %T", payload)
//			}
//		}
//	}
package client

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"net"
	"time"

	"github.com/suifei/godesk/internal/protocol"
	"github.com/suifei/godesk/pkg/log"
	"github.com/suifei/godesk/pkg/network"
)

type ClientHandler struct {
	conn    *network.TCPConnection
	display *Display
	input   *InputHandler
	running bool
}

func NewClientHandler(serverAddr string) (*ClientHandler, error) {
	log.Debugf("Attempting to connect to %s", serverAddr)
	tcpconn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		log.Errorf("Failed to connect to server: %v", err)
		return nil, err
	}
	log.Debugf("Successfully connected to %s", serverAddr)

	conn := network.NewTCPConnection(tcpconn)
	if err != nil {
		log.Errorf("Failed to connect to server: %v", err)
		return nil, err
	}
	log.Debugf("Successfully connected to %s", serverAddr)

	display, err := NewDisplay(800, 600) // You can adjust these dimensions
	if err != nil {
		log.Errorf("Failed to create display: %v", err)
		conn.Close()
		return nil, err
	}

	handler := &ClientHandler{
		conn:    conn,
		display: display,
		running: true,
	}

	handler.input = NewInputHandler(display)

	return handler, nil
}

func (h *ClientHandler) Handle() {
	defer h.conn.Close()

	h.input.Start()
	go h.handleServerMessages()
	go h.handleInputEvents()

	h.display.Run() // This will block until the window is closed
}

func (h *ClientHandler) handleServerMessages() {
	for h.running {
		msg, err := h.conn.Receive()
		if err != nil {
			log.Errorf("Error receiving message from server: %v", err)
			h.running = false
			h.display.Close()
			return
		}

		switch payload := msg.Payload.(type) {
		case *protocol.Message_ScreenUpdate:
			h.handleScreenUpdate(payload.ScreenUpdate)
		default:
			log.Warnf("Received unknown message type: %T", payload)
		}
	}
}

func (h *ClientHandler) handleScreenUpdate(update *protocol.ScreenUpdate) {
	log.Debugf("Received screen update: %dx%d, %d bytes",
		update.Width, update.Height, len(update.ImageData))

	var img image.Image
	var err error

	// 尝试解码图像数据
	if update.CompressionType == protocol.CompressionType_PNG {
		img, err = png.Decode(bytes.NewReader(update.ImageData))
	} else if update.CompressionType == protocol.CompressionType_JPEG {
		img, err = jpeg.Decode(bytes.NewReader(update.ImageData))
	} else {
		// 假设是原始 RGBA 数据
		expectedSize := int(update.Width) * int(update.Height) * 4 // 4 bytes per pixel (RGBA)
		if len(update.ImageData) != expectedSize {
			log.Errorf("Received data size (%d) does not match expected size (%d)", len(update.ImageData), expectedSize)
			return
		}
		img = &image.RGBA{
			Pix:    update.ImageData,
			Stride: int(update.Width) * 4,
			Rect:   image.Rect(0, 0, int(update.Width), int(update.Height)),
		}
	}

	if err != nil {
		log.Errorf("Failed to decode image data: %v", err)
		return
	}

	// 转换为 RGBA 格式（如果不是的话）
	bounds := img.Bounds()
	rgbaImg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgbaImg.Set(x, y, img.At(x, y))
		}
	}

	// 更新屏幕
	if update.IsPartial {
		h.display.UpdatePartialScreen(rgbaImg, int(update.X), int(update.Y))
	} else {
		h.display.UpdateScreen(rgbaImg)
	}
}

// If your image data is encoded (e.g., PNG or JPEG), you would use this function instead:
func decodeImage(data []byte) (*image.RGBA, error) {
	// Decode the image data
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Convert to RGBA if it's not already
	bounds := img.Bounds()
	rgbaImg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgbaImg.Set(x, y, img.At(x, y))
		}
	}

	return rgbaImg, nil
}
func (h *ClientHandler) handleInputEvents() {
	for h.running {
		event := h.input.NextEvent()
		if event == nil {
			time.Sleep(time.Millisecond * 16) // Avoid busy-waiting
			continue
		}

		msg := &protocol.Message{
			Payload: &protocol.Message_InputEvent{
				InputEvent: event,
			},
		}

		if err := h.conn.Send(msg); err != nil {
			log.Errorf("Error sending input event to server: %v", err)
			h.running = false
			h.display.Close()
			return
		}
	}
}
