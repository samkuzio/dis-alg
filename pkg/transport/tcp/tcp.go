package tcp

import (
	"net"

	"dis-alg/pkg/core"
	"dis-alg/pkg/protocol"
)

type tcpConnection struct {
	conn net.Conn
}

func (c *tcpConnection) ReadPacket() (*core.Packet, error) {
	return protocol.ReadPacket(c.conn)
}

func (c *tcpConnection) WritePacket(p *core.Packet) error {
	return protocol.WritePacket(c.conn, p)
}

func (c *tcpConnection) Close() error {
	return c.conn.Close()
}

func (c *tcpConnection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

type tcpListener struct {
	listener net.Listener
}

// NewListener creates a new TCP listener on the specified address.
func NewListener(address string) (core.Listener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &tcpListener{listener: l}, nil
}

func (l *tcpListener) Accept() (core.Connection, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &tcpConnection{conn: conn}, nil
}

func (l *tcpListener) Close() error {
	return l.listener.Close()
}

type tcpDialer struct{}

// NewDialer creates a new TCP dialer for outbound connections.
func NewDialer() core.Dialer {
	return &tcpDialer{}
}

func (d *tcpDialer) Dial(address string) (core.Connection, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &tcpConnection{conn: conn}, nil
}
