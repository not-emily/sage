// Package cli implements the sage command-line interface.
package cli

import (
	"fmt"
)

// Version is set at build time.
var Version = "0.1.0"

// Run executes the CLI with the given arguments.
func Run(args []string) error {
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
		return showVersion()
	case "help", "-h", "--help":
		return showHelp()
	default:
		return fmt.Errorf("unknown command: %s\nRun 'sage help' for usage", args[0])
	}
}

func showVersion() error {
	fmt.Printf("sage v%s\n", Version)
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

