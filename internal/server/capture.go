package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"time"

	"github.com/suifei/godesk/pkg/log"

	"github.com/kbinani/screenshot"
	"github.com/suifei/godesk/internal/protocol"
)

type Capturer struct {
	interval    time.Duration
	updates     chan *protocol.ScreenUpdate
	stop        chan struct{}
	lastCapture *image.RGBA
}

func NewCapturer(interval time.Duration) *Capturer {
	return &Capturer{
		interval: interval,
		updates:  make(chan *protocol.ScreenUpdate),
		stop:     make(chan struct{}),
	}
}

func (c *Capturer) Start() {
	log.Infoln("Screen capturer started")
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			update, err := c.captureScreen()
			if err != nil {
				log.Errorf("Error capturing screen: %v", err)
				continue
			}
			if update != nil {
				log.Debugf("Screen captured: %dx%d, %d bytes",
					update.Width, update.Height, len(update.ImageData))
				c.updates <- update
			}
		case <-c.stop:
			log.Infoln("Screen capturer stopped")
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
		log.Errorf("Error capturing screen: %v", err)
		return nil, err
	}

	if c.lastCapture == nil {
		c.lastCapture = img
		return c.encodeFullImage(img)
	}

	diffRect, hasDiff := findDifference(c.lastCapture, img)
	if !hasDiff {
		return nil, nil // No difference, no need to send update
	}

	c.lastCapture = img
	return c.encodePartialImage(img, diffRect)
}

func (c *Capturer) encodeFullImage(img *image.RGBA) (*protocol.ScreenUpdate, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode full image: %v", err)
	}
	// compressionType, encodedData, err := c.encodeImage(img, false, img.Bounds())
	// if err != nil {
	// 	return nil, err
	// }

	return &protocol.ScreenUpdate{
		Width:           int32(img.Bounds().Dx()),
		Height:          int32(img.Bounds().Dy()),
		ImageData:       buf.Bytes(), // encodedData,
		Timestamp:       time.Now().UnixNano(),
		IsPartial:       false,
		CompressionType: protocol.CompressionType_PNG, // compressionType,
	}, nil
}

func (c *Capturer) encodePartialImage(img *image.RGBA, rect image.Rectangle) (*protocol.ScreenUpdate, error) {
	subImg := img.SubImage(rect).(*image.RGBA)
	//compressionType, encodedData, err := c.encodeImage(subImg, true, rect)
	// if err != nil {
	// 	return nil, err
	// }
	var buf bytes.Buffer
	if err := png.Encode(&buf, subImg); err != nil {
		return nil, fmt.Errorf("failed to encode partial image: %v", err)
	}

	return &protocol.ScreenUpdate{
		Width:           int32(rect.Dx()),
		Height:          int32(rect.Dy()),
		ImageData:       buf.Bytes(), //encodedData,
		Timestamp:       time.Now().UnixNano(),
		IsPartial:       true,
		X:               int32(rect.Min.X),
		Y:               int32(rect.Min.Y),
		CompressionType: protocol.CompressionType_PNG, //compressionType,
	}, nil
}

func (c *Capturer) encodeImage(img *image.RGBA, isPartial bool, rect image.Rectangle) (protocol.CompressionType, []byte, error) {
	var buf bytes.Buffer
	var compressionType protocol.CompressionType

	log.Debugf("Encoding image: isPartial=%v, rect=%v", isPartial, rect)

	if img == nil {
		return protocol.CompressionType_PNG, nil, fmt.Errorf("input image is nil")
	}

	if img.Bounds().Empty() {
		return protocol.CompressionType_PNG, nil, fmt.Errorf("input image is empty")
	}

	// Determine the best compression method
	isPhoto, isScreen := analyzeImageContent(img)
	log.Debugf("Content analysis: isPhoto=%v, isScreen=%v", isPhoto, isScreen)

	var err error
	if isPhoto {
		compressionType = protocol.CompressionType_JPEG
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	} else if isScreen {
		compressionType = protocol.CompressionType_RLE
		err = encodeRLE(&buf, img)
	} else {
		compressionType = protocol.CompressionType_PNG
		err = png.Encode(&buf, img)
	}

	if err != nil {
		return protocol.CompressionType_PNG, nil, fmt.Errorf("failed to encode image: %v", err)
	}

	log.Debugf("Image encoded successfully: type=%v, size=%d bytes", compressionType, buf.Len())

	return compressionType, buf.Bytes(), nil
}

func analyzeImageContent(img *image.RGBA) (isPhoto bool, isScreen bool) {
	bounds := img.Bounds()
	var totalVariation float64
	sameColorCount := 0
	totalPixels := bounds.Dx() * bounds.Dy()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		lastColor := img.RGBAAt(bounds.Min.X, y)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			currentColor := img.RGBAAt(x, y)
			if currentColor == lastColor {
				sameColorCount++
			}
			if x < bounds.Max.X-1 && y < bounds.Max.Y-1 {
				rightColor := img.RGBAAt(x+1, y)
				bottomColor := img.RGBAAt(x, y+1)
				totalVariation += colorDifference(currentColor, rightColor) + colorDifference(currentColor, bottomColor)
			}
			lastColor = currentColor
		}
	}

	avgVariation := totalVariation / float64(totalPixels)
	sameColorRatio := float64(sameColorCount) / float64(totalPixels)

	isPhoto = avgVariation > 30     // This threshold may need adjustment
	isScreen = sameColorRatio > 0.5 // This threshold may need adjustment

	return isPhoto, isScreen
}

func colorDifference(c1, c2 color.RGBA) float64 {
	return math.Sqrt(float64(
		(int(c1.R)-int(c2.R))*(int(c1.R)-int(c2.R)) +
			(int(c1.G)-int(c2.G))*(int(c1.G)-int(c2.G)) +
			(int(c1.B)-int(c2.B))*(int(c1.B)-int(c2.B))))
}

func encodeRLE(w io.Writer, img *image.RGBA) error {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		var count byte = 1
		lastColor := img.RGBAAt(bounds.Min.X, y)
		for x := bounds.Min.X + 1; x < bounds.Max.X; x++ {
			if img.RGBAAt(x, y) == lastColor && count < 255 {
				count++
			} else {
				w.Write([]byte{count})
				binary.Write(w, binary.LittleEndian, lastColor)
				count = 1
				lastColor = img.RGBAAt(x, y)
			}
		}
		w.Write([]byte{count})
		binary.Write(w, binary.LittleEndian, lastColor)
	}
	return nil
}

func findDifference(img1, img2 *image.RGBA) (image.Rectangle, bool) {
	bounds := img1.Bounds()
	if bounds != img2.Bounds() {
		return bounds, true
	}

	var minX, minY, maxX, maxY int
	hasDiff := false

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img1.RGBAAt(x, y) != img2.RGBAAt(x, y) {
				if !hasDiff {
					minX, minY, maxX, maxY = x, y, x, y
					hasDiff = true
				} else {
					minX = min(minX, x)
					minY = min(minY, y)
					maxX = max(maxX, x)
					maxY = max(maxY, y)
				}
			}
		}
	}

	if !hasDiff {
		return image.Rectangle{}, false
	}

	return image.Rect(minX, minY, maxX+1, maxY+1), true
}
func (c *Capturer) CaptureScreen() (*protocol.ScreenUpdate, error) {
	return c.captureScreen()
}
