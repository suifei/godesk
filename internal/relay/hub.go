// internal/relay/hub.go

package relay

import (
    "net"
    "sync"
)

type Hub struct {
    sessions map[string]*Session
    mu       sync.Mutex
}

func NewHub() *Hub {
    return &Hub{
        sessions: make(map[string]*Session),
    }
}

func (h *Hub) AddSession(id string, conn net.Conn) {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.sessions[id] = NewSession(id, conn)
}

func (h *Hub) RemoveSession(id string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    delete(h.sessions, id)
}

func (h *Hub) GetSession(id string) (*Session, bool) {
    h.mu.Lock()
    defer h.mu.Unlock()
    session, ok := h.sessions[id]
    return session, ok
}