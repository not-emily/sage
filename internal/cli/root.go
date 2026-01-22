// Package cli implements the sage command-line interface.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
)

// Version is set at build time via ldflags.
var Version = "dev"

// getVersion returns the version, preferring ldflags but falling back to module info.
func getVersion() string {
	// If Version was set via ldflags, use it
	if Version != "dev" {
		return Version
	}

	// Otherwise try to get version from build info (for go install)
	if info, ok := debug.ReadBuildInfo(); ok {
		// Check for version in build settings (VCS info)
		var revision, modified string
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = setting.Value
			case "vcs.modified":
				modified = setting.Value
			}
		}

		// If we have a module version and it's not (devel), use it
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			version := info.Main.Version
			// Add revision info if available
			if revision != "" {
				version += " (" + revision[:7] + ")"
			}
			if modified == "true" {
				version += "-dirty"
			}
			return version
		}

		// Fall back to revision if available
		if revision != "" {
			rev := revision
			if len(rev) > 7 {
				rev = rev[:7]
			}
			if modified == "true" {
				rev += "-dirty"
			}
			return rev
		}
	}

	return "dev"
}

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
	case "update":
		return runUpdate(args[1:])
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

	version := getVersion()

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]string{
			"version": version,
		})
	}

	fmt.Printf("sage %s\n", version)
	return nil
}

func runUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: sage update

Updates sage to the latest version using 'go install'.

Note: This only works if sage was installed via 'go install'.
`)
	}

	fs.Parse(args)

	fmt.Println("Updating sage to latest version...")
	fmt.Println("Running: go install github.com/not-emily/sage/cmd/sage@latest")

	// Use exec to run go install
	cmd := exec.Command("go", "install", "github.com/not-emily/sage/cmd/sage@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update sage: %w", err)
	}

	fmt.Println("\nSage updated successfully!")
	fmt.Println("Run 'sage version' to see the new version.")

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
  update      Update sage to the latest version
  help        Show this help

Run 'sage <command> --help' for command-specific help.
`
	fmt.Print(help)
	return nil
}

