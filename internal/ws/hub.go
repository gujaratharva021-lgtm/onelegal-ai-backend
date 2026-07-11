// Package ws is the WebRTC signaling relay: it exchanges offers, answers,
// ICE candidates, and call-lifecycle messages between two authenticated
// users' WebSocket connections. It never touches audio/video — media flows
// directly device-to-device via WebRTC (P2P) once signaling completes.
package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// client pairs a connection with its own write lock. gorilla/websocket
// forbids concurrent writes on a single *Conn, but that lock must be
// per-connection, not the hub-wide lock — otherwise one slow/stalled client
// blocks Register/Unregister/SendJSON for every other user while its write
// is in flight.
type client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]*client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[uuid.UUID]*client)}
}

func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Replace any stale connection for this user (e.g. app reconnect).
	if existing, ok := h.clients[userID]; ok {
		existing.conn.Close()
	}
	h.clients[userID] = &client{conn: conn}
}

func (h *Hub) Unregister(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if current, ok := h.clients[userID]; ok && current.conn == conn {
		delete(h.clients, userID)
	}
}

func (h *Hub) IsOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// SendJSON relays a message to a specific user if they're currently
// connected. Returns false if the user is offline (caller can surface a
// "user offline" signal back to the sender).
func (h *Hub) SendJSON(userID uuid.UUID, v interface{}) bool {
	h.mu.RLock()
	cl, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	data, err := json.Marshal(v)
	if err != nil {
		return false
	}
	cl.mu.Lock()
	defer cl.mu.Unlock()
	_ = cl.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := cl.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return false
	}
	return true
}
