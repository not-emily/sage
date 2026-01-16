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
// Accounts maps account name -> field name -> value (non-secret fields only).
// Secret fields are stored in secrets.enc with key format "provider:account:field".
type ProviderConfig struct {
	Accounts map[string]map[string]string `json:"accounts"`
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
		// Check if this is an old format config
		if isOldConfigFormat(data) {
			return nil, fmt.Errorf("config format has changed in v0.3.0. Please remove ~/.config/sage/config.json and re-add your providers with 'sage provider add <provider>'")
		}
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}

	// Initialize maps if nil
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	// Initialize account maps if nil
	for name, pc := range cfg.Providers {
		if pc.Accounts == nil {
			pc.Accounts = make(map[string]map[string]string)
			cfg.Providers[name] = pc
		}
	}

	return &cfg, nil
}

// isOldConfigFormat detects the pre-v0.3.0 config format where accounts
// was an array of strings or had a base_url at provider level.
func isOldConfigFormat(data []byte) bool {
	var raw struct {
		Providers map[string]json.RawMessage `json:"providers"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return false
	}

	for _, providerData := range raw.Providers {
		// Try to detect old format: {"accounts": ["default"], "base_url": "..."}
		var oldFormat struct {
			Accounts []string `json:"accounts"`
			BaseURL  string   `json:"base_url"`
		}
		if err := json.Unmarshal(providerData, &oldFormat); err == nil {
			// If accounts is an array of strings, it's old format
			if len(oldFormat.Accounts) > 0 {
				return true
			}
			// If base_url exists at provider level, it's old format
			if oldFormat.BaseURL != "" {
				return true
			}
		}
	}
	return false
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

// GetAccountFields returns all fields for an account, merging config and secrets.
// The fieldDefs parameter specifies which fields are secrets.
func (c *Config) GetAccountFields(provider, account string, fieldDefs []FieldDef) (map[string]string, error) {
	providerConfig, ok := c.Providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not configured: %s", provider)
	}

	accountFields, ok := providerConfig.Accounts[account]
	if !ok {
		return nil, fmt.Errorf("account not found: %s:%s", provider, account)
	}

	// Start with config fields (non-secrets)
	result := make(map[string]string)
	for k, v := range accountFields {
		result[k] = v
	}

	// Add secret fields
	for _, fd := range fieldDefs {
		if fd.Secret {
			value, err := GetSecretField(provider, account, fd.Key)
			if err == nil {
				result[fd.Key] = value
			}
		}
	}

	return result, nil
}

// SetAccountFields stores fields for an account, splitting between config and secrets.
// The fieldDefs parameter specifies which fields are secrets.
func (c *Config) SetAccountFields(provider, account string, fields map[string]string, fieldDefs []FieldDef) error {
	// Build a set of secret field keys for quick lookup
	secretFields := make(map[string]bool)
	for _, fd := range fieldDefs {
		if fd.Secret {
			secretFields[fd.Key] = true
		}
	}

	// Ensure provider config exists
	if _, ok := c.Providers[provider]; !ok {
		c.Providers[provider] = ProviderConfig{
			Accounts: make(map[string]map[string]string),
		}
	}

	// Ensure accounts map exists
	providerConfig := c.Providers[provider]
	if providerConfig.Accounts == nil {
		providerConfig.Accounts = make(map[string]map[string]string)
	}

	// Initialize account fields map
	accountFields := make(map[string]string)

	// Split fields between config and secrets
	for key, value := range fields {
		if secretFields[key] {
			// Store in secrets
			if err := SetSecretField(provider, account, key, value); err != nil {
				return fmt.Errorf("failed to store secret %s: %w", key, err)
			}
		} else {
			// Store in config
			accountFields[key] = value
		}
	}

	providerConfig.Accounts[account] = accountFields
	c.Providers[provider] = providerConfig

	return c.Save()
}

// FieldDef describes a field for the purpose of storage (used by SetAccountFields/GetAccountFields).
type FieldDef struct {
	Key    string
	Secret bool
}

// HasAccount checks if an account exists for a provider.
func (c *Config) HasAccount(provider, account string) bool {
	providerConfig, ok := c.Providers[provider]
	if !ok {
		return false
	}
	_, exists := providerConfig.Accounts[account]
	return exists
}

// ListAccounts returns all account names for a provider.
func (c *Config) ListAccounts(provider string) []string {
	providerConfig, ok := c.Providers[provider]
	if !ok {
		return nil
	}
	accounts := make([]string, 0, len(providerConfig.Accounts))
	for name := range providerConfig.Accounts {
		accounts = append(accounts, name)
	}
	return accounts
}

// RemoveAccount removes an account from the config.
// Note: This does NOT remove secrets - call DeleteAccountSecrets separately.
func (c *Config) RemoveAccount(provider, account string) error {
	providerConfig, ok := c.Providers[provider]
	if !ok {
		return fmt.Errorf("provider not configured: %s", provider)
	}

	if _, ok := providerConfig.Accounts[account]; !ok {
		return fmt.Errorf("account not found: %s:%s", provider, account)
	}

	delete(providerConfig.Accounts, account)
	c.Providers[provider] = providerConfig

	return c.Save()
}
