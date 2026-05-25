package tcp

import (
	"reflect"
	"testing"
	"time"

	"dis-alg/pkg/core"
)

func TestTCPTransport_Loopback(t *testing.T) {
	// Start listener on a random available port
	listener, err := NewListener("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Extract the actual assigned address for the test
	tcpL := listener.(*tcpListener)
	addr := tcpL.listener.Addr().String()

	// Run Echo Server
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return // Server closed
		}
		defer conn.Close()

		for {
			p, err := conn.ReadPacket()
			if err != nil {
				return
			}
			_ = conn.WritePacket(p)
		}
	}()

	// Give the server a tiny moment to start listening
	time.Sleep(10 * time.Millisecond)

	// Dial the server
	dialer := NewDialer()
	clientConn, err := dialer.Dial(addr)
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer clientConn.Close()

	// Create and send a test packet
	original := &core.Packet{
		SourceID:     101,
		PacketNumber: 202,
		Payload:      []byte("tcp-transport-integration-test"),
	}

	if err := clientConn.WritePacket(original); err != nil {
		t.Fatalf("Failed to write packet: %v", err)
	}

	// Read the echo from the server
	echoed, err := clientConn.ReadPacket()
	if err != nil {
		t.Fatalf("Failed to read packet: %v", err)
	}

	// Assert the packet passed through the network and protocol stack perfectly
	if !reflect.DeepEqual(original, echoed) {
		t.Fatalf("Echoed packet does not match original.\nGot: %+v\nWant: %+v", echoed, original)
	}
}
