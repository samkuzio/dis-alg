package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"dis-alg/pkg/hub"
	"dis-alg/pkg/logger"
	"dis-alg/pkg/terminal"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) >= 2 {
		arg := os.Args[1]
		if arg == "-h" || arg == "--help" || arg == "help" {
			printHelp()
			os.Exit(0)
		}
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: dis-alg <hub|terminal> [args...]")
		fmt.Println("Run 'dis-alg --help' for more information.")
		os.Exit(1)
	}

	mode := os.Args[1]
	switch mode {
	case "hub":
		fs := flag.NewFlagSet("hub", flag.ExitOnError)
		var verbose bool
		fs.BoolVar(&verbose, "v", false, "Enable verbose debug logging")
		fs.BoolVar(&verbose, "verbose", false, "Enable verbose debug logging")
		fs.Parse(os.Args[2:])

		args := fs.Args()
		if len(args) < 2 {
			fmt.Println("Usage: dis-alg hub [-v] [transport] [bind-ip]:[port]")
			os.Exit(1)
		}
		transport := args[0]
		address := args[1]

		if transport != "tcp" {
			fmt.Printf("Error: unsupported transport '%s'. Only 'tcp' is supported.\n", transport)
			os.Exit(1)
		}

		log, err := logger.New(verbose)
		if err != nil {
			fmt.Printf("Failed to initialize logger: %v\n", err)
			os.Exit(1)
		}
		defer log.Sync()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Println("\nShutting down...")
			cancel()
		}()

		if err := hub.RunServer(ctx, transport, address, log); err != nil {
			log.Error("Hub server failed", zap.Error(err))
			os.Exit(1)
		}
	case "terminal":
		terminal.Run()
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		fmt.Println("Usage: dis-alg <hub|terminal> [args...]")
		os.Exit(1)
	}
}
