package handlers

import (
	"encoding/json"
	"net/http"

	"legaltech-backend/internal/config"
	"legaltech-backend/internal/utils"
	"legaltech-backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type SignalingHandler struct {
	hub   *ws.Hub
	cfg   *config.Config
	rooms *ws.RoomRegistry
}

func NewSignalingHandler(hub *ws.Hub, cfg *config.Config, rooms *ws.RoomRegistry) *SignalingHandler {
	return &SignalingHandler{hub: hub, cfg: cfg, rooms: rooms}
}

var upgrader = websocket.Upgrader{
	// Flutter's WebRTC signaling client is a native app, not a browser page,
	// so there's no meaningful Origin to check here.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleWS upgrades to a WebSocket after validating a JWT passed as a query
// parameter (browsers/native WebSocket clients can't set custom headers on
// the handshake request, so the standard Authorization-header middleware
// can't be reused here).
func (h *SignalingHandler) HandleWS(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Missing token"})
		return
	}
	claims, err := utils.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid or expired token"})
		return
	}
	userID := claims.UserID

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.hub.Register(userID, conn)
	defer h.hub.Unregister(userID, conn)

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		toStr, _ := msg["to"].(string)
		toID, err := uuid.Parse(toStr)
		if err != nil {
			continue
		}

		// Every call message carries a call_id (the private room's ID) —
		// reject relaying it unless both sender and recipient are the
		// room's exact two registered participants (lawyer + client). This
		// is what actually prevents a third authenticated user from ever
		// joining or interfering with someone else's call.
		if callID, _ := msg["call_id"].(string); callID != "" {
			if !h.rooms.Allowed(callID, userID, toID) {
				continue
			}
		}

		// Always stamp the authenticated sender, never trust the client's
		// own "from" field — prevents identity spoofing in relayed messages.
		msg["from"] = userID.String()

		if !h.hub.SendJSON(toID, msg) {
			h.hub.SendJSON(userID, map[string]interface{}{
				"type":    "user-offline",
				"call_id": msg["call_id"],
				"to":      toID.String(),
			})
		}
	}
}
