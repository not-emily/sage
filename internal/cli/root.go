// Package cli implements the sage command-line interface.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// Version is set at build time via ldflags.
var Version = "dev"

// JSONOutput is set to true when --json flag is detected.
// Used by main.go to format errors as JSON.
var JSONOutput bool

// Run executes the CLI with the given arguments.
func Run(args []string) error {
	// Pre-scan for --json flag so we can format errors correctly
	for _, arg := range args {
		if arg == "--json" || arg == "-json" {
			JSONOutput = true
			break
		}
	}

	if len(args) == 0 {
		return showHelp()
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "complete":
		return runComplete(args[1:])
	case "provider":
		return runProvider(args[1:])
	case "profile":
		return runProfile(args[1:])
	case "version":
		return runVersion(args[1:])
	case "help", "-h", "--help":
		return showHelp()
	default:
		return fmt.Errorf("unknown command: %s\nRun 'sage help' for usage", args[0])
	}
}

func runVersion(args []string) error {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "output JSON")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: sage version [--json]\n")
	}

	fs.Parse(args)

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]string{
			"version": Version,
		})
	}

	fmt.Printf("sage %s\n", Version)
	return nil
}

func showHelp() error {
	help := `sage - unified CLI for LLM providers

Usage:
  sage <command> [flags]

Commands:
  init        Initialize sage (create config, generate master key)
  complete    Send a completion request
  provider    Manage provider accounts
  profile     Manage profiles
  version     Show version
  help        Show this help

Run 'sage <command> --help' for command-specific help.
`
	fmt.Print(help)
	return nil
}

