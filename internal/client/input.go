package client

import (
	"github.com/faiface/pixel/pixelgl"
)

type InputHandler struct {
	window *pixelgl.Window
	events chan interface{}
}

type MouseEvent struct {
	EventType int
	X, Y      float64
}

type KeyEvent struct {
	EventType int
	KeyCode   int
}

func NewInputHandler(window *pixelgl.Window) *InputHandler {
	return &InputHandler{
		window: window,
		events: make(chan interface{}, 100),
	}
}

func (h *InputHandler) Start() {
	go h.pollEvents()
}

func (h *InputHandler) Events() <-chan interface{} {
	return h.events
}

func (h *InputHandler) pollEvents() {
	for !h.window.Closed() {
		if h.window.JustPressed(pixelgl.MouseButtonLeft) {
			pos := h.window.MousePosition()
			h.events <- MouseEvent{EventType: 1, X: pos.X, Y: pos.Y}
		}
		if h.window.JustReleased(pixelgl.MouseButtonLeft) {
			pos := h.window.MousePosition()
			h.events <- MouseEvent{EventType: 2, X: pos.X, Y: pos.Y}
		}
		// Add more mouse and keyboard event handling here
		h.window.Update()
	}
	close(h.events)
}
