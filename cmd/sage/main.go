package main

import (
	"fmt"
	"os"

	"github.com/not-emily/sage/internal/cli"

	// Import providers to register them via init()
	_ "github.com/not-emily/sage/pkg/sage/providers"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
