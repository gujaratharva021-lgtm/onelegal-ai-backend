// Package ws is the WebRTC signaling relay: it exchanges offers, answers,
// ICE candidates, and call-lifecycle messages between two authenticated
// users' WebSocket connections. It never touches audio/video — media flows
// directly device-to-device via WebRTC (P2P) once signaling completes.
package ws

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{clients: make(map[uuid.UUID]*websocket.Conn)}
}

func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Replace any stale connection for this user (e.g. app reconnect).
	if existing, ok := h.clients[userID]; ok {
		existing.Close()
	}
	h.clients[userID] = conn
}

func (h *Hub) Unregister(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if current, ok := h.clients[userID]; ok && current == conn {
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
	conn, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	data, err := json.Marshal(v)
	if err != nil {
		return false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return false
	}
	return true
}
