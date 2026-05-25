package hub

import (
	"fmt"
)

func Run(transport, address string) {
	fmt.Printf("Running in hub node mode\n")
	fmt.Printf("Transport: %s\n", transport)
	fmt.Printf("Listening on: %s\n", address)
}
