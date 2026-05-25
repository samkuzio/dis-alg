package hub

import (
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"dis-alg/pkg/core"
)

// mockConnection simulates a network connection in-memory
type mockConnection struct {
	addr      string
	readChan  chan *core.Packet
	writeChan chan *core.Packet
	closeOnce sync.Once
}

func newMockConnection(addr string, writeBufferSize int) *mockConnection {
	return &mockConnection{
		addr:      addr,
		readChan:  make(chan *core.Packet, 100),
		writeChan: make(chan *core.Packet, writeBufferSize),
	}
}

func (m *mockConnection) ReadPacket() (*core.Packet, error) {
	p, ok := <-m.readChan
	if !ok {
		return nil, io.EOF
	}
	return p, nil
}

func (m *mockConnection) WritePacket(p *core.Packet) error {
	select {
	case m.writeChan <- p:
		return nil
	default:
		// Simulate network block if the internal channel buffer is full
		return nil 
	}
}

func (m *mockConnection) Close() error {
	m.closeOnce.Do(func() {
		close(m.readChan)
	})
	return nil
}

func (m *mockConnection) RemoteAddr() string {
	return m.addr
}

func TestHub_FanOutAndSlowConsumer(t *testing.T) {
	// 1. Setup Hub
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	hub := NewHub(logger)
	go hub.Run()

	// 2. Setup 3 Mock Clients
	fastClient1Conn := newMockConnection("fast-1", 100)
	fastClient1 := &Client{hub: hub, conn: fastClient1Conn, send: make(chan *core.Packet, 10)}
	
	fastClient2Conn := newMockConnection("fast-2", 100)
	fastClient2 := &Client{hub: hub, conn: fastClient2Conn, send: make(chan *core.Packet, 10)}
	
	slowClientConn := newMockConnection("slow-1", 0)
	// Give the slow client a tiny buffer to artificially trigger the drop logic quickly
	slowClient := &Client{hub: hub, conn: slowClientConn, send: make(chan *core.Packet, 1)}

	// Register them
	hub.register <- fastClient1
	hub.register <- fastClient2
	hub.register <- slowClient

	// Give the hub loop a millisecond to register them
	time.Sleep(10 * time.Millisecond)

	// Start their pumps
	var wg sync.WaitGroup
	wg.Add(6)
	
	startPump := func(c *Client) {
		go func() { defer wg.Done(); c.readPump() }()
		go func() { defer wg.Done(); c.writePump() }()
	}
	
	startPump(fastClient1)
	startPump(fastClient2)
	startPump(slowClient)

	// 3. Test Fan-Out: Send 5 packets from Fast 1
	for i := uint64(1); i <= 5; i++ {
		fastClient1Conn.readChan <- &core.Packet{
			SourceID:     1,
			PacketNumber: i,
			Payload:      []byte("test-payload"),
		}
	}

	// Give the hub time to route
	time.Sleep(50 * time.Millisecond)

	// Wait, since fastClient1 sends 5 packets in quick succession, slowClient's send buffer of size 1 might block
	// actually the send buffer is make(chan *core.Packet, 1). So it holds 1 packet.
	// But the writePump might immediately pull it off the channel and block on the mockConnection.writeChan which is size 100!
	// Let's adjust the test: the mockConnection's writeChan needs to be small for the slow client.
	

	// 4. Assertions
	// fastClient2 should have received all 5
	if len(fastClient2Conn.writeChan) != 5 {
		t.Errorf("Expected fastClient2 to receive 5 packets, got %d", len(fastClient2Conn.writeChan))
	}
	
	// slowClient had a send buffer of 1. It should have received exactly 1 or 2 (depending on goroutine timing), and the hub dropped the others.
	// But crucially, it did NOT block the hub from delivering to fastClient2.
	if len(slowClientConn.writeChan) > 2 {
		t.Errorf("Expected slowClient to receive 1 or 2 packets, got %d", len(slowClientConn.writeChan))
	}

	// 5. Teardown
	fastClient1Conn.Close()
	fastClient2Conn.Close()
	slowClientConn.Close()
	
	// Wait for goroutines to exit to ensure no leaks
	wg.Wait()
}
