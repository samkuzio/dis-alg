package terminal

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"dis-alg/pkg/logger"
	"dis-alg/pkg/transport/tcp"
)

func Run() {
	fs := flag.NewFlagSet("terminal", flag.ExitOnError)

	var simAddr string
	var hubAddr string
	var transportType string
	var verbose bool

	// Support standard flags
	fs.StringVar(&simAddr, "simulation", "", "The IP and Port on which the terminal node will listen to simulation traffic (e.g. 192.168.1.255:3000)")
	fs.StringVar(&simAddr, "s", "", "Alias for --simulation")

	fs.StringVar(&hubAddr, "hub", "", "The IP and port on which the terminal node will connect to the hub (e.g. 10.0.0.1:8080)")
	fs.StringVar(&hubAddr, "h", "", "Alias for --hub")

	fs.StringVar(&transportType, "transport", "tcp", "The transport protocol (default tcp)")
	fs.StringVar(&transportType, "t", "tcp", "Alias for --transport")

	fs.BoolVar(&verbose, "verbose", false, "Enable verbose debug logging")
	fs.BoolVar(&verbose, "v", false, "Alias for --verbose")

	// Parse arguments (skipping 'dis-alg terminal')
	if len(os.Args) < 2 {
		fs.Usage()
		os.Exit(1)
	}

	fs.Parse(os.Args[2:])

	// Basic validation
	if simAddr == "" || hubAddr == "" {
		fmt.Println("Error: --simulation and --hub flags are required.")
		fs.Usage()
		os.Exit(1)
	}

	if transportType != "tcp" {
		fmt.Printf("Error: unsupported transport '%s'. Only 'tcp' is supported.\n", transportType)
		os.Exit(1)
	}

	log, err := logger.New(verbose)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Generate a random source ID for this terminal node
	rand.Seed(time.Now().UnixNano())
	sourceID := rand.Uint32()

	// Initialize TCP dialer
	dialer := tcp.NewDialer()

	node, err := NewTerminalNode(sourceID, simAddr, hubAddr, dialer, log)
	if err != nil {
		log.Error("Failed to initialize terminal node", zap.Error(err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down terminal node...")
		cancel()
	}()

	if err := node.Run(ctx); err != nil {
		log.Error("Terminal node failed", zap.Error(err))
		os.Exit(1)
	}
}
