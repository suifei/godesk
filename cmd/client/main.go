package main

import (
	"flag"
	"runtime/debug"

	"github.com/suifei/godesk/internal/client"
	"github.com/suifei/godesk/pkg/log"
)

func run() {

	serverIP := flag.String("server", "localhost", "Server IP address")
	serverPort := flag.String("port", "3388", "Server port")
	flag.Parse()

	serverAddr := *serverIP + ":" + *serverPort
	log.Infof("Connecting to server at %s", serverAddr)

	clientHandler, err := client.NewClientHandler(serverAddr)
	if err != nil {
		log.Fatalf("Failed to create client handler: %v", err)
	}

	log.Infof("Connected to server at %s", serverAddr)

	clientHandler.Handle()
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			// Print call stack when the program panics
			log.Errorf("Panic: %v", r)
			debug.PrintStack()

			debug.FreeOSMemory()

		}
	}()
	run()
}
