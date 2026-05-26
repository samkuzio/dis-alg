package terminal

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"dis-alg/pkg/core"
)

// MockConnection for testing Hub <-> Spoke
type MockConnection struct {
	readChan  chan *core.Packet
	writeChan chan *core.Packet
	closed    bool
	mu        sync.Mutex
}

func (m *MockConnection) ReadPacket() (*core.Packet, error) {
	p, ok := <-m.readChan
	if !ok {
		return nil, net.ErrClosed
	}
	return p, nil
}

func (m *MockConnection) WritePacket(p *core.Packet) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return net.ErrClosed
	}
	m.mu.Unlock()
	m.writeChan <- p
	return nil
}

func (m *MockConnection) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		m.closed = true
		close(m.readChan)
	}
	return nil
}

func (m *MockConnection) RemoteAddr() string {
	return "mock-remote"
}

// MockDialer
type MockDialer struct {
	conn *MockConnection
	err  error
}

func (m *MockDialer) Dial(address string) (core.Connection, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.conn, nil
}

// MockUDPTransceiver
type MockUDPTransceiver struct {
	readChan  chan []byte
	writeChan chan []byte
	closed    bool
	mu        sync.Mutex
}

func (m *MockUDPTransceiver) ReadFrom(b []byte) (int, net.Addr, error) {
	data, ok := <-m.readChan
	if !ok {
		return 0, nil, net.ErrClosed
	}
	n := copy(b, data)
	return n, &net.UDPAddr{}, nil
}

func (m *MockUDPTransceiver) WriteTo(b []byte, addr net.Addr) (int, error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return 0, net.ErrClosed
	}
	m.mu.Unlock()

	// Make a copy so caller can reuse buffer
	cp := make([]byte, len(b))
	copy(cp, b)

	select {
	case m.writeChan <- cp:
	default: // non-blocking for test
	}
	return len(b), nil
}

func (m *MockUDPTransceiver) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		m.closed = true
		close(m.readChan)
	}
	return nil
}

func TestTerminalNode_Ingress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockConn := &MockConnection{
		readChan:  make(chan *core.Packet, 10),
		writeChan: make(chan *core.Packet, 10),
	}
	mockDialer := &MockDialer{conn: mockConn}

	node, err := NewTerminalNode(123, "127.0.0.1:0", "hub:8080", mockDialer)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	mockUDP := &MockUDPTransceiver{
		readChan:  make(chan []byte, 10),
		writeChan: make(chan []byte, 10),
	}

	// Override the UDP transceiver with our mock
	// (We need to start runIngress directly since Run() initializes its own UDP)
	node.udp = mockUDP

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		node.runIngress(ctx, mockConn)
	}()

	// Send a packet from the Hub
	payload := []byte("test-payload")
	mockConn.readChan <- &core.Packet{
		SourceID:     456,
		PacketNumber: 1,
		Payload:      payload,
	}

	// Verify it was written to UDP
	select {
	case out := <-mockUDP.writeChan:
		if string(out) != string(payload) {
			t.Errorf("Expected payload %q, got %q", payload, out)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for UDP write")
	}

	// Verify it was added to the echo cache
	if !node.echoCache.Contains(payload) {
		t.Error("Payload was not added to echo cache")
	}

	cancel()
	mockConn.Close()
	wg.Wait()
}

func TestTerminalNode_Egress_And_Echo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockConn := &MockConnection{
		readChan:  make(chan *core.Packet, 10),
		writeChan: make(chan *core.Packet, 10),
	}
	mockDialer := &MockDialer{conn: mockConn}

	node, err := NewTerminalNode(123, "127.0.0.1:0", "hub:8080", mockDialer)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	mockUDP := &MockUDPTransceiver{
		readChan:  make(chan []byte, 10),
		writeChan: make(chan []byte, 10),
	}
	node.udp = mockUDP

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		node.runEgress(ctx, mockConn)
	}()

	// 1. Send a local UDP packet
	payload := []byte("local-sim-packet")
	mockUDP.readChan <- payload

	// Verify it is forwarded to the Hub
	select {
	case p := <-mockConn.writeChan:
		if p.SourceID != 123 {
			t.Errorf("Expected SourceID 123, got %d", p.SourceID)
		}
		if p.PacketNumber != 1 {
			t.Errorf("Expected PacketNumber 1, got %d", p.PacketNumber)
		}
		if string(p.Payload) != string(payload) {
			t.Errorf("Expected payload %q, got %q", payload, p.Payload)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for Hub write")
	}

	// 2. Simulate an ECHO (packet that was already in the echo cache)
	echoPayload := []byte("echo-packet")
	node.echoCache.Add(echoPayload)

	// Send the echo packet via UDP
	mockUDP.readChan <- echoPayload

	// Verify it is DROPPED and not forwarded to the Hub
	select {
	case <-mockConn.writeChan:
		t.Fatal("Echo packet was forwarded to Hub, expected it to be dropped")
	case <-time.After(500 * time.Millisecond):
		// Expected timeout (packet dropped)
	}

	cancel()
	mockUDP.Close()
	wg.Wait()
}
