package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/not-emily/sage/pkg/sage"
	"github.com/not-emily/sage/pkg/sage/providers"
)

func runProviderModels(args []string) error {
	fs := flag.NewFlagSet("provider models", flag.ExitOnError)
	account := fs.String("account", "", "provider account to use (defaults to first configured)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: sage provider models <provider> [flags]

List available models from a provider.

Providers: %s

Flags:
`, strings.Join(providers.List(), ", "))
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  sage provider models openai
  sage provider models anthropic
  sage provider models ollama
  sage provider models openai --account=work
`)
	}

	fs.Parse(reorderArgs(args))

	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("provider name required")
	}
	providerName := fs.Arg(0)

	// Validate provider name
	if !providers.Exists(providerName) {
		return fmt.Errorf("unknown provider: %s\nSupported: %s", providerName, strings.Join(providers.List(), ", "))
	}

	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	models, err := client.ListModels(providerName, *account)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		fmt.Println("No models found.")
		return nil
	}

	fmt.Printf("Models for %s:\n\n", providerName)
	for _, m := range models {
		if m.Description != "" {
			fmt.Printf("  %s - %s\n", m.ID, m.Description)
		} else if m.Name != "" && m.Name != m.ID {
			fmt.Printf("  %s (%s)\n", m.ID, m.Name)
		} else {
			fmt.Printf("  %s\n", m.ID)
		}
	}

	return nil
}
