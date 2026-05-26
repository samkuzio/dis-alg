package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"dis-alg/pkg/hub"
	"dis-alg/pkg/terminal"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dis-alg <hub|terminal> [args...]")
		os.Exit(1)
	}

	mode := os.Args[1]
	switch mode {
	case "hub":
		if len(os.Args) < 4 {
			fmt.Println("Usage: dis-alg hub [transport] [bind-ip]:[port]")
			os.Exit(1)
		}
		transport := os.Args[2]
		address := os.Args[3]

		if transport != "tcp" {
			fmt.Printf("Error: unsupported transport '%s'. Only 'tcp' is supported.\n", transport)
			os.Exit(1)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Println("\nShutting down...")
			cancel()
		}()

		if err := hub.RunServer(ctx, transport, address); err != nil {
			fmt.Printf("Hub server failed: %v\n", err)
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
