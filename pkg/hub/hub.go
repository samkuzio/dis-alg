package hub

import (
	"log/slog"

	"dis-alg/pkg/core"
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *core.Packet

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Logger for structured observability.
	logger *slog.Logger
}

// NewHub creates a new Hub instance.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *core.Packet),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the hub event loop. It blocks indefinitely.
func (h *Hub) Run() {
	h.logger.Info("Hub event loop started")
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.Info("Client registered", "addr", client.conn.RemoteAddr(), "total_clients", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.logger.Info("Client unregistered", "addr", client.conn.RemoteAddr(), "total_clients", len(h.clients))
			}

		case packet := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- packet:
					// Packet successfully buffered for the client's writePump
				default:
					// Slow consumer detected: buffer is full.
					// Architecture dictates "best-effort" delivery, so we drop it.
					h.logger.Warn("Slow consumer detected, dropping packet", "addr", client.conn.RemoteAddr(), "packet_num", packet.PacketNumber)
				}
			}
		}
	}
}
