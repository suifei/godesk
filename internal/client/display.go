package client

import (
	"bytes"
	"image/png"
	"log"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/suifei/godesk/internal/protocol"
)

type Display struct {
	Window     *pixelgl.Window
	sprite     *pixel.Sprite
	spriteRect pixel.Rect
	mutex      sync.Mutex
}

func NewDisplay() *Display {
	cfg := pixelgl.WindowConfig{
		Title:     "GoDesk Client",
		Bounds:    pixel.R(0, 0, 800, 600),
		VSync:     true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}
	return &Display{Window: win}
}

func (d *Display) UpdateScreen(update *protocol.ScreenUpdate) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if update == nil || len(update.ImageData) == 0 {
		log.Println("Received empty screen update")
		return
	}

	img, err := png.Decode(bytes.NewReader(update.ImageData))
	if err != nil {
		log.Printf("Error decoding screen update: %v", err)
		return
	}

	pic := pixel.PictureDataFromImage(img)
	d.sprite = pixel.NewSprite(pic, pic.Bounds())
	d.spriteRect = pic.Bounds()
	log.Printf("Updated sprite with new image: %dx%d", d.spriteRect.W(), d.spriteRect.H())
}

func (d *Display) Update() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.Window.Clear(pixel.RGB(0, 0, 0))
	if d.sprite != nil {
		// 计算缩放比例
		scale := min(
			d.Window.Bounds().W()/d.spriteRect.W(),
			d.Window.Bounds().H()/d.spriteRect.H(),
		)

		// 计算绘制位置，使图像居中
		pos := d.Window.Bounds().Center()

		// 绘制精灵
		d.sprite.Draw(d.Window, pixel.IM.Scaled(pixel.ZV, scale).Moved(pos))
		log.Println("Drew sprite to window")
	} else {
		log.Println("No sprite to draw")
	}
	d.Window.Update()
}

func (d *Display) ShouldClose() bool {
	return d.Window.Closed()
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
