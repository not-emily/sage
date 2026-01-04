package sage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the sage configuration.
type Config struct {
	Providers      map[string]ProviderConfig `json:"providers"`
	Profiles       map[string]Profile        `json:"profiles"`
	DefaultProfile string                    `json:"default_profile"`
}

// ProviderConfig stores provider-specific settings.
type ProviderConfig struct {
	Accounts []string `json:"accounts"`
	BaseURL  string   `json:"base_url,omitempty"`
}

// ConfigDir returns the sage config directory path, creating it if needed.
// Default: ~/.config/sage/
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	dir := filepath.Join(home, ".config", "sage")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}

	return dir, nil
}

// ConfigPath returns the path to config.json.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadConfig reads config from ~/.config/sage/config.json.
// Returns an empty config if the file doesn't exist.
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Return empty config if file doesn't exist
			return &Config{
				Providers: make(map[string]ProviderConfig),
				Profiles:  make(map[string]Profile),
			}, nil
		}
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}

	// Initialize maps if nil
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return &cfg, nil
}

// Save writes the config to ~/.config/sage/config.json.
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	return nil
}

// GetProfile returns a profile by name, or the default profile if name is empty.
func (c *Config) GetProfile(name string) (*Profile, error) {
	if name == "" {
		name = c.DefaultProfile
	}
	if name == "" {
		return nil, errors.New("no profile specified and no default set")
	}

	profile, ok := c.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", name)
	}

	profile.Name = name
	return &profile, nil
}

// GetProvider returns provider config by name.
func (c *Config) GetProvider(name string) (*ProviderConfig, error) {
	provider, ok := c.Providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return &provider, nil
}
