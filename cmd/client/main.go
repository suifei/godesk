package main

import (
    "flag"
    "log"
    "net"

    "github.com/suifei/godesk/internal/client"
    "github.com/suifei/godesk/pkg/network"
)

func main() {
    serverAddr := flag.String("server", "localhost:8000", "Server address")
    flag.Parse()

    conn, err := net.Dial("tcp", *serverAddr)
    if err != nil {
        log.Fatalf("Failed to connect to server: %v", err)
    }
    defer conn.Close()

    log.Printf("Connected to server at %s", *serverAddr)

    tcpConn := network.NewTCPConnection(conn)
    clientHandler := client.NewClientHandler(tcpConn)
    clientHandler.Handle()
}