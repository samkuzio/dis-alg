package hub

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"dis-alg/pkg/core"
)

type failingMockConnection struct {
	*mockConnection
}

func (m *failingMockConnection) WritePacket(p *core.Packet) error {
	return errors.New("simulated write error")
}

func TestClient_WritePumpError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	hub := NewHub(logger)
	go hub.Run()

	mockConn := newMockConnection("error-client", 10)
	// Create client with the failing mock connection wrapper
	client := &Client{
		hub:  hub,
		conn: &failingMockConnection{mockConnection: mockConn},
		send: make(chan *core.Packet, 1),
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Send a packet directly to send chan
	client.send <- &core.Packet{}

	done := make(chan struct{})
	go func() {
		client.writePump() // Should exit due to simulated error
		close(done)
	}()

	select {
	case <-done:
		// Success! it exited gracefully
	case <-time.After(1 * time.Second):
		t.Fatal("writePump did not exit on write error")
	}
}

func TestRunServer_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)
	go func() {
		errChan <- RunServer(ctx, "tcp", "127.0.0.1:0")
	}()

	time.Sleep(50 * time.Millisecond) // Let it start listening
	cancel()                          // Trigger shutdown

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("RunServer returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("RunServer did not shut down when context was canceled")
	}
}
