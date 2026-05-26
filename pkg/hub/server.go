package hub

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"dis-alg/pkg/core"
	"dis-alg/pkg/transport/tcp"
)

// clientSendBufferSize is the maximum number of packets that can be queued for a single Spoke.
const clientSendBufferSize = 256

// RunServer is the main entry point to start the Hub logic.
// It initializes the listener, starts the Hub event loop, and accepts connections.
// It will stop accepting connections and exit when the provided context is canceled.
func RunServer(ctx context.Context, transport, address string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	coreListener, err := tcp.NewListener(address)
	if err != nil {
		logger.Error("Failed to start listener", "error", err.Error())
		return err
	}

	// Graceful shutdown mechanism
	go func() {
		<-ctx.Done()
		coreListener.Close()
	}()

	hub := NewHub(logger)
	go hub.Run()

	logger.Info("Listening for connections", "transport", transport, "address", address)

	for {
		conn, err := coreListener.Accept()
		if err != nil {
			// Expected error if listener is closed from shutdown
			if strings.Contains(err.Error(), "use of closed network connection") || err.Error() == "EOF" {
				return nil
			}
			logger.Error("Failed to accept connection", "error", err.Error())
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
