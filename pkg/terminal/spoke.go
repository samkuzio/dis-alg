package terminal

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"

	"dis-alg/pkg/core"
)

// TerminalNode represents a local spoke that bridges UDP to a remote Hub.
type TerminalNode struct {
	simAddrStr string
	simAddr    *net.UDPAddr
	udp        UDPTransceiver
	dialer     core.Dialer
	hubAddr    string
	sourceID   uint32
	echoCache  *EchoCache

	packetNum uint64
}

// NewTerminalNode creates a new TerminalNode.
func NewTerminalNode(sourceID uint32, simAddrStr, hubAddr string, dialer core.Dialer) (*TerminalNode, error) {
	// Parse the target UDP address to broadcast to
	simAddr, err := net.ResolveUDPAddr("udp", simAddrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid simulation address: %w", err)
	}

	return &TerminalNode{
		simAddrStr: simAddrStr,
		simAddr:    simAddr,
		dialer:     dialer,
		hubAddr:    hubAddr,
		sourceID:   sourceID,
		echoCache:  NewEchoCache(512), // Cache last 512 packets
	}, nil
}

// Run starts the terminal node logic blocking until context is cancelled.
func (t *TerminalNode) Run(ctx context.Context) error {
	// 1. Bind to local UDP network
	udpConn, err := NewUDPTransceiver(t.simAddrStr)
	if err != nil {
		return fmt.Errorf("failed to initialize UDP transceiver: %w", err)
	}
	t.udp = udpConn
	defer t.udp.Close()
	slog.Info("Terminal listening for local UDP", "addr", t.simAddrStr)

	// 2. Connect to the remote Hub
	hubConn, err := t.dialer.Dial(t.hubAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to Hub: %w", err)
	}
	defer hubConn.Close()
	slog.Info("Terminal connected to Hub", "hub_addr", t.hubAddr)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 2.5 Force UDP close when context is cancelled to unblock ReadFrom
	go func() {
		<-ctx.Done()
		t.udp.Close()
	}()

	// 3. Start Ingress goroutine (Hub -> Local UDP)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel() // if ingress fails, shut down the whole node
		t.runIngress(ctx, hubConn)
	}()

	// 4. Start Egress goroutine (Local UDP -> Hub)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel() // if egress fails, shut down the whole node
		t.runEgress(ctx, hubConn)
	}()

	// Wait for shutdown or error
	wg.Wait()
	return nil
}

// runIngress reads packets from the Hub and broadcasts them to the local simulation network.
func (t *TerminalNode) runIngress(ctx context.Context, hubConn core.Connection) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		packet, err := hubConn.ReadPacket()
		if err != nil {
			slog.Error("Failed to read from Hub", "error", err)
			return
		}

		// Add to echo cache before broadcasting so our own listener drops it
		t.echoCache.Add(packet.Payload)

		// Broadcast to local network
		_, err = t.udp.WriteTo(packet.Payload, t.simAddr)
		if err != nil {
			slog.Error("Failed to write to local UDP", "error", err)
			// Continue on UDP write errors (might be transient)
		} else {
			slog.Debug("Bridged packet to UDP", "sourceID", packet.SourceID, "packetNum", packet.PacketNumber, "len", len(packet.Payload))
		}
	}
}

// runEgress reads packets from the local simulation network and forwards them to the Hub.
func (t *TerminalNode) runEgress(ctx context.Context, hubConn core.Connection) {
	buf := make([]byte, 65535) // Max UDP size

	// We use a small timeout on UDP read if possible to check context cancellation,
	// but standard net.UDPConn doesn't easily support select {}.
	// For this PoC, we let it block. If the connection is closed during shutdown, ReadFrom will return an error.

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, _, err := t.udp.ReadFrom(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				return // Expected during shutdown
			default:
				slog.Error("Failed to read from local UDP", "error", err)
				return
			}
		}

		payload := make([]byte, n)
		copy(payload, buf[:n])

		// Echo cancellation check
		if t.echoCache.Contains(payload) {
			slog.Debug("Dropped echoed packet", "len", n)
			continue
		}

		// Prepare packet for Hub
		packet := &core.Packet{
			SourceID:     t.sourceID,
			PacketNumber: atomic.AddUint64(&t.packetNum, 1),
			Payload:      payload,
		}

		err = hubConn.WritePacket(packet)
		if err != nil {
			slog.Error("Failed to write to Hub", "error", err)
			return
		}

		slog.Debug("Bridged packet to Hub", "packetNum", packet.PacketNumber, "len", n)
	}
}
