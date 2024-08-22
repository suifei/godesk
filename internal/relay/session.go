
package relay

import (
    "net"
)

type Session struct {
    ID   string
    Conn net.Conn
}

func NewSession(id string, conn net.Conn) *Session {
    return &Session{
        ID:   id,
        Conn: conn,
    }
}