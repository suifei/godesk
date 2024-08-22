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
	log.Println("Screen capturer started")
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
			log.Printf("Screen captured: %dx%d, %d bytes",
				update.Width, update.Height, len(update.ImageData))
			c.updates <- update
		case <-c.stop:
			log.Println("Screen capturer stopped")
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
		log.Printf("Error capturing screen: %v", err)
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("Error encoding screen: %v", err)
		return nil, err
	}

	update := &protocol.ScreenUpdate{
		Width:     int32(bounds.Dx()),
		Height:    int32(bounds.Dy()),
		ImageData: buf.Bytes(),
		Timestamp: time.Now().UnixNano(),
	}

	log.Printf("Captured screen: %dx%d", update.Width, update.Height)

	return update, nil
}

func (c *Capturer) CaptureScreen() (*protocol.ScreenUpdate, error) {
    return c.captureScreen()
}