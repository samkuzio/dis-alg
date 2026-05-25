package core

// Connection represents a bi-directional stream to a Spoke.
// The underlying implementation handles the byte-level framing.
type Connection interface {
	// ReadPacket blocks until a full framed packet is read.
	ReadPacket() (*Packet, error)
	
	// WritePacket serializes and sends a packet over the wire.
	WritePacket(p *Packet) error
	
	// Close terminates the connection.
	Close() error
	
	// RemoteAddr returns the endpoint address for logging/observability.
	RemoteAddr() string
}

// Listener represents a server listening for incoming Spoke connections.
type Listener interface {
	// Accept blocks and returns the next incoming Spoke connection.
	Accept() (Connection, error)
	
	// Close stops listening.
	Close() error
}

// Dialer represents a client mechanism for a Spoke to connect to the Hub.
type Dialer interface {
	// Dial connects to the target address and returns a Connection.
	Dial(address string) (Connection, error)
}
