package server

import (
	"bytes"
	"image/png"
	"log"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/suifei/godesk/internal/protocol"
)

type Capturer struct {
	interval time.Duration
	updates  chan *protocol.ScreenUpdate
	stop     chan struct{}
}

func NewCapturer(interval time.Duration) *Capturer {
	return &Capturer{
		interval: interval,
		updates:  make(chan *protocol.ScreenUpdate),
		stop:     make(chan struct{}),
	}
}

func (c *Capturer) Start() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			update, err := c.captureScreen()
			if err != nil {
				log.Printf("Error capturing screen: %v", err)
				continue
			}
			c.updates <- update
		case <-c.stop:
			return
		}
	}
}

func (c *Capturer) Stop() {
	close(c.stop)
}

func (c *Capturer) Updates() <-chan *protocol.ScreenUpdate {
	return c.updates
}

func (c *Capturer) captureScreen() (*protocol.ScreenUpdate, error) {
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}

	// 将图像编码为PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return &protocol.ScreenUpdate{
		Width:     int32(bounds.Dx()),
		Height:    int32(bounds.Dy()),
		ImageData: buf.Bytes(),
		Timestamp: time.Now().UnixNano(),
	}, nil
}
