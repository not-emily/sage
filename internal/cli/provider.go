package cli

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/not-emily/sage/pkg/sage"
	"github.com/not-emily/sage/pkg/sage/providers"
)

func runProvider(args []string) error {
	if len(args) == 0 {
		return showProviderHelp()
	}

	switch args[0] {
	case "list":
		return runProviderList(args[1:])
	case "add":
		return runProviderAdd(args[1:])
	case "remove":
		return runProviderRemove(args[1:])
	case "models":
		return runProviderModels(args[1:])
	case "fields":
		return runProviderFields(args[1:])
	case "help", "-h", "--help":
		return showProviderHelp()
	default:
		return fmt.Errorf("unknown provider command: %s\nRun 'sage provider help' for usage", args[0])
	}
}

func showProviderHelp() error {
	help := `Usage: sage provider <command> [flags]

Commands:
  list      List configured providers and accounts
  add       Add a provider account
  remove    Remove a provider account
  models    List available models from a provider
  fields    List configuration fields for a provider

Examples:
  sage provider list
  sage provider add openai
  sage provider add openai --account=work
  sage provider add openai --api-key-env=OPENAI_API_KEY
  sage provider models openai
  sage provider fields openai
  sage provider remove openai --account=work

  # Scripting (fields via stdin):
  echo '{"api_key":"sk-..."}' | sage provider add openai --stdin --json
`
	fmt.Print(help)
	return nil
}

func runProviderList(args []string) error {
	fs := flag.NewFlagSet("provider list", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "output JSON")
	fs.Parse(args)

	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	providerList := client.ListProviders()

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"providers": providerList,
		})
	}

	if len(providerList) == 0 {
		fmt.Println("No providers configured.")
		fmt.Println("\nAvailable providers:", strings.Join(providers.List(), ", "))
		fmt.Println("\nRun 'sage provider add <name>' to add one.")
		return nil
	}

	for _, p := range providerList {
		fmt.Printf("%s:\n", p.Name)
		for _, account := range p.Accounts {
			fmt.Printf("  - %s\n", account)
		}
	}
	return nil
}

func runProviderAdd(args []string) error {
	fs := flag.NewFlagSet("provider add", flag.ExitOnError)
	account := fs.String("account", "default", "account name")
	apiKeyEnv := fs.String("api-key-env", "", "environment variable containing API key")
	baseURLFlag := fs.String("base-url", "", "custom base URL (overrides prompt)")
	stdinFlag := fs.Bool("stdin", false, "read fields as JSON from stdin (for scripting)")
	jsonOutput := fs.Bool("json", false, "output JSON")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: sage provider add <provider> [flags]

Add a provider account. You will be prompted for required fields.

Providers: %s

Flags:
`, strings.Join(providers.List(), ", "))
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  sage provider add openai
  sage provider add openai --account=work
  sage provider add openai --api-key-env=OPENAI_API_KEY
  sage provider add ollama --base-url=http://remote:11434

  # For scripting (fields as JSON via stdin):
  echo '{"api_key":"sk-..."}' | sage provider add openai --stdin --json
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

	// Get provider field definitions
	fieldDefs, err := sage.GetProviderFields(providerName)
	if err != nil {
		return err
	}

	// Sort fields: required first, then by original order
	sortedFields := make([]sage.ProviderField, len(fieldDefs))
	copy(sortedFields, fieldDefs)
	sortFields(sortedFields)

	// Build fields map
	fields := make(map[string]string)

	// If --stdin, read JSON fields from stdin first
	if *stdinFlag {
		stdinFields, err := readFieldsFromStdin()
		if err != nil {
			return fmt.Errorf("failed to read fields from stdin: %w", err)
		}
		for k, v := range stdinFields {
			fields[k] = v
		}
	}

	// Apply flag overrides (flags take precedence over stdin)
	if *apiKeyEnv != "" {
		apiKey := os.Getenv(*apiKeyEnv)
		if apiKey == "" {
			return fmt.Errorf("environment variable %s is not set", *apiKeyEnv)
		}
		fields["api_key"] = apiKey
	}
	if *baseURLFlag != "" {
		fields["base_url"] = *baseURLFlag
	}

	// Handle missing fields
	for _, f := range sortedFields {
		// Skip if already provided
		if _, ok := fields[f.Key]; ok {
			continue
		}

		// Apply default if available
		if f.Default != "" {
			fields[f.Key] = f.Default
			continue
		}

		// In stdin mode, don't prompt - error if required field is missing
		if *stdinFlag {
			if f.Required {
				return fmt.Errorf("missing required field: %s", f.Key)
			}
			continue
		}

		// Interactive mode: prompt for the field
		value, err := promptForField(f)
		if err != nil {
			return err
		}
		if value != "" {
			fields[f.Key] = value
		}
	}

	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	// Add the provider account with fields
	if err := client.AddProviderAccount(providerName, *account, fields); err != nil {
		return err
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"success":  true,
			"provider": providerName,
			"account":  *account,
		})
	}

	fmt.Printf("Added %s:%s\n", providerName, *account)
	return nil
}

// readFieldsFromStdin reads JSON field values from stdin.
func readFieldsFromStdin() (map[string]string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	var fields map[string]string
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return fields, nil
}

// sortFields sorts fields with required fields first, preserving original order within each group.
func sortFields(fields []sage.ProviderField) {
	// Simple stable sort: required fields first
	n := len(fields)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			// Swap if fields[j] is required and fields[i] is not
			if fields[j].Required && !fields[i].Required {
				fields[i], fields[j] = fields[j], fields[i]
			}
		}
	}
}

// promptForField prompts the user for a field value and returns it.
func promptForField(f sage.ProviderField) (string, error) {
	// Build prompt: "Label" or "Label (optional)" or "Label [default]" or "Label (optional) [default]"
	prompt := f.Label
	if !f.Required {
		prompt += " (optional)"
	}
	if f.Default != "" {
		prompt += fmt.Sprintf(" [%s]", f.Default)
	}
	prompt += ": "

	fmt.Print(prompt)

	value, err := readLine()
	if err != nil {
		return "", err
	}
	value = strings.TrimSpace(value)

	// Apply default if empty and default exists
	if value == "" && f.Default != "" {
		return f.Default, nil
	}

	return value, nil
}

func runProviderRemove(args []string) error {
	fs := flag.NewFlagSet("provider remove", flag.ExitOnError)
	account := fs.String("account", "default", "account name to remove")
	jsonOutput := fs.Bool("json", false, "output JSON")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: sage provider remove <provider> [flags]

Remove a provider account.

Flags:
`)
		fs.PrintDefaults()
	}

	fs.Parse(reorderArgs(args))

	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("provider name required")
	}
	providerName := fs.Arg(0)

	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	if err := client.RemoveProviderAccount(providerName, *account); err != nil {
		return err
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"success":  true,
			"provider": providerName,
			"account":  *account,
		})
	}

	fmt.Printf("Removed %s:%s\n", providerName, *account)
	return nil
}

// readLine reads a line from stdin.
func readLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\n"), nil
}

// reorderArgs moves flags before positional arguments.
// This allows "provider add openai --api-key-env=X" to work the same as
// "provider add --api-key-env=X openai".
func reorderArgs(args []string) []string {
	var flags, positional []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
		} else {
			positional = append(positional, arg)
		}
	}
	return append(flags, positional...)
}
