// package client

// import (
// 	"bytes"
// 	"encoding/binary"
// 	"image"
// 	"image/color"
// 	"image/draw"
// 	"image/jpeg"
// 	"image/png"
// 	"io"
// 	"sync"

// 	"github.com/faiface/pixel"
// 	"github.com/faiface/pixel/pixelgl"
// 	"github.com/suifei/godesk/internal/protocol"
// 	"github.com/suifei/godesk/pkg/log"
// )

// type Display struct {
// 	Window     *pixelgl.Window
// 	sprite     *pixel.Sprite
// 	spriteRect pixel.Rect
// 	mutex      sync.Mutex
// 	fullImage  *image.RGBA
// 	canvas     *pixelgl.Canvas
// }

// func NewDisplay() *Display {
// 	cfg := pixelgl.WindowConfig{
// 		Title:     "GoDesk Client",
// 		Bounds:    pixel.R(0, 0, 800, 600),
// 		VSync:     true,
// 		Resizable: true,
// 	}
// 	win, err := pixelgl.NewWindow(cfg)
// 	if err != nil {
// 		log.Fatalf("Failed to create window: %v", err)
// 	}
// 	return &Display{
// 		Window: win,
// 		canvas: pixelgl.NewCanvas(pixel.R(0, 0, 800, 600)),
// 	}
// }

// func (d *Display) UpdateScreen(update *protocol.ScreenUpdate) {
// 	d.mutex.Lock()
// 	defer d.mutex.Unlock()

// 	if update == nil || len(update.ImageData) == 0 {
// 		log.Warnln("Received empty screen update")
// 		return
// 	}
// 	var img image.Image
// 	var err error

// 	switch update.CompressionType {
// 	case protocol.CompressionType_PNG:
// 		img, err = png.Decode(bytes.NewReader(update.ImageData))
// 	case protocol.CompressionType_JPEG:
// 		img, err = jpeg.Decode(bytes.NewReader(update.ImageData))
// 	case protocol.CompressionType_RLE:
// 		img, err = decodeRLE(bytes.NewReader(update.ImageData), int(update.Width), int(update.Height))
// 	default:
// 		log.Warnf("Unknown compression type: %v", update.CompressionType)
// 		return
// 	}
// 	if err != nil {
// 		log.Warnf("Error decoding screen update: %v", err)
// 		return
// 	}

// 	if !update.IsPartial {
// 		rgbaImg, ok := img.(*image.RGBA)
// 		if !ok {
// 			rgbaImg = imageToRGBA(img)
// 		}
// 		d.fullImage = rgbaImg
// 		d.updateSprite(d.fullImage)
// 		log.Debugf("Updated full image: %dx%d", update.Width, update.Height)
// 	} else {
// 		if d.fullImage == nil {
// 			log.Warnln("Received partial update before full image")
// 			return
// 		}
// 		draw.Draw(d.fullImage, image.Rect(int(update.X), int(update.Y), int(update.X+update.Width), int(update.Y+update.Height)),
// 			img, image.Point{}, draw.Src)
// 		d.updateSprite(d.fullImage)
// 		log.Debugf("Updated partial image: %dx%d at (%d,%d)", update.Width, update.Height, update.X, update.Y)
// 	}
// }
// func decodeRLE(r io.Reader, width, height int) (image.Image, error) {
// 	img := image.NewRGBA(image.Rect(0, 0, width, height))
// 	for y := 0; y < height; y++ {
// 		x := 0
// 		for x < width {
// 			var count byte
// 			if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
// 				return nil, err
// 			}
// 			var color color.RGBA
// 			if err := binary.Read(r, binary.LittleEndian, &color); err != nil {
// 				return nil, err
// 			}
// 			for i := 0; i < int(count) && x < width; i++ {
// 				img.Set(x, y, color)
// 				x++
// 			}
// 		}
// 	}
// 	return img, nil
// }
// func (d *Display) updateSprite(img *image.RGBA) {
// 	pic := pixel.PictureDataFromImage(img)
// 	d.sprite = pixel.NewSprite(pic, pic.Bounds())
// 	d.spriteRect = pic.Bounds()

// 	// Update canvas size if necessary
// 	if d.canvas.Bounds().W() != pic.Bounds().W() || d.canvas.Bounds().H() != pic.Bounds().H() {
// 		d.canvas = pixelgl.NewCanvas(pic.Bounds())
// 	}
// 	d.sprite.Draw(d.canvas, pixel.IM.Moved(pic.Bounds().Center()))
// }

// func (d *Display) Update() {
// 	d.mutex.Lock()
// 	defer d.mutex.Unlock()

// 	d.Window.Clear(pixel.RGB(0, 0, 0))
// 	if d.canvas != nil {
// 		scale := min(
// 			d.Window.Bounds().W()/d.canvas.Bounds().W(),
// 			d.Window.Bounds().H()/d.canvas.Bounds().H(),
// 		)
// 		transform := pixel.IM.Scaled(pixel.ZV, scale).Moved(d.Window.Bounds().Center())
// 		d.canvas.Draw(d.Window, transform)
// 		log.Debugln("Drew canvas to window")
// 	} else {
// 		log.Debugln("No canvas to draw")
// 	}
// 	d.Window.Update()
// }

// func (d *Display) ShouldClose() bool {
// 	return d.Window.Closed()
// }

// func imageToRGBA(img image.Image) *image.RGBA {
// 	bounds := img.Bounds()
// 	rgba := image.NewRGBA(bounds)
// 	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
// 	return rgba
// }

//	func min(a, b float64) float64 {
//		if a < b {
//			return a
//		}
//		return b
//	}
package client

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/suifei/godesk/pkg/log"
)

var (
	modGdi32 = syscall.NewLazyDLL("gdi32.dll")

	procCreateCompatibleDC     = modGdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = modGdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = modGdi32.NewProc("SelectObject")
	procBitBlt                 = modGdi32.NewProc("BitBlt")
	procDeleteObject           = modGdi32.NewProc("DeleteObject")
	procSetDIBitsToDevice      = modGdi32.NewProc("SetDIBitsToDevice")

	modKernel32      = syscall.NewLazyDLL("kernel32.dll")
	procGlobalAlloc  = modKernel32.NewProc("GlobalAlloc")
	procGlobalFree   = modKernel32.NewProc("GlobalFree")
	procGlobalLock   = modKernel32.NewProc("GlobalLock")
	procGlobalUnlock = modKernel32.NewProc("GlobalUnlock")
)

const (
	GMEM_FIXED    = 0x0000
	GMEM_MOVEABLE = 0x0002
	GMEM_ZEROINIT = 0x0040
	GMEM_MODIFY   = 0x0080
	GMEM_GHND     = GMEM_MOVEABLE | GMEM_ZEROINIT
	GMEM_GPTR     = GMEM_FIXED | GMEM_ZEROINIT
)

// HDC and HBITMAP types
type HDC uintptr
type HBITMAP uintptr
type HGLOBAL uintptr

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
}

const (
	BI_RGB       = 0
	BI_RLE8      = 1
	BI_RLE4      = 2
	BI_BITFIELDS = 3
	BI_JPEG      = 4
	BI_PNG       = 5
)
const (
	DIB_RGB_COLORS = 0
)

func CreateCompatibleDC(hdc HDC) HDC {
	ret, _, _ := procCreateCompatibleDC.Call(uintptr(hdc))
	return HDC(ret)
}

func CreateCompatibleBitmap(hdc HDC, width, height int32) HBITMAP {
	ret, _, _ := procCreateCompatibleBitmap.Call(uintptr(hdc), uintptr(width), uintptr(height))
	return HBITMAP(ret)
}

func SelectObject(hdc HDC, hObject uintptr) uintptr {
	ret, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(hObject))
	return ret
}

func BitBlt(hdcDest, xDest, yDest int32, width, height int32, hdcSrc HDC, xSrc, ySrc int32, dwRop uint32) bool {
	ret, _, _ := procBitBlt.Call(uintptr(hdcDest), uintptr(xDest), uintptr(yDest), uintptr(width), uintptr(height), uintptr(hdcSrc), uintptr(xSrc), uintptr(ySrc), uintptr(dwRop))
	return ret != 0
}

func DeleteObject(hObject uintptr) bool {
	ret, _, _ := procDeleteObject.Call(uintptr(hObject))
	return ret != 0
}

func SetDIBitsToDevice(hdc HDC, xDest, yDest, dwWidth, dwHeight, xSrc, ySrc uint32, uStartScan, cScanLines uint32, lpBits unsafe.Pointer, lpbi *BITMAPINFO, uColorUse uint32) bool {
	ret, _, _ := procSetDIBitsToDevice.Call(uintptr(hdc), uintptr(xDest), uintptr(yDest), uintptr(dwWidth), uintptr(dwHeight), uintptr(xSrc), uintptr(ySrc), uintptr(uStartScan), uintptr(cScanLines), uintptr(lpBits), uintptr(unsafe.Pointer(lpbi)), uintptr(uColorUse))
	return ret != 0
}

func GlobalAlloc(uFlags uint32, dwBytes uint32) HGLOBAL {
	ret, _, _ := procGlobalAlloc.Call(uintptr(uFlags), uintptr(dwBytes))
	return HGLOBAL(ret)
}

func GlobalFree(hMem HGLOBAL) bool {
	ret, _, _ := procGlobalFree.Call(uintptr(hMem))
	return ret != 0
}

func GlobalLock(hMem HGLOBAL) unsafe.Pointer {
	ret, _, _ := procGlobalLock.Call(uintptr(hMem))
	return unsafe.Pointer(ret)
}

func GlobalUnlock(hMem HGLOBAL) bool {
	ret, _, _ := procGlobalUnlock.Call(uintptr(hMem))
	return ret != 0
}

type Display struct {
	window    ui.WindowMain
	width     int
	height    int
	imageData *image.RGBA
	inputChan chan InputEvent
}

func NewDisplay(width, height int) (*Display, error) {
	d := &Display{
		width:     width,
		height:    height,
		inputChan: make(chan InputEvent, 100),
	}

	d.window = ui.NewWindowMain(
		ui.WindowMainOpts().
			Title("GoDesk Client").
			ClientArea(win.SIZE{Cx: int32(width), Cy: int32(height)}),
	)

	d.setupEventHandlers()

	return d, nil
}
func (d *Display) setupEventHandlers() {
	d.window.On().WmCreate(func(p wm.Create) int {
		log.Infoln("Window created")
		return 0
	})

	d.window.On().WmDestroy(func() {
		log.Infoln("Window destroyed")
		close(d.inputChan)
	})

	d.window.On().WmPaint(func() {
		if d.imageData != nil {
			// Implement drawing logic here
			log.Debugln("Drawing updated screen")
		}
	})

	d.window.On().WmKeyDown(func(p wm.Key) {
		d.inputChan <- InputEvent{
			Type:    KeyboardEvent,
			KeyCode: int(p.VirtualKeyCode()),
			Down:    true,
		}
		log.Infof("Key down: %d", p.VirtualKeyCode())
	})

	d.window.On().WmKeyUp(func(p wm.Key) {
		d.inputChan <- InputEvent{
			Type:    KeyboardEvent,
			KeyCode: int(p.VirtualKeyCode()),
			Down:    false,
		}
		log.Infof("Key up: %d", p.VirtualKeyCode())
	})

	d.window.On().WmLButtonDown(func(p wm.Mouse) {
		d.inputChan <- InputEvent{
			Type:   MouseEvent,
			X:      int(p.Pos().X),
			Y:      int(p.Pos().Y),
			Button: LeftButton,
			Down:   true,
		}
		log.Infof("Mouse down: %d, %d", p.Pos().X, p.Pos().Y)
	})

	d.window.On().WmLButtonUp(func(p wm.Mouse) {
		d.inputChan <- InputEvent{
			Type:   MouseEvent,
			X:      int(p.Pos().X),
			Y:      int(p.Pos().Y),
			Button: LeftButton,
			Down:   false,
		}
		log.Infof("Mouse up: %d, %d", p.Pos().X, p.Pos().Y)
	})

	d.window.On().WmRButtonDown(func(p wm.Mouse) {
		d.inputChan <- InputEvent{
			Type:   MouseEvent,
			X:      int(p.Pos().X),
			Y:      int(p.Pos().Y),
			Button: RightButton,
			Down:   true,
		}
		log.Infof("Mouse down: %d, %d", p.Pos().X, p.Pos().Y)
	})

	d.window.On().WmRButtonUp(func(p wm.Mouse) {
		d.inputChan <- InputEvent{
			Type:   MouseEvent,
			X:      int(p.Pos().X),
			Y:      int(p.Pos().Y),
			Button: RightButton,
			Down:   false,
		}
		log.Infof("Mouse up: %d, %d", p.Pos().X, p.Pos().Y)
	})

	d.window.On().WmMouseMove(func(p wm.Mouse) {
		d.inputChan <- InputEvent{
			Type:   MouseEvent,
			X:      int(p.Pos().X),
			Y:      int(p.Pos().Y),
			Button: NoButton,
		}
		log.Infof("Mouse move: %d, %d", p.Pos().X, p.Pos().Y)
	})
}

func (d *Display) Run() {
	runtime.LockOSThread()
	d.window.RunAsMain()
}

func (d *Display) Close() {
	// windigo handles window destruction automatically
}

func (d *Display) UpdateScreen(update *image.RGBA) {
	d.imageData = update

	// 将图像转换为 PNG 格式
	var buf bytes.Buffer
	if err := png.Encode(&buf, d.imageData); err != nil {
		log.Fatalf("图像编码为 PNG 失败: %v", err)
	}
	pngData := buf.Bytes()

	// 获取设备上下文 (DC)
	hdc := d.window.Hwnd().GetDC()
	defer hdc.DeleteDC()

	// 调用绘制 PNG 的逻辑
	d.drawPNG(HDC(hdc), pngData, d.imageData.Bounds().Dx(), d.imageData.Bounds().Dy())

	log.Debugf("Screen updated: %dx%d", d.imageData.Bounds().Dx(), d.imageData.Bounds().Dy())
}

func (d *Display) UpdatePartialScreen(update *image.RGBA, x, y int) {
	// 更新部分图像数据
	bounds := update.Bounds()
	if d.imageData == nil {
		// 如果还没有完整的图像，创建一个
		d.imageData = image.NewRGBA(image.Rect(0, 0, d.width, d.height))
	}
	draw.Draw(d.imageData, image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()), update, bounds.Min, draw.Src)

	// 将更新的部分转换为 PNG 格式
	var buf bytes.Buffer
	if err := png.Encode(&buf, update); err != nil {
		log.Fatalf("部分图像编码为 PNG 失败: %v", err)
	}
	pngData := buf.Bytes()

	// 获取设备上下文 (DC)
	hdc := d.window.Hwnd().GetDC()
	defer d.window.Hwnd().ReleaseDC(hdc)

	// 调用绘制 PNG 的逻辑
	d.drawPartialPNG(HDC(hdc), pngData, bounds.Dx(), bounds.Dy(), x, y)

	log.Debugf("Partial screen updated: %dx%d at (%d,%d)", bounds.Dx(), bounds.Dy(), x, y)
}
func (d *Display) drawPNG(hdc HDC, pngData []byte, width, height int) {
	// Decode PNG data to image
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		log.Fatalf("Failed to decode PNG data: %v", err)
	}

	// Create a compatible DC
	hdcMem := CreateCompatibleDC(hdc)
	defer DeleteObject(uintptr(hdcMem))

	// Create a compatible bitmap
	hBitmap := CreateCompatibleBitmap(hdc, int32(width), int32(height))
	defer DeleteObject(uintptr(hBitmap))

	// Select the bitmap into the compatible DC
	oldBitmap := SelectObject(hdcMem, uintptr(hBitmap))
	defer SelectObject(hdcMem, oldBitmap)

	// Prepare BITMAPINFO
	bi := BITMAPINFO{
		BmiHeader: BITMAPINFOHEADER{
			BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
			BiWidth:       int32(width),
			BiHeight:      -int32(height), // Top-down DIB
			BiPlanes:      1,
			BiBitCount:    32,
			BiCompression: 0,
		},
	}

	// Calculate the size of the bitmap data
	dataSize := uint32(width * height * 4) // 4 bytes per pixel (RGBA)
	hMem := GlobalAlloc(GMEM_MOVEABLE, dataSize)
	defer GlobalFree(hMem)

	// Lock the global memory
	memPtr := GlobalLock(hMem)
	defer GlobalUnlock(hMem)

	// Copy image data to memory block
	rgba := img.(*image.RGBA)
	copy((*[1 << 30]byte)(memPtr)[:dataSize], rgba.Pix)

	// Draw the bitmap to the device context
	SetDIBitsToDevice(hdc, 0, 0, uint32(width), uint32(height), 0, 0, 0, uint32(height), memPtr, &bi, DIB_RGB_COLORS)
}

func (d *Display) drawPartialPNG(hdc HDC, pngData []byte, width, height, x, y int) {
	// Decode PNG data to image
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		log.Fatalf("Failed to decode PNG data: %v", err)
	}

	// // Create a compatible DC
	// hdcMem := hdc.CreateCompatibleDC()
	// defer hdcMem.DeleteDC()

	// // Create a compatible bitmap
	// hBitmap := hdc.CreateCompatibleBitmap(int32(width), int32(height))
	// defer hBitmap.DeleteObject()

	// Select the bitmap into the compatible DC
	// oldBitmap := hdcMem.SelectObject(win.HGDIOBJ(hBitmap))
	// defer hdcMem.SelectObject(oldBitmap)

	// Create a compatible DC
	hdcMem := CreateCompatibleDC(hdc)
	defer DeleteObject(uintptr(hdcMem))

	// Create a compatible bitmap
	hBitmap := CreateCompatibleBitmap(hdc, int32(width), int32(height))
	defer DeleteObject(uintptr(hBitmap))

	// Select the bitmap into the compatible DC
	oldBitmap := SelectObject(hdcMem, uintptr(hBitmap))
	defer SelectObject(hdcMem, oldBitmap)

	// Prepare BITMAPINFO
	bi := BITMAPINFO{
		BmiHeader: BITMAPINFOHEADER{
			BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
			BiWidth:       int32(width),
			BiHeight:      -int32(height), // Top-down DIB
			BiPlanes:      1,
			BiBitCount:    32,
			BiCompression: BI_RGB,
		},
	}

	// Calculate the size of the bitmap data
	dataSize := uint32(width * height * 4) // 4 bytes per pixel (RGBA)

	// Allocate memory for bitmap data
	rawMem := GlobalAlloc(GMEM_FIXED|GMEM_ZEROINIT, dataSize)
	defer GlobalFree(rawMem)

	bmpSlice := GlobalLock(rawMem)
	defer GlobalUnlock(rawMem)

	// Copy image data to memory block
	rgba := img.(*image.RGBA)
	copy((*[1 << 30]byte)(bmpSlice)[:dataSize], rgba.Pix)

	// // Copy RGBA data to the allocated memory
	// rgba := img.(*image.RGBA)
	// for y := 0; y < height; y++ {
	// 	for x := 0; x < width; x++ {
	// 		i := y*width*4 + x*4
	// 		r, g, b, a := rgba.At(x, y).RGBA()
	// 		bmpSlice[i] = byte(b >> 8)
	// 		bmpSlice[i+1] = byte(g >> 8)
	// 		bmpSlice[i+2] = byte(r >> 8)
	// 		bmpSlice[i+3] = byte(a >> 8)
	// 	}
	// }

	// Draw the bitmap to the device context
	SetDIBitsToDevice(
		hdc,
		uint32(x), uint32(y), // Destination position
		uint32(width), uint32(height), // Width and height of destination area
		0, 0, // Source position
		0, uint32(height), // Scanlines to copy
		bmpSlice,
		&bi,
		DIB_RGB_COLORS,
	)

}
func (d *Display) InputEvents() <-chan InputEvent {
	return d.inputChan
}

// You might need to implement a custom drawing method
// This is just a placeholder and needs to be implemented properly
func (d *Display) draw() {
	if d.imageData == nil {
		return
	}

	// Implement drawing logic here
	// You might need to use GDI or Direct2D for efficient drawing
	log.Debugln("Drawing updated screen")
}
