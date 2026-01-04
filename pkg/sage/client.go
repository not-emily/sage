package sage

import (
	"fmt"
	"sort"

	"github.com/not-emily/sage/pkg/sage/providers"
)

// Client provides the high-level API for LLM completions.
type Client struct {
	config  *Config
	secrets map[string]string
}

// NewClient creates a new client, loading config and secrets.
func NewClient() (*Client, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	secrets, err := LoadSecrets()
	if err != nil {
		return nil, fmt.Errorf("failed to load secrets: %w", err)
	}

	return &Client{
		config:  config,
		secrets: secrets,
	}, nil
}

// Complete sends a completion request using the specified profile.
// If profileName is empty, the default profile is used.
func (c *Client) Complete(profileName string, req Request) (*Response, error) {
	providerReq, err := c.buildProviderRequest(profileName, req)
	if err != nil {
		return nil, err
	}

	profile, _ := c.config.GetProfile(profileName)
	provider, err := providers.Get(profile.Provider)
	if err != nil {
		return nil, err
	}

	providerResp, err := provider.Complete(providerReq)
	if err != nil {
		return nil, err
	}

	return &Response{
		Content: providerResp.Content,
		Model:   providerResp.Model,
		Usage: Usage{
			PromptTokens:     providerResp.Usage.PromptTokens,
			CompletionTokens: providerResp.Usage.CompletionTokens,
		},
	}, nil
}

// CompleteStream sends a streaming completion request.
// If profileName is empty, the default profile is used.
func (c *Client) CompleteStream(profileName string, req Request) (<-chan Chunk, error) {
	providerReq, err := c.buildProviderRequest(profileName, req)
	if err != nil {
		return nil, err
	}

	profile, _ := c.config.GetProfile(profileName)
	provider, err := providers.Get(profile.Provider)
	if err != nil {
		return nil, err
	}

	providerCh, err := provider.CompleteStream(providerReq)
	if err != nil {
		return nil, err
	}

	// Convert provider chunks to sage chunks
	ch := make(chan Chunk)
	go func() {
		defer close(ch)
		for providerChunk := range providerCh {
			ch <- Chunk{
				Content: providerChunk.Content,
				Done:    providerChunk.Done,
				Error:   providerChunk.Error,
			}
		}
	}()

	return ch, nil
}

// buildProviderRequest creates a provider request from a sage request.
func (c *Client) buildProviderRequest(profileName string, req Request) (providers.Request, error) {
	profile, err := c.config.GetProfile(profileName)
	if err != nil {
		return providers.Request{}, err
	}

	// Get API key for this provider:account
	secretKey := profile.Provider + ":" + profile.Account
	apiKey := c.secrets[secretKey]

	// Get provider config for BaseURL
	var baseURL string
	if providerConfig, ok := c.config.Providers[profile.Provider]; ok {
		baseURL = providerConfig.BaseURL
	}

	return providers.Request{
		Model:     profile.Model,
		System:    req.System,
		Prompt:    req.Prompt,
		MaxTokens: req.MaxTokens,
		APIKey:    apiKey,
		BaseURL:   baseURL,
	}, nil
}

// --- Profile Management ---

// GetDefaultProfile returns the name of the default profile.
func (c *Client) GetDefaultProfile() string {
	return c.config.DefaultProfile
}

// GetProfile returns a profile by name. If name is empty, returns the default.
func (c *Client) GetProfile(name string) (*Profile, error) {
	return c.config.GetProfile(name)
}

// ListProfiles returns all configured profiles.
func (c *Client) ListProfiles() []Profile {
	profiles := make([]Profile, 0, len(c.config.Profiles))
	for name, p := range c.config.Profiles {
		p.Name = name
		profiles = append(profiles, p)
	}
	// Sort by name for consistent ordering
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})
	return profiles
}

// AddProfile adds or updates a profile.
func (c *Client) AddProfile(name string, p Profile) error {
	// Validate provider exists
	if !providers.Exists(p.Provider) {
		return fmt.Errorf("unknown provider: %s", p.Provider)
	}

	c.config.Profiles[name] = p
	return c.config.Save()
}

// RemoveProfile removes a profile.
func (c *Client) RemoveProfile(name string) error {
	if _, ok := c.config.Profiles[name]; !ok {
		return fmt.Errorf("profile not found: %s", name)
	}

	// Don't allow removing the default profile
	if c.config.DefaultProfile == name {
		return fmt.Errorf("cannot remove default profile: %s", name)
	}

	delete(c.config.Profiles, name)
	return c.config.Save()
}

// SetDefaultProfile sets the default profile.
func (c *Client) SetDefaultProfile(name string) error {
	if _, ok := c.config.Profiles[name]; !ok {
		return fmt.Errorf("profile not found: %s", name)
	}

	c.config.DefaultProfile = name
	return c.config.Save()
}

// --- Provider Account Management ---

// AddProviderAccount adds a provider account with an API key.
func (c *Client) AddProviderAccount(providerName, account, apiKey string) error {
	// Validate provider exists
	if !providers.Exists(providerName) {
		return fmt.Errorf("unknown provider: %s", providerName)
	}

	// Add account to provider config
	providerConfig := c.config.Providers[providerName]

	// Check if account already exists
	for _, a := range providerConfig.Accounts {
		if a == account {
			// Account exists, just update the key
			c.secrets[providerName+":"+account] = apiKey
			return SaveSecrets(c.secrets)
		}
	}

	// Add new account
	providerConfig.Accounts = append(providerConfig.Accounts, account)
	c.config.Providers[providerName] = providerConfig

	// Store the API key
	c.secrets[providerName+":"+account] = apiKey

	// Save both config and secrets
	if err := c.config.Save(); err != nil {
		return err
	}
	return SaveSecrets(c.secrets)
}

// RemoveProviderAccount removes a provider account and its API key.
func (c *Client) RemoveProviderAccount(providerName, account string) error {
	providerConfig, ok := c.config.Providers[providerName]
	if !ok {
		return fmt.Errorf("provider not configured: %s", providerName)
	}

	// Find and remove account
	found := false
	newAccounts := make([]string, 0, len(providerConfig.Accounts))
	for _, a := range providerConfig.Accounts {
		if a == account {
			found = true
		} else {
			newAccounts = append(newAccounts, a)
		}
	}

	if !found {
		return fmt.Errorf("account not found: %s:%s", providerName, account)
	}

	providerConfig.Accounts = newAccounts
	c.config.Providers[providerName] = providerConfig

	// Remove the secret
	delete(c.secrets, providerName+":"+account)

	// Save both
	if err := c.config.Save(); err != nil {
		return err
	}
	return SaveSecrets(c.secrets)
}

// ListProviders returns all configured providers with their accounts.
func (c *Client) ListProviders() []ProviderInfo {
	infos := make([]ProviderInfo, 0, len(c.config.Providers))
	for name, config := range c.config.Providers {
		infos = append(infos, ProviderInfo{
			Name:     name,
			Accounts: config.Accounts,
			BaseURL:  config.BaseURL,
		})
	}
	// Sort by name for consistent ordering
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})
	return infos
}

// HasProviderAccount checks if a provider account exists.
func (c *Client) HasProviderAccount(providerName, account string) bool {
	providerConfig, ok := c.config.Providers[providerName]
	if !ok {
		return false
	}
	for _, a := range providerConfig.Accounts {
		if a == account {
			return true
		}
	}
	return false
}
