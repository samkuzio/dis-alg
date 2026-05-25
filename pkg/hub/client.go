package hub

import (
	"dis-alg/pkg/core"
)

// Client represents an active Spoke connection.
type Client struct {
	hub  *Hub
	conn core.Connection

	// Buffered channel of outbound packets.
	send chan *core.Packet
}

// readPump pumps messages from the core.Connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		packet, err := c.conn.ReadPacket()
		if err != nil {
			c.hub.logger.Info("Client read error or disconnected", "addr", c.conn.RemoteAddr(), "error", err.Error())
			break
		}
		c.hub.broadcast <- packet
	}
}

// writePump pumps messages from the hub to the core.Connection.
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		packet, ok := <-c.send
		if !ok {
			// The hub closed the channel (e.g., client unregistered)
			return
		}
		if err := c.conn.WritePacket(packet); err != nil {
			c.hub.logger.Info("Client write error", "addr", c.conn.RemoteAddr(), "error", err.Error())
			return
		}
	}
}
