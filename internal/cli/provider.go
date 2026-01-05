package cli

import (
	"bufio"
	"flag"
	"fmt"
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

Examples:
  sage provider list
  sage provider add openai
  sage provider add openai --account=work
  sage provider add openai --api-key-env=OPENAI_API_KEY
  sage provider models openai
  sage provider remove openai --account=work
`
	fmt.Print(help)
	return nil
}

func runProviderList(args []string) error {
	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	providerList := client.ListProviders()
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
		if p.BaseURL != "" {
			fmt.Printf("  base_url: %s\n", p.BaseURL)
		}
	}
	return nil
}

func runProviderAdd(args []string) error {
	fs := flag.NewFlagSet("provider add", flag.ExitOnError)
	account := fs.String("account", "default", "account name")
	apiKeyEnv := fs.String("api-key-env", "", "environment variable containing API key")
	baseURL := fs.String("base-url", "", "custom base URL (for proxies or compatible APIs)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: sage provider add <provider> [flags]

Add a provider account with an API key.

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

	// Get API key (optional for ollama)
	var apiKey string
	if providerName == "ollama" && *apiKeyEnv == "" {
		// Ollama typically doesn't need an API key
		fmt.Print("Enter API key (press Enter to skip for local Ollama): ")
		key, err := readLine()
		if err != nil {
			return err
		}
		apiKey = strings.TrimSpace(key)
	} else if *apiKeyEnv != "" {
		apiKey = os.Getenv(*apiKeyEnv)
		if apiKey == "" {
			return fmt.Errorf("environment variable %s is not set", *apiKeyEnv)
		}
	} else {
		// Interactive prompt
		fmt.Print("Enter API key: ")
		key, err := readLine()
		if err != nil {
			return err
		}
		apiKey = strings.TrimSpace(key)
		if apiKey == "" && providerName != "ollama" {
			return fmt.Errorf("API key required for %s", providerName)
		}
	}

	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	// Add the provider account
	if err := client.AddProviderAccount(providerName, *account, apiKey); err != nil {
		return err
	}

	// Update base URL if provided
	if *baseURL != "" {
		// Need to update config directly for base URL
		config, err := sage.LoadConfig()
		if err != nil {
			return err
		}
		providerConfig := config.Providers[providerName]
		providerConfig.BaseURL = *baseURL
		config.Providers[providerName] = providerConfig
		if err := config.Save(); err != nil {
			return err
		}
	}

	fmt.Printf("Added %s:%s\n", providerName, *account)
	return nil
}

func runProviderRemove(args []string) error {
	fs := flag.NewFlagSet("provider remove", flag.ExitOnError)
	account := fs.String("account", "default", "account name to remove")

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
