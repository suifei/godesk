package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/suifei/godesk/pkg/log"

	"bytes"
	"image/png"

	"github.com/suifei/godesk/internal/server"
	"github.com/suifei/godesk/pkg/network"
)

func main() {
	// 定义命令行参数
	port := flag.String("port", "3388", "Port to listen on")
	flag.Parse()

	// 测试屏幕捕获
	testScreenCapture()

	// 构建监听地址
	addr := fmt.Sprintf("0.0.0.0:%s", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Infof("Server listening on %s", addr)

	for {
		log.Infof("Waiting for new connection...")
		conn, err := listener.Accept()
		if err != nil {
			log.Infof("Failed to accept connection: %v", err)
			continue
		}

		log.Infof("New connection accepted from %s", conn.RemoteAddr())
		tcpConn := network.NewTCPConnection(conn)
		go handleClient(tcpConn)
	}
}

func handleClient(conn *network.TCPConnection) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	log.Infof("Handling client from %s", remoteAddr)

	clientHandler := server.NewClientHandler(conn)
	log.Infof("Starting to handle client from %s", remoteAddr)

	// 添加一个 done channel 来通知主 goroutine 何时完成
	done := make(chan struct{})
	go func() {
		clientHandler.Handle()
		close(done)
	}()

	// 等待处理完成或超时
	select {
	case <-done:
		log.Errorf("Finished handling client from %s", remoteAddr)
	case <-time.After(5 * time.Minute):
		log.Errorf("Client handling timed out for %s", remoteAddr)
	}
}
func testScreenCapture() {
	capturer := server.NewCapturer(100 * time.Millisecond)
	update, err := capturer.CaptureScreen()
	if err != nil {
		log.Fatalf("Failed to capture screen: %v", err)
	}

	log.Debugf("Test capture successful: %dx%d, %d bytes",
		update.Width, update.Height, len(update.ImageData))

	// 保存捕获的图像
	img, err := png.Decode(bytes.NewReader(update.ImageData))
	if err != nil {
		log.Fatalf("Failed to decode captured image: %v", err)
	}

	outputDir := "test_captures"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	outputPath := filepath.Join(outputDir, "test_capture.png")
	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Fatalf("Failed to save captured image: %v", err)
	}

	log.Debugf("Test capture saved to: %s", outputPath)
}
