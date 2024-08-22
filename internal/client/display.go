package client

import (
	"bytes"
	"image/png"
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/suifei/godesk/internal/protocol"
)

type Display struct {
	Window *pixelgl.Window // Changed from window to Window to make it public
}

func NewDisplay() *Display {
	cfg := pixelgl.WindowConfig{
		Title:  "GoDesk Client",
		Bounds: pixel.R(0, 0, 800, 600),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}
	return &Display{Window: win} // Changed from window to Window
}
func (d *Display) Update(update *protocol.ScreenUpdate) {
	img, err := png.Decode(bytes.NewReader(update.ImageData))
	if err != nil {
		log.Printf("Error decoding screen update: %v", err)
		return
	}

	bounds := img.Bounds()
	pic := pixel.PictureDataFromImage(img)
	sprite := pixel.NewSprite(pic, pic.Bounds())

	d.Window.Clear(pixel.RGB(0, 0, 0))
	sprite.Draw(d.Window, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			float64(d.Window.Bounds().W())/float64(bounds.Dx()),
			float64(d.Window.Bounds().H())/float64(bounds.Dy()),
		)).
		Moved(d.Window.Bounds().Center()))
	d.Window.Update() // Changed from window to Window
}

func (d *Display) ShouldClose() bool {
	return d.Window.Closed() // Changed from window to Window
}
