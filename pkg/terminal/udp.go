package terminal

import (
	"fmt"
	"net"
)

// UDPTransceiver defines the interface for interacting with the local simulation network.
// This allows mocking for unit tests.
type UDPTransceiver interface {
	ReadFrom(b []byte) (int, net.Addr, error)
	WriteTo(b []byte, addr net.Addr) (int, error)
	Close() error
}

// standardUDPConn implements UDPTransceiver using a standard *net.UDPConn.
type standardUDPConn struct {
	conn *net.UDPConn
}

func (c *standardUDPConn) ReadFrom(b []byte) (int, net.Addr, error) {
	return c.conn.ReadFrom(b)
}

func (c *standardUDPConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	return c.conn.WriteTo(b, addr)
}

func (c *standardUDPConn) Close() error {
	return c.conn.Close()
}

// NewUDPTransceiver binds a UDP socket to listen for incoming DIS packets.
func NewUDPTransceiver(bindAddr string) (UDPTransceiver, error) {
	addr, err := net.ResolveUDPAddr("udp", bindAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to bind UDP socket: %w", err)
	}

	return &standardUDPConn{conn: conn}, nil
}
