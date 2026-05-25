package hub

import (
	"log/slog"
	"os"

	"dis-alg/pkg/core"
	"dis-alg/pkg/transport/tcp"
)

const clientSendBufferSize = 256

// RunServer is the main entry point to start the Hub logic.
// It initializes the listener, starts the Hub event loop, and accepts connections.
func RunServer(transport, address string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	
	coreListener, err := tcp.NewListener(address)
	if err != nil {
		logger.Error("Failed to start listener", "error", err.Error())
		return err
	}
	defer coreListener.Close()

	hub := NewHub(logger)
	go hub.Run()

	logger.Info("Listening for connections", "transport", transport, "address", address)

	for {
		conn, err := coreListener.Accept()
		if err != nil {
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
