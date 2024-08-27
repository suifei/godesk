package client

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"runtime"
	"sync"
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

var (
	gdiplus                       = windows.NewLazySystemDLL("gdiplus.dll")
	procGdiplusStartup            = gdiplus.NewProc("GdiplusStartup")
	procGdiplusShutdown           = gdiplus.NewProc("GdiplusShutdown")
	procGdipCreateBitmapFromScan0 = gdiplus.NewProc("GdipCreateBitmapFromScan0")
	procGdipCreateFromHDC         = gdiplus.NewProc("GdipCreateFromHDC")
	procGdipDrawImageRectRect     = gdiplus.NewProc("GdipDrawImageRectRect")
	procGdipDeleteGraphics        = gdiplus.NewProc("GdipDeleteGraphics")
	procGdipDisposeImage          = gdiplus.NewProc("GdipDisposeImage")
	procGdipGraphicsClear         = gdiplus.NewProc("GdipGraphicsClear")
	procGdipCreatePen1            = gdiplus.NewProc("GdipCreatePen1")
	procGdipDeletePen             = gdiplus.NewProc("GdipDeletePen")
	procGdipDrawRectangle         = gdiplus.NewProc("GdipDrawRectangle")
)

type GdiplusStartupInput struct {
	GdiplusVersion           uint32
	DebugEventCallback       uintptr
	SuppressBackgroundThread int32
	SuppressExternalCodecs   int32
}

// GDI+ types
type GpImage uintptr
type GpGraphics uintptr
type GpPen uintptr
type GpBitmap uintptr

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

type WMDrawInfo struct {
	rgba         *image.RGBA
	x, y         int
	screenWidth  int
	screenHeight int
}

type Display struct {
	window       ui.WindowMain
	width        int
	height       int
	imageData    *image.RGBA
	inputChan    chan InputEvent
	gdiplusToken uintptr

	lastDrawInfo WMDrawInfo
	drawMutex    sync.Mutex
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
			ClientArea(win.SIZE{Cx: int32(width), Cy: int32(height)}).
			WndStyles(co.WS_CAPTION | co.WS_SYSMENU | co.WS_OVERLAPPEDWINDOW |
				co.WS_BORDER | co.WS_VISIBLE | co.WS_MINIMIZEBOX | co.WS_CLIPCHILDREN |
				co.WS_MAXIMIZEBOX | co.WS_SIZEBOX),
	)
	// 初始化 GDI+
	var token uintptr
	startup := GdiplusStartupInput{GdiplusVersion: 1}
	ret, _, _ := procGdiplusStartup.Call(
		uintptr(unsafe.Pointer(&token)),
		uintptr(unsafe.Pointer(&startup)),
		0,
	)
	if ret != 0 {
		return nil, fmt.Errorf("GdiplusStartup failed with status %d", ret)
	}
	d.gdiplusToken = token
	log.Debug("GDI+ initialized successfully")
	d.setupEventHandlers()

	return d, nil
}

func (d *Display) GetClientSize() (int, int) {

	// 获取窗口的客户区大小
	var rect RECT
	GetClientRect(uintptr(d.window.Hwnd()), &rect)
	clientWidth := int(rect.Right - rect.Left)
	clientHeight := int(rect.Bottom - rect.Top)
	return clientWidth, clientHeight

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
		log.Debug("WM_PAINT received")
		var drawInfo WMDrawInfo
		d.drawMutex.Lock()
		drawInfo = d.lastDrawInfo
		d.drawMutex.Unlock()

		if drawInfo.rgba != nil {
			var ps win.PAINTSTRUCT
			hdc := d.window.Hwnd().BeginPaint(&ps)
			// hdc := d.window.Hwnd().GetDC()
			err := d.drawImage(hdc, drawInfo.rgba, drawInfo.x, drawInfo.y, drawInfo.screenWidth, drawInfo.screenHeight)
			d.window.Hwnd().EndPaint(&ps)
			if err != nil {
				log.Errorf("Failed to draw image: %v", err)
			} else {
				log.Debug("Image drawn successfully")
			}
		} else {
			log.Debug("No image data to draw")
		}

	})
	d.window.On().WmSize(func(p wm.Size) {
		d.window.Hwnd().InvalidateRect(nil, false)
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
		// log.Infof("Mouse move: %d, %d", p.Pos().X, p.Pos().Y)
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
	procGdiplusShutdown.Call(d.gdiplusToken)
}

func (d *Display) InputEvents() <-chan InputEvent {
	return d.inputChan
}

func (d *Display) UpdateScreen(update *protocol.ScreenUpdate, rgba *image.RGBA, x, y, screenWidth, screenHeight int) {
	if rgba == nil || rgba.Bounds().Empty() {
		log.Error("UpdateScreen called with invalid rgba")
		return
	}
	log.Debugf("UpdateScreen called: x=%d, y=%d, width=%d, height=%d, screenWidth=%d, screenHeight=%d",
		x, y, rgba.Bounds().Dx(), rgba.Bounds().Dy(), screenWidth, screenHeight)

	d.drawMutex.Lock()
	d.lastDrawInfo =  WMDrawInfo{
		rgba:         rgba,
		x:            x,
		y:            y,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}
	d.drawMutex.Unlock()

	// 使用 PostMessage 而不是直接调用 InvalidateRect
	d.window.Hwnd().PostMessage(co.WM_PAINT, 0, 0)

	d.window.Hwnd().UpdateWindow()
	log.Debug("Posted WM_PAINT message")
}

func (d *Display) drawImage(hdc win.HDC, rgba *image.RGBA, x, y, screenWidth, screenHeight int) error {
	if hdc == 0 {
		return fmt.Errorf("invalid HDC")
	}

	if rgba == nil || len(rgba.Pix) == 0 {
		return fmt.Errorf("invalid image data")
	}
	log.Debugf("Image data: %dx%d, Stride: %d, len(Pix): %d", rgba.Bounds().Dx(), rgba.Bounds().Dy(), rgba.Stride, len(rgba.Pix))

	// 获取窗口的客户区大小
	var rect RECT
	GetClientRect(uintptr(d.window.Hwnd()), &rect)
	clientWidth := int(rect.Right - rect.Left)
	clientHeight := int(rect.Bottom - rect.Top)

	// 计算缩放比例
	scaleX := float64(clientWidth) / float64(screenWidth)
	scaleY := float64(clientHeight) / float64(screenHeight)
	scale := math.Min(scaleX, scaleY)

	log.Debugf("Server dimensions: %dx%d, client dimensions: %dx%d", screenWidth, screenHeight, clientWidth, clientHeight)
	log.Debugf("Client dimensions: %dx%d, scale: %f", clientWidth, clientHeight, scale)

	// 计算缩放后的位置和尺寸
	scaledX := int(float64(x) * scale)
	scaledY := int(float64(y) * scale)
	scaledWidth := int(float64(rgba.Bounds().Dx()) * scale)
	scaledHeight := int(float64(rgba.Bounds().Dy()) * scale)

	// 确保缩放后的尺寸至少为1像素
	if scaledWidth < 1 {
		scaledWidth = 1
	}
	if scaledHeight < 1 {
		scaledHeight = 1
	}

	log.Debugf("Scaled dimensions: x=%d, y=%d, width=%d, height=%d", scaledX, scaledY, scaledWidth, scaledHeight)

	var bitmap GpBitmap
	ret, _, err := procGdipCreateBitmapFromScan0.Call(
		uintptr(rgba.Bounds().Dx()),
		uintptr(rgba.Bounds().Dy()),
		uintptr(rgba.Stride),
		uintptr(0x26200A), // PixelFormat32bppARGB
		uintptr(unsafe.Pointer(&rgba.Pix[0])),
		uintptr(unsafe.Pointer(&bitmap)),
	)
	if ret != 0 {
		return fmt.Errorf("GdipCreateBitmapFromScan0 failed with status %d: %v", ret, err)
	}
	defer procGdipDisposeImage.Call(uintptr(bitmap))
	log.Debug("GDI+ bitmap created successfully")

	var graphics GpGraphics
	ret, _, err = procGdipCreateFromHDC.Call(
		uintptr(hdc),
		uintptr(unsafe.Pointer(&graphics)),
	)
	if ret != 0 {
		return fmt.Errorf("GdipCreateFromHDC failed with status %d: %v", ret, err)
	}
	defer procGdipDeleteGraphics.Call(uintptr(graphics))
	log.Debug("GDI+ graphics created successfully")

	// 尝试清除背景
	// ret, _, err = procGdipGraphicsClear.Call(
	// 	uintptr(graphics),
	// 	uintptr(0xFFFFFFFF), // White color
	// )
	// if ret != 0 {
	// 	return fmt.Errorf("GdipGraphicsClear failed with status %d: %v", ret, err)
	// }
	// log.Debug("GDI+ background cleared successfully")

	ret, _, err = procGdipDrawImageRectRect.Call(
		uintptr(graphics),
		uintptr(bitmap),
		uintptr(float32(scaledX)),
		uintptr(float32(scaledY)),
		uintptr(float32(scaledWidth)),
		uintptr(float32(scaledHeight)),
		0,
		0,
		uintptr(float32(rgba.Bounds().Dx())),
		uintptr(float32(rgba.Bounds().Dy())),
		3, // Unit_Pixel
		0,
		0,
		0,
	)
	if ret != 0 {
		return fmt.Errorf("GdipDrawImageRectRect failed with status %d: %v", ret, err)
	}

	// 创建一个红色画笔
	var pen GpPen
	ret, _, _ = procGdipCreatePen1.Call(
		uintptr(0xFFFF0000), // 红色
		uintptr(float32(5)), // 线宽
		3,                   // 单位：像素
		uintptr(unsafe.Pointer(&pen)),
	)
	if ret != 0 {
		return fmt.Errorf("GdipCreatePen1 failed with status %d", ret)
	}
	defer procGdipDeletePen.Call(uintptr(pen))

	// 绘制一个矩形
	ret, _, _ = procGdipDrawRectangle.Call(
		uintptr(graphics),
		uintptr(pen),
		uintptr(float32(10)),
		uintptr(float32(10)),
		uintptr(float32(clientWidth-10)),
		uintptr(float32(clientHeight-10)),
	)
	if ret != 0 {
		return fmt.Errorf("GdipDrawRectangle failed with status %d", ret)
	}

	log.Debug("GDI+ drawing completed successfully")

	return nil
}

func (d *Display) drawImage3(hdc HDC, rgba *image.RGBA, width, height, x, y, screenWidth, screenHeight int) error {
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
