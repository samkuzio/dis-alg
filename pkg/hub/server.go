package hub

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"dis-alg/pkg/core"
	"dis-alg/pkg/transport/tcp"
)

// clientSendBufferSize is the maximum number of packets that can be queued for a single Spoke.
const clientSendBufferSize = 256

// RunServer is the main entry point to start the Hub logic.
func RunServer(ctx context.Context, transport, address string, logger *zap.Logger) error {
	coreListener, err := tcp.NewListener(address)
	if err != nil {
		logger.Error("Failed to start listener", zap.Error(err))
		return err
	}

	// Graceful shutdown mechanism
	go func() {
		<-ctx.Done()
		coreListener.Close()
	}()

	hub := NewHub(logger)
	go hub.Run()

	logger.Info("Listening for connections", zap.String("transport", transport), zap.String("address", address))

	for {
		conn, err := coreListener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") || err.Error() == "EOF" {
				return nil
			}
			logger.Error("Failed to accept connection", zap.Error(err))
			continue
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan *core.Packet, clientSendBufferSize),
		}

		hub.register <- client

		go client.readPump()
		go client.writePump()
	}
}
