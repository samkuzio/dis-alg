package main

import (
	"fmt"
	"os"

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
		
		hub.Run(transport, address)
	case "terminal":
		terminal.Run()
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		fmt.Println("Usage: dis-alg <hub|terminal> [args...]")
		os.Exit(1)
	}
}
