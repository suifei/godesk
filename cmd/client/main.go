package main

import (
	"flag"
	"log"

	"github.com/faiface/pixel/pixelgl"
	"github.com/suifei/godesk/internal/client"
)

func run() {
	serverIP := flag.String("server", "localhost", "Server IP address")
	serverPort := flag.String("port", "3388", "Server port")
	flag.Parse()

	serverAddr := *serverIP + ":" + *serverPort
	log.Printf("Connecting to server at %s", serverAddr)

	clientHandler, err := client.NewClientHandler(serverAddr)
	if err != nil {
		log.Fatalf("Failed to create client handler: %v", err)
	}

	log.Printf("Connected to server at %s", serverAddr)

	clientHandler.Handle()
}

func main() {
	pixelgl.Run(run)
}
