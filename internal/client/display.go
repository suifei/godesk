package client

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/suifei/godesk/internal/protocol"
	"github.com/suifei/godesk/pkg/log"
	"golang.org/x/sys/windows"
)

const (
	BITSPIXEL = 12
	HORZRES   = 8
	VERTRES   = 10
	NUMCOLORS = 24
	SRCCOPY   = 0x00CC0020
)

// RECT 结构体定义
type RECT struct {
	Left, Top, Right, Bottom int32
}
type POINT struct {
	X, Y int32
}

type CURSORINFO struct {
	CbSize      uint32
	Flags       uint32
	HCursor     syscall.Handle
	PtScreenPos POINT
}

const CURSOR_SHOWING = 0x00000001

var (
	modGdi32          = syscall.NewLazyDLL("gdi32.dll")
	moduser32         = windows.NewLazySystemDLL("user32.dll")
	procGetCursorInfo = moduser32.NewProc("GetCursorInfo")
	procSetCursor     = moduser32.NewProc("SetCursor")

	procStretchBlt             = modGdi32.NewProc("StretchBlt")
	procGetClientRect          = moduser32.NewProc("GetClientRect")
	procGetDeviceCaps          = modGdi32.NewProc("GetDeviceCaps")
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

// StretchBlt 函数声明
func StretchBlt(hdcDest HDC, nXOriginDest, nYOriginDest, nWidthDest, nHeightDest int32,
	hdcSrc HDC, nXOriginSrc, nYOriginSrc, nWidthSrc, nHeightSrc int32, dwRop uint32) error {
	ret, _, err := procStretchBlt.Call(
		uintptr(hdcDest),
		uintptr(nXOriginDest),
		uintptr(nYOriginDest),
		uintptr(nWidthDest),
		uintptr(nHeightDest),
		uintptr(hdcSrc),
		uintptr(nXOriginSrc),
		uintptr(nYOriginSrc),
		uintptr(nWidthSrc),
		uintptr(nHeightSrc),
		uintptr(dwRop),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// GetClientRect 函数声明
func GetClientRect(hWnd uintptr, lpRect *RECT) error {
	ret, _, err := procGetClientRect.Call(
		uintptr(hWnd),
		uintptr(unsafe.Pointer(lpRect)),
	)
	if ret == 0 {
		return err
	}
	return nil
}

func GetCursorInfo() (*CURSORINFO, error) {
	var ci CURSORINFO
	ci.CbSize = uint32(unsafe.Sizeof(ci))
	ret, _, err := procGetCursorInfo.Call(uintptr(unsafe.Pointer(&ci)))
	if ret == 0 {
		return nil, err
	}
	return &ci, nil
}

func SetCursor(hCursor syscall.Handle) syscall.Handle {
	log.Infof("Setting cursor: %v", hCursor)
	ret, _, err := procSetCursor.Call(uintptr(hCursor))
	if ret == 0 {
		log.Error(err)
	}
	return syscall.Handle(ret)
}

func GetDeviceCaps(hdc HDC, index int) int {
	ret, _, _ := procGetDeviceCaps.Call(uintptr(hdc), uintptr(index))
	return int(ret)
}

func CreateCompatibleDC(hdc HDC) (HDC, error) {
	ret, _, err := procCreateCompatibleDC.Call(uintptr(hdc))
	if ret == 0 {
		return 0, err
	}
	return HDC(ret), nil
}

func CreateCompatibleBitmap(hdc HDC, width, height int32) (HBITMAP, error) {
	ret, _, err := procCreateCompatibleBitmap.Call(uintptr(hdc), uintptr(width), uintptr(height))
	if ret == 0 {
		return 0, err
	}
	return HBITMAP(ret), nil
}

func SelectObject(hdc HDC, hObject uintptr) (uintptr, error) {
	ret, _, err := procSelectObject.Call(uintptr(hdc), uintptr(hObject))
	if ret == 0 {
		return 0, err
	}
	return ret, nil
}

func BitBlt(hdcDest, xDest, yDest int32, width, height int32, hdcSrc HDC, xSrc, ySrc int32, dwRop uint32) error {
	ret, _, err := procBitBlt.Call(uintptr(hdcDest), uintptr(xDest), uintptr(yDest), uintptr(width), uintptr(height), uintptr(hdcSrc), uintptr(xSrc), uintptr(ySrc), uintptr(dwRop))
	if ret == 0 {
		return err
	}
	return nil
}

func DeleteObject(hObject uintptr) error {
	ret, _, err := procDeleteObject.Call(uintptr(hObject))
	if ret == 0 {
		return err
	}
	return nil
}

func SetDIBitsToDevice(hdc HDC, xDest, yDest, dwWidth, dwHeight, xSrc, ySrc uint32, uStartScan, cScanLines uint32, lpBits unsafe.Pointer, lpbi *BITMAPINFO, uColorUse uint32) error {
	ret, _, err := procSetDIBitsToDevice.Call(uintptr(hdc), uintptr(xDest), uintptr(yDest), uintptr(dwWidth), uintptr(dwHeight), uintptr(xSrc), uintptr(ySrc), uintptr(uStartScan), uintptr(cScanLines), uintptr(lpBits), uintptr(unsafe.Pointer(lpbi)), uintptr(uColorUse))
	if ret == 0 {
		return err
	}
	return nil
}

func GlobalAlloc(uFlags uint32, dwBytes uint32) (HGLOBAL, error) {
	ret, _, err := procGlobalAlloc.Call(uintptr(uFlags), uintptr(dwBytes))
	if ret == 0 {
		return 0, err
	}
	return HGLOBAL(ret), nil
}

func GlobalFree(hMem HGLOBAL) error {
	ret, _, err := procGlobalFree.Call(uintptr(hMem))
	if ret == 0 {
		return err
	}
	return nil
}

func GlobalLock(hMem HGLOBAL) (unsafe.Pointer, error) {
	ret, _, err := procGlobalLock.Call(uintptr(hMem))
	if ret == 0 {
		return nil, err
	}
	return unsafe.Pointer(ret), nil
}

func GlobalUnlock(hMem HGLOBAL) error {
	ret, _, err := procGlobalUnlock.Call(uintptr(hMem))
	if ret == 0 {
		return err
	}
	return nil
}

type Display struct {
	window     ui.WindowMain
	width      int
	height     int
	imageData  *image.RGBA
	inputChan  chan InputEvent
}

func NewDisplay(width, height int) (*Display, error) {
	d := &Display{
		width:      width,
		height:     height,
		inputChan:  make(chan InputEvent, 100),
	}

	d.window = ui.NewWindowMain(
		ui.WindowMainOpts().
			Title("GoDesk Client").
			ClientArea(win.SIZE{Cx: int32(width), Cy: int32(height)}).
			WndStyles(co.WS_CAPTION | co.WS_SYSMENU | co.WS_CLIPCHILDREN |
				co.WS_BORDER | co.WS_VISIBLE | co.WS_MINIMIZEBOX |
				co.WS_MAXIMIZEBOX | co.WS_SIZEBOX),
	)

	d.setupEventHandlers()

	return d, nil
}

var isMouseDown = false

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

	d.window.On().WmSysKeyDown(func(p wm.Key) {
		d.HandleKeyEvent(p.Msg.WParam, true)
	})
	d.window.On().WmKeyDown(func(p wm.Key) {
		d.HandleKeyEvent(p.Msg.WParam, true)
	})

	d.window.On().WmGetDlgCode(func(p wm.GetDlgCode) co.DLGC {
		return co.DLGC_WANTALLKEYS
	})

	d.window.On().WmSysKeyUp(func(p wm.Key) {
		d.HandleKeyEvent(p.Msg.WParam, false)
	})
	d.window.On().WmKeyUp(func(p wm.Key) {
		d.HandleKeyEvent(p.Msg.WParam, false)
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
		isMouseDown = true
	})

	d.window.On().WmLButtonUp(func(p wm.Mouse) {
		if isMouseDown {
			d.inputChan <- InputEvent{
				Type:   MouseEvent,
				X:      int(p.Pos().X),
				Y:      int(p.Pos().Y),
				Button: LeftButton,
				Down:   false,
			}
			log.Infof("Mouse up: %d, %d", p.Pos().X, p.Pos().Y)
			isMouseDown = false
		}
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
		isMouseDown = true
	})

	d.window.On().WmRButtonUp(func(p wm.Mouse) {
		if isMouseDown {
			d.inputChan <- InputEvent{
				Type:   MouseEvent,
				X:      int(p.Pos().X),
				Y:      int(p.Pos().Y),
				Button: RightButton,
				Down:   false,
			}
			log.Infof("Mouse up: %d, %d", p.Pos().X, p.Pos().Y)
			isMouseDown = false
		}
	})

	d.window.On().WmMButtonDown(func(p wm.Mouse) {
		d.inputChan <- InputEvent{
			Type:   MouseEvent,
			X:      int(p.Pos().X),
			Y:      int(p.Pos().Y),
			Button: MiddleButton,
			Down:   true,
		}
		log.Infof("Mouse down: %d, %d", p.Pos().X, p.Pos().Y)
		isMouseDown = true
	})

	d.window.On().WmMButtonUp(func(p wm.Mouse) {
		if isMouseDown {
			d.inputChan <- InputEvent{
				Type:   MouseEvent,
				X:      int(p.Pos().X),
				Y:      int(p.Pos().Y),
				Button: MiddleButton,
				Down:   false,
			}
			log.Infof("Mouse up: %d, %d", p.Pos().X, p.Pos().Y)
			isMouseDown = false
		}
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

	d.window.On().WmLButtonDblClk(func(p wm.Mouse) {
		d.inputChan <- InputEvent{
			Type:   MouseEvent,
			X:      int(p.Pos().X),
			Y:      int(p.Pos().Y),
			Button: LeftButtonDbClick,
		}
		log.Infof("Left button double click")
	})
	d.window.On().WmRButtonDblClk(func(p wm.Mouse) {
		d.inputChan <- InputEvent{
			Type:   MouseEvent,
			X:      int(p.Pos().X),
			Y:      int(p.Pos().Y),
			Button: RightButtonDbClick,
		}
		log.Infof("Right button double click")
	})
	d.window.On().WmMButtonDblClk(func(p wm.Mouse) {
		d.inputChan <- InputEvent{
			Type:   MouseEvent,
			X:      int(p.Pos().X),
			Y:      int(p.Pos().Y),
			Button: MiddleButtonDbClick,
		}
		log.Infof("Middle button double click")
	})
	const WM_MOUSEWHEEL = 0x020A
	d.window.On().Wm(WM_MOUSEWHEEL, func(p wm.Any) uintptr {
		scrollDelta := int16(p.WParam.HiWord())
		d.inputChan <- InputEvent{
			Type:        MouseEvent,
			Button:      Scroll,
			ScrollDelta: int(scrollDelta),
		}
		log.Infof("Mouse wheel: %d", scrollDelta)
		return 0
	})

}

func (d *Display) isKeyPressed(vKey co.VK) bool {
	state := win.GetAsyncKeyState(vKey)
	return (state & 0x8000) != 0
}
func (d *Display) HandleKeyEvent(wParam win.WPARAM, isKeyDown bool) {
	keyCode := int(wParam)
	// 创建 KeyEvent 消息
	keyEvent := InputEvent{
		Type:    KeyboardEvent,
		KeyCode: keyCode,
		Down:    isKeyDown,
		Shift:   d.isKeyPressed(co.VK_SHIFT),
		Ctrl:    d.isKeyPressed(co.VK_CONTROL),
		Alt:     d.isKeyPressed(co.VK_MENU),
		Meta:    d.isKeyPressed(co.VK_LWIN) || d.isKeyPressed(co.VK_RWIN),
	}

	d.inputChan <- keyEvent

	if isKeyDown {
		log.Infof("Key down: %d, shift:%v ctrl:%v alt:%v meta:%v ", keyCode, keyEvent.Shift, keyEvent.Ctrl, keyEvent.Alt, keyEvent.Meta)
	} else {
		log.Infof("Key up: %d, shift:%v ctrl:%v alt:%v meta:%v ", keyCode, keyEvent.Shift, keyEvent.Ctrl, keyEvent.Alt, keyEvent.Meta)
	}
}
func (d *Display) Run() {
	runtime.LockOSThread()
	d.window.RunAsMain()
}

func (d *Display) Close() {
	// windigo handles window destruction automatically
}

func (d *Display) InputEvents() <-chan InputEvent {
	return d.inputChan
}

func (d *Display) UpdateScreen(update *protocol.ScreenUpdate, rgba *image.RGBA, x, y, screenWidth, screenHeight int) {
	d.imageData = rgba

	// 获取设备上下文 (DC)
	hdc := d.window.Hwnd().GetDC()
	defer d.window.Hwnd().ReleaseDC(hdc)

	// 检查设备上下文的设置
	checkDCSettings(HDC(hdc))

	// 设置光标
	// if update.Cursor != nil {
	// 	hcursor := update.Cursor.HCursor
	// 	SetCursor(syscall.Handle(hcursor))
	// }

	// 调用绘制逻辑
	if err := d.drawImage(HDC(hdc), d.imageData, d.imageData.Bounds().Dx(), d.imageData.Bounds().Dy(), x, y, screenWidth, screenHeight); err != nil {
		log.Errorf("Failed to draw image: %v", err)
		return
	}
	log.Debugf("Screen updated: %dx%d at (%d,%d), remote(%d,%d)", d.imageData.Bounds().Dx(), d.imageData.Bounds().Dy(), x, y, screenWidth, screenHeight)
}
func (d *Display) drawImage(hdc HDC, rgba *image.RGBA, width, height, x, y, screenWidth, screenHeight int) error {
	// 获取窗口的客户区大小
	var rect RECT
	GetClientRect(uintptr(d.window.Hwnd()), &rect)
	clientWidth := int(rect.Right - rect.Left)
	clientHeight := int(rect.Bottom - rect.Top)

	// 完整远程桌面的尺寸
	fullWidth := screenWidth
	fullHeight := screenHeight

	// 计算缩放比例
	scaleX := float64(clientWidth) / float64(fullWidth)
	scaleY := float64(clientHeight) / float64(fullHeight)
	scale := math.Min(scaleX, scaleY) // 使用较小的缩放比例以保持宽高比

	// 计算缩放后的完整图像尺寸
	scaledFullWidth := int(float64(fullWidth) * scale)
	scaledFullHeight := int(float64(fullHeight) * scale)

	// 计算居中偏移
	offsetX := (clientWidth - scaledFullWidth) / 2
	offsetY := (clientHeight - scaledFullHeight) / 2

	// 计算局部更新区域在缩放后的位置和尺寸
	scaledX := int(float64(x)*scale) + offsetX
	scaledY := int(float64(y)*scale) + offsetY
	scaledWidth := int(float64(width) * scale)
	scaledHeight := int(float64(height) * scale)

	fmt.Println("clientWidth:", clientWidth, "clientHeight:", clientHeight, "fullWidth:", fullWidth, "fullHeight:", fullHeight)
	fmt.Println ( "scaledX:", scaledX, "scaledY:", scaledY, "scaledWidth:", scaledWidth, "scaledHeight:", scaledHeight)
	// 创建兼容的DC和位图
	hdcMem, err := CreateCompatibleDC(hdc)
	if err != nil {
		return fmt.Errorf("failed to create compatible DC: %v", err)
	}
	defer DeleteObject(uintptr(hdcMem))

	hBitmap, err := CreateCompatibleBitmap(hdc, int32(width), int32(height))
	if err != nil {
		return fmt.Errorf("failed to create compatible bitmap: %v", err)
	}
	defer DeleteObject(uintptr(hBitmap))

	oldBitmap, err := SelectObject(hdcMem, uintptr(hBitmap))
	if err != nil {
		return fmt.Errorf("failed to select bitmap into DC: %v", err)
	}
	defer SelectObject(hdcMem, oldBitmap)

	// 准备BITMAPINFO
	bi := BITMAPINFO{
		BmiHeader: BITMAPINFOHEADER{
			BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
			BiWidth:       int32(width),
			BiHeight:      -int32(height),
			BiPlanes:      1,
			BiBitCount:    32,
			BiCompression: BI_RGB,
		},
	}

	// 将图像数据复制到位图
	if err := SetDIBitsToDevice(
		hdcMem,
		0, 0,
		uint32(width), uint32(height),
		0, 0,
		0, uint32(height),
		unsafe.Pointer(&rgba.Pix[0]),
		&bi,
		DIB_RGB_COLORS,
	); err != nil {
		return fmt.Errorf("failed to set DIB bits: %v", err)
	}

	// 使用StretchBlt函数进行缩放绘制

	if err := StretchBlt(
		hdc,
		int32(scaledX), int32(scaledY),
		int32(scaledWidth), int32(scaledHeight),
		hdcMem,
		0, 0,
		int32(width), int32(height),
		SRCCOPY,
	); err != nil {
		return fmt.Errorf("failed to stretch blit: %v", err)
	}

	return nil
}

// func (d *Display) drawImage(hdc HDC, rgba *image.RGBA, width, height, x, y int) error {
// 	// Check image dimensions
// 	if rgba.Bounds().Dx() != width || rgba.Bounds().Dy() != height {
// 		return fmt.Errorf("Image dimensions do not match: expected %dx%d, got %dx%d", width, height, rgba.Bounds().Dx(), rgba.Bounds().Dy())
// 	}

// 	// Create a compatible DC
// 	hdcMem, err := CreateCompatibleDC(hdc)
// 	if err != nil {
// 		return fmt.Errorf("failed to create compatible DC: %v", err)
// 	}
// 	defer DeleteObject(uintptr(hdcMem))

// 	// Create a compatible bitmap
// 	hBitmap, err := CreateCompatibleBitmap(hdc, int32(width), int32(height))
// 	if err != nil {
// 		return fmt.Errorf("failed to create compatible bitmap: %v", err)
// 	}
// 	defer DeleteObject(uintptr(hBitmap))

// 	// Select the bitmap into the compatible DC
// 	oldBitmap, err := SelectObject(hdcMem, uintptr(hBitmap))
// 	if err != nil {
// 		return fmt.Errorf("failed to select bitmap into DC: %v", err)
// 	}
// 	defer SelectObject(hdcMem, oldBitmap)

// 	// Prepare BITMAPINFO
// 	bi := BITMAPINFO{
// 		BmiHeader: BITMAPINFOHEADER{
// 			BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})), // Size of this structure
// 			BiWidth:       int32(width),                              // Width of bitmap
// 			BiHeight:      -int32(height),                            // Top-down DIB
// 			BiPlanes:      1,                                         // 1 plane
// 			BiBitCount:    32,                                        // 32 bits per pixel (RGBA)
// 			BiCompression: BI_RGB,                                    // RGB encoding
// 		},
// 	}

// 	// Calculate the size of the bitmap data
// 	dataSize := uint32(width * height * 4) // 4 bytes per pixel (RGBA)

// 	// Allocate memory for bitmap data
// 	rawMem, err := GlobalAlloc(GMEM_FIXED|GMEM_ZEROINIT, dataSize)
// 	if err != nil {
// 		return fmt.Errorf("failed to allocate memory: %v", err)
// 	}
// 	defer GlobalFree(rawMem)

// 	bmpSlice, err := GlobalLock(rawMem)
// 	if err != nil {
// 		return fmt.Errorf("failed to lock memory: %v", err)
// 	}
// 	defer GlobalUnlock(rawMem)

// 	// Copy image data to memory block
// 	copy((*[1 << 30]byte)(bmpSlice)[:dataSize], rgba.Pix)

// 	// Draw the bitmap to the device context
// 	if err := SetDIBitsToDevice(
// 		hdc,
// 		uint32(x), uint32(y), // Destination position
// 		uint32(width), uint32(height), // Width and height of destination area
// 		0, 0, // Source position
// 		0, uint32(height), // Scanlines to copy
// 		bmpSlice,
// 		&bi,
// 		DIB_RGB_COLORS,
// 	); err != nil {
// 		return fmt.Errorf("failed to draw bitmap: %v", err)
// 	}
// 	return nil
// }

func checkDCSettings(hdc HDC) {
	// 获取颜色深度
	bitsPerPixel := GetDeviceCaps(hdc, BITSPIXEL)
	log.Debugf("Device context color depth: %d bits per pixel", bitsPerPixel)

	// 获取水平和垂直分辨率
	horizontalRes := GetDeviceCaps(hdc, HORZRES)
	verticalRes := GetDeviceCaps(hdc, VERTRES)
	log.Debugf("Device context resolution: %dx%d", horizontalRes, verticalRes)

	// 获取调色板大小
	numColors := GetDeviceCaps(hdc, NUMCOLORS)
	log.Debugf("Device context number of colors: %d", numColors)

	// 检查 BITMAPINFOHEADER.BiBitCount 是否与 bitsPerPixel 一致
	if bitsPerPixel != 32 {
		log.Fatalf("Mismatch in color depth: Device context is %d bits per pixel, but BITMAPINFOHEADER is set to 32 bits per pixel", bitsPerPixel)
	}
}
func decodeRLE(r io.Reader, width, height int) (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		x := 0
		for x < width {
			var count byte
			if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
				return nil, err
			}
			var color color.RGBA
			if err := binary.Read(r, binary.LittleEndian, &color); err != nil {
				return nil, err
			}
			for i := 0; i < int(count) && x < width; i++ {
				img.Set(x, y, color)
				x++
			}
		}
	}
	return img, nil
}
