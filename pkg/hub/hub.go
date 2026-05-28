package hub

import (
	"go.uber.org/zap"

	"dis-alg/pkg/core"
)

type BroadcastMessage struct {
	Sender *Client
	Packet *core.Packet
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan BroadcastMessage
	register   chan *Client
	unregister chan *Client
	logger     *zap.Logger
}

// NewHub creates a new Hub instance.
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan BroadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the hub event loop.
func (h *Hub) Run() {
	h.logger.Info("Hub event loop started")
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.Info("Client registered", zap.String("addr", client.conn.RemoteAddr()), zap.Int("total_clients", len(h.clients)))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.logger.Info("Client unregistered", zap.String("addr", client.conn.RemoteAddr()), zap.Int("total_clients", len(h.clients)))
			}

		case msg := <-h.broadcast:
			for client := range h.clients {
				// Requirement 1: Do not send the packet back to the originator
				if client == msg.Sender {
					continue
				}
				select {
				case client.send <- msg.Packet:
				default:
					h.logger.Warn("Slow consumer detected, dropping packet", zap.String("addr", client.conn.RemoteAddr()), zap.Uint64("packet_num", msg.Packet.PacketNumber))
				}
			}
		}
	}
}
