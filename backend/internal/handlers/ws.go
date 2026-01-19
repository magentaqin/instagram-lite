package handlers

import (
	"net/http"

	"instagram-lite-backend/internal/realtime"

	"github.com/gorilla/websocket"
)

type WSHandler struct {
	hub *realtime.Hub
}

func NewWSHandler(hub *realtime.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// Upgrade Http to Websocket
var upgrader = websocket.Upgrader{
	// In this homework, simply allow all origins.
	// In production, we should validate Origin properly.
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "websocket upgrade failed", http.StatusBadRequest)
		return
	}

	c := realtime.NewClient(conn)

	// Register client with hub.
	h.hub.Register(c)

	// Start pumps.
	go c.WritePump()
	go c.ReadPump(h.hub)
}