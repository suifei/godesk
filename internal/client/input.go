package client

import (
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type MouseEvent struct {
	EventType int
	X, Y      float64
}

type KeyEvent struct {
	EventType int
	KeyCode   pixelgl.Button
}

type InputHandler struct {
	window  *pixelgl.Window
	events  chan interface{}
	display *Display // 添加对 Display 的引用
	mutex   sync.Mutex
}

func NewInputHandler(window *pixelgl.Window, display *Display) *InputHandler {
	return &InputHandler{
		window:  window,
		events:  make(chan interface{}, 100),
		display: display,
	}
}

func (h *InputHandler) Start() {
	go h.pollEvents()
}

func (h *InputHandler) NextEvent() interface{} {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	select {
	case event := <-h.events:
		return event
	default:
		return nil
	}
}

func (h *InputHandler) pollEvents() {
	for !h.window.Closed() {
		h.mutex.Lock()
		if h.window.JustPressed(pixelgl.MouseButtonLeft) {
			pos := h.window.MousePosition()
			scaledPos := h.scaleMousePosition(pos)
			h.events <- MouseEvent{EventType: 1, X: scaledPos.X, Y: scaledPos.Y}
		}
		if h.window.JustReleased(pixelgl.MouseButtonLeft) {
			pos := h.window.MousePosition()
			scaledPos := h.scaleMousePosition(pos)
			h.events <- MouseEvent{EventType: 2, X: scaledPos.X, Y: scaledPos.Y}
		}

		// 添加键盘事件处理
		for key := pixelgl.Key0; key <= pixelgl.KeyWorld2; key++ {
			if h.window.JustPressed(key) {
				h.events <- KeyEvent{EventType: 1, KeyCode: key}
			}
			if h.window.JustReleased(key) {
				h.events <- KeyEvent{EventType: 2, KeyCode: key}
			}
		}

		h.mutex.Unlock()
		h.window.UpdateInput()
	}
	close(h.events)
}

func (h *InputHandler) scaleMousePosition(pos pixel.Vec) pixel.Vec {
	scale := min(
		h.window.Bounds().W()/h.display.spriteRect.W(),
		h.window.Bounds().H()/h.display.spriteRect.H(),
	)

	// 计算图像在窗口中的实际位置
	imagePos := h.window.Bounds().Center().Sub(h.display.spriteRect.Center().Scaled(scale))

	// 将鼠标位置从窗口坐标转换为图像坐标
	imageX := (pos.X - imagePos.X) / scale
	imageY := (pos.Y - imagePos.Y) / scale

	return pixel.V(imageX, imageY)
}
