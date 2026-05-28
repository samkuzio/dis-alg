package terminal

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"

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
	logger    *zap.Logger
}

// NewTerminalNode creates a new TerminalNode.
func NewTerminalNode(sourceID uint32, simAddrStr, hubAddr string, dialer core.Dialer, logger *zap.Logger) (*TerminalNode, error) {
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
		logger:     logger,
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
	t.logger.Info("Terminal listening for local UDP", zap.String("addr", t.simAddrStr))

	// 2. Connect to the remote Hub
	hubConn, err := t.dialer.Dial(t.hubAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to Hub: %w", err)
	}
	defer hubConn.Close()
	t.logger.Info("Terminal connected to Hub", zap.String("hub_addr", t.hubAddr))

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 2.5 Force UDP and Hub close when context is cancelled to unblock ReadFrom
	go func() {
		<-ctx.Done()
		t.udp.Close()
		hubConn.Close()
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
			t.logger.Error("Failed to read from Hub", zap.Error(err))
			return
		}

		t.logger.Debug("Received packet from Hub", zap.Uint32("sourceID", packet.SourceID), zap.Uint64("packetNum", packet.PacketNumber), zap.Int("len", len(packet.Payload)))

		// Add to echo cache before broadcasting so our own listener drops it
		t.echoCache.Add(packet.Payload)

		// Broadcast to local network
		_, err = t.udp.WriteTo(packet.Payload, t.simAddr)
		if err != nil {
			t.logger.Error("Failed to write to local UDP", zap.Error(err))
			// Continue on UDP write errors (might be transient)
		}
	}
}

// runEgress reads packets from the local simulation network and forwards them to the Hub.
func (t *TerminalNode) runEgress(ctx context.Context, hubConn core.Connection) {
	buf := make([]byte, 65535) // Max UDP size

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
				t.logger.Error("Failed to read from local UDP", zap.Error(err))
				return
			}
		}

		payload := make([]byte, n)
		copy(payload, buf[:n])

		t.logger.Debug("Received packet from simulation socket", zap.Int("len", n))

		// Echo cancellation check
		if t.echoCache.Contains(payload) {
			t.logger.Debug("Dropped echoed packet", zap.Int("len", n))
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
			t.logger.Error("Failed to write to Hub", zap.Error(err))
			return
		}

		t.logger.Debug("Sent packet to Hub", zap.Uint64("packetNum", packet.PacketNumber), zap.Int("len", n))
	}
}
