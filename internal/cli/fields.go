package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/not-emily/sage/pkg/sage"
	"github.com/not-emily/sage/pkg/sage/providers"
)

func runProviderFields(args []string) error {
	fs := flag.NewFlagSet("provider fields", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "output JSON")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: sage provider fields <provider> [flags]

List configuration fields required by a provider.

Providers: %s

Flags:
`, strings.Join(providers.List(), ", "))
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  sage provider fields openai
  sage provider fields anthropic
  sage provider fields ollama
  sage provider fields openai --json
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

	fields, err := sage.GetProviderFields(providerName)
	if err != nil {
		return fmt.Errorf("failed to get fields: %w", err)
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"provider": providerName,
			"fields":   fields,
		})
	}

	if len(fields) == 0 {
		fmt.Printf("%s requires no configuration fields.\n", providerName)
		return nil
	}

	fmt.Printf("Fields for %s:\n\n", providerName)
	for _, f := range fields {
		req := ""
		if f.Required {
			req = " (required)"
		}
		def := ""
		if f.Default != "" {
			def = fmt.Sprintf(" [default: %s]", f.Default)
		}
		fmt.Printf("  %s: %s%s%s\n", f.Key, f.Label, req, def)
	}

	return nil
}
