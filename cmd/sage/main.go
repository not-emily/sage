package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/not-emily/sage/internal/cli"

	// Import providers to register them via init()
	_ "github.com/not-emily/sage/pkg/sage/providers"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		if cli.JSONOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(map[string]string{"error": err.Error()})
		} else {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
}
