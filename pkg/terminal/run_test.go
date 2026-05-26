package terminal

import (
	"context"
	"testing"
	"time"

	"dis-alg/pkg/core"
)

// A listener to satisfy MockDialer for testing node.Run()
type MockListener struct {
	addr string
}

func TestTerminalNode_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mockConn := &MockConnection{
		readChan:  make(chan *core.Packet, 10),
		writeChan: make(chan *core.Packet, 10),
	}
	mockDialer := &MockDialer{conn: mockConn}

	// Use a random local port for UDP
	node, err := NewTerminalNode(123, "127.0.0.1:0", "hub:8080", mockDialer)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- node.Run(ctx)
	}()

	// Wait briefly for the node to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown the node
	cancel()
	mockConn.Close() // also close the mock connection so ingress exits

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected nil error on clean shutdown, got %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for node to shut down")
	}
}

func TestNewUDPTransceiver(t *testing.T) {
	udp, err := NewUDPTransceiver("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create UDP transceiver: %v", err)
	}
	defer udp.Close()

	// Just a quick check to make sure it implemented the interface
	if udp == nil {
		t.Fatal("Transceiver is nil")
	}
}
