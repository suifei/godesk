package client

import (
	"log"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/suifei/godesk/internal/protocol"
)

type MouseEvent struct {
	EventType protocol.MouseEvent_EventType
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
		// 添加鼠标事件处理
		// JustPressed 和 JustReleased 方法分别检查鼠标按钮是否刚刚被按下或释放
		// 如果是，则返回 true，否则返回 false

		// MouseEvent_MOVE        MouseEvent_EventType = 0
		// MouseEvent_LEFT_DOWN   MouseEvent_EventType = 1
		// MouseEvent_LEFT_UP     MouseEvent_EventType = 2
		// MouseEvent_RIGHT_DOWN  MouseEvent_EventType = 3
		// MouseEvent_RIGHT_UP    MouseEvent_EventType = 4
		// MouseEvent_MIDDLE_DOWN MouseEvent_EventType = 5
		// MouseEvent_MIDDLE_UP   MouseEvent_EventType = 6
		// MouseEvent_SCROLL      MouseEvent_EventType = 7

		// if h.window.JustMoved() {
		// 	pos := h.window.MousePosition()
		// 	scaledPos := h.scaleMousePosition(pos)
		// 	h.events <- MouseEvent{EventType: protocol.MouseEvent_MOVE, X: scaledPos.X, Y: scaledPos.Y}
		// }

		if h.window.JustPressed(pixelgl.MouseButtonMiddle) {
			pos := h.window.MousePosition()
			scaledPos := h.scaleMousePosition(pos)
			h.events <- MouseEvent{EventType: protocol.MouseEvent_MIDDLE_DOWN, X: scaledPos.X, Y: scaledPos.Y}
		}

		if h.window.JustReleased(pixelgl.MouseButtonMiddle) {
			pos := h.window.MousePosition()
			scaledPos := h.scaleMousePosition(pos)
			h.events <- MouseEvent{EventType: protocol.MouseEvent_MIDDLE_UP, X: scaledPos.X, Y: scaledPos.Y}
		}
		if h.window.JustReleased(pixelgl.MouseButtonLeft) {
			pos := h.window.MousePosition()
			scaledPos := h.scaleMousePosition(pos)
			h.events <- MouseEvent{EventType: protocol.MouseEvent_LEFT_DOWN, X: scaledPos.X, Y: scaledPos.Y}
		}
		if h.window.JustReleased(pixelgl.MouseButtonLeft) {
			pos := h.window.MousePosition()
			scaledPos := h.scaleMousePosition(pos)
			h.events <- MouseEvent{EventType: protocol.MouseEvent_LEFT_UP, X: scaledPos.X, Y: scaledPos.Y}
		}
		if h.window.JustReleased(pixelgl.MouseButtonRight) {
			pos := h.window.MousePosition()
			scaledPos := h.scaleMousePosition(pos)
			h.events <- MouseEvent{EventType: protocol.MouseEvent_RIGHT_DOWN, X: scaledPos.X, Y: scaledPos.Y}
		}
		if h.window.JustReleased(pixelgl.MouseButtonRight) {
			pos := h.window.MousePosition()
			scaledPos := h.scaleMousePosition(pos)
			h.events <- MouseEvent{EventType: protocol.MouseEvent_RIGHT_UP, X: scaledPos.X, Y: scaledPos.Y}
		}

		// 添加键盘事件处理
		for key := pixelgl.KeyUnknown; key <= pixelgl.KeyLast; key++ {
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

	log.Println("Mouse position:", pos, "Scaled position:", pixel.V(imageX, imageY))
	return pixel.V(imageX, imageY)
}
