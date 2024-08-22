package main

import (
	"flag"
	"log"
	"net"

	"github.com/suifei/godesk/internal/server"
	"github.com/suifei/godesk/pkg/network"
)

func main() {
	port := flag.String("port", "8000", "Port to listen on")
	flag.Parse()

	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Server listening on :%s", *port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		tcpConn := network.NewTCPConnection(conn)
		go handleClient(tcpConn)
	}
}

func handleClient(conn *network.TCPConnection) {
	defer conn.Close()

	clientHandler := server.NewClientHandler(conn)
	clientHandler.Handle()
}
