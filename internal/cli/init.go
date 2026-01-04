package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/not-emily/sage/pkg/sage"
)

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: sage init

Initialize sage configuration.

Creates:
  ~/.config/sage/config.json   Configuration file
  ~/.config/sage/master.key    Encryption key for API secrets
  ~/.config/sage/secrets.enc   Encrypted secrets storage

`)
	}
	fs.Parse(args)

	// Get config directory
	configDir, err := sage.ConfigDir()
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if already initialized
	keyPath, err := sage.MasterKeyPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(keyPath); err == nil {
		fmt.Printf("Sage already initialized at %s\n", configDir)
		return nil
	}

	// Initialize secrets (creates master key)
	if err := sage.InitSecrets(); err != nil {
		return fmt.Errorf("failed to initialize secrets: %w", err)
	}

	// Create empty config
	config := &sage.Config{
		Providers: make(map[string]sage.ProviderConfig),
		Profiles:  make(map[string]sage.Profile),
	}
	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	fmt.Printf("Sage initialized at %s\n", configDir)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add a provider:  sage provider add openai")
	fmt.Println("  2. Add a profile:   sage profile add default --provider=openai --model=gpt-4o-mini")
	fmt.Println("  3. Set as default:  sage profile set-default default")
	fmt.Println("  4. Test it:         sage complete \"Hello, world!\"")

	return nil
}
