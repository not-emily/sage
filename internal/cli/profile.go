package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/not-emily/sage/pkg/sage"
)

func runProfile(args []string) error {
	if len(args) == 0 {
		return showProfileHelp()
	}

	switch args[0] {
	case "list":
		return runProfileList(args[1:])
	case "add":
		return runProfileAdd(args[1:])
	case "remove":
		return runProfileRemove(args[1:])
	case "set-default":
		return runProfileSetDefault(args[1:])
	case "help", "-h", "--help":
		return showProfileHelp()
	default:
		return fmt.Errorf("unknown profile command: %s\nRun 'sage profile help' for usage", args[0])
	}
}

func showProfileHelp() error {
	help := `Usage: sage profile <command> [flags]

Commands:
  list        List configured profiles
  add         Add a profile
  remove      Remove a profile
  set-default Set the default profile

Examples:
  sage profile list
  sage profile add default --provider=openai --model=gpt-4o
  sage profile add fast --provider=anthropic --model=claude-3-5-haiku-latest
  sage profile set-default fast
  sage profile remove default
`
	fmt.Print(help)
	return nil
}

func runProfileList(args []string) error {
	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	profiles := client.ListProfiles()
	defaultProfile := client.GetDefaultProfile()

	if len(profiles) == 0 {
		fmt.Println("No profiles configured.")
		fmt.Println("\nRun 'sage profile add <name> --provider=X --model=Y' to create one.")
		return nil
	}

	for _, p := range profiles {
		marker := ""
		if p.Name == defaultProfile {
			marker = " (default)"
		}
		fmt.Printf("%s%s\n", p.Name, marker)
		fmt.Printf("  provider: %s\n", p.Provider)
		fmt.Printf("  account:  %s\n", p.Account)
		fmt.Printf("  model:    %s\n", p.Model)
	}
	return nil
}

func runProfileAdd(args []string) error {
	fs := flag.NewFlagSet("profile add", flag.ExitOnError)
	provider := fs.String("provider", "", "provider name (required)")
	account := fs.String("account", "default", "provider account")
	model := fs.String("model", "", "model name (required)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: sage profile add <name> --provider=X --model=Y [--account=Z]

Create a profile that binds a provider account to a specific model.

Flags:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  sage profile add default --provider=openai --model=gpt-4o
  sage profile add fast --provider=anthropic --model=claude-3-5-haiku-latest
  sage profile add local --provider=ollama --model=llama3.2 --account=default
`)
	}

	fs.Parse(reorderArgs(args))

	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("profile name required")
	}
	profileName := fs.Arg(0)

	if *provider == "" {
		return fmt.Errorf("--provider is required")
	}
	if *model == "" {
		return fmt.Errorf("--model is required")
	}

	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	// Validate provider account exists
	if !client.HasProviderAccount(*provider, *account) {
		return fmt.Errorf("provider account %s:%s not configured\nRun 'sage provider add %s' first", *provider, *account, *provider)
	}

	profile := sage.Profile{
		Name:     profileName,
		Provider: *provider,
		Account:  *account,
		Model:    *model,
	}

	if err := client.AddProfile(profileName, profile); err != nil {
		return err
	}

	fmt.Printf("Profile '%s' created\n", profileName)
	return nil
}

func runProfileRemove(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sage profile remove <name>")
	}
	profileName := args[0]

	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	if err := client.RemoveProfile(profileName); err != nil {
		return err
	}

	fmt.Printf("Profile '%s' removed\n", profileName)
	return nil
}

func runProfileSetDefault(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sage profile set-default <name>")
	}
	profileName := args[0]

	client, err := sage.NewClient()
	if err != nil {
		return err
	}

	if err := client.SetDefaultProfile(profileName); err != nil {
		return err
	}

	fmt.Printf("Default profile set to '%s'\n", profileName)
	return nil
}
