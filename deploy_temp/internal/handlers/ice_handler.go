package handlers

import (
	"legaltech-backend/internal/config"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type ICEHandler struct {
	cfg *config.Config
}

func NewICEHandler(cfg *config.Config) *ICEHandler {
	return &ICEHandler{cfg: cfg}
}

// GetServers returns the ICE server list (STUN + optional TURN) for the
// Flutter client's RTCPeerConnection config. TURN is only included when
// configured via .env — STUN-only is the default.
func (h *ICEHandler) GetServers(c *gin.Context) {
	servers := []gin.H{
		{"urls": "stun:stun.l.google.com:19302"},
	}

	if h.cfg.TurnURL != "" {
		turn := gin.H{"urls": h.cfg.TurnURL}
		if h.cfg.TurnUsername != "" {
			turn["username"] = h.cfg.TurnUsername
		}
		if h.cfg.TurnCredential != "" {
			turn["credential"] = h.cfg.TurnCredential
		}
		servers = append(servers, turn)
	}

	response.Success(c, 200, "ice servers", gin.H{"ice_servers": servers})
}
