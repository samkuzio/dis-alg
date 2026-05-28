package hub

import (
	"go.uber.org/zap"
	"dis-alg/pkg/core"
)

type Client struct {
	hub  *Hub
	conn core.Connection
	send chan *core.Packet
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		packet, err := c.conn.ReadPacket()
		if err != nil {
			c.hub.logger.Error("Client read error or unexpected disconnect", zap.String("addr", c.conn.RemoteAddr()), zap.Error(err))
			break
		}
		c.hub.logger.Debug("Packet received at hub", zap.String("origin", c.conn.RemoteAddr()), zap.Uint32("source_id", packet.SourceID), zap.Uint64("packet_num", packet.PacketNumber))
		
		c.hub.broadcast <- BroadcastMessage{
			Sender: c,
			Packet: packet,
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		packet, ok := <-c.send
		if !ok {
			return
		}
		if err := c.conn.WritePacket(packet); err != nil {
			c.hub.logger.Error("Client write error", zap.String("addr", c.conn.RemoteAddr()), zap.Error(err))
			return
		}
	}
}
