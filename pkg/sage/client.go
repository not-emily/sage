package sage

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/not-emily/sage/pkg/sage/providers"
)

// Client provides the high-level API for LLM completions.
type Client struct {
	config     *Config
	secrets    map[string]string
	HTTPClient *http.Client // Optional HTTP client; nil uses defaultHTTPClient()
}

// defaultHTTPClient returns an HTTP client with sensible timeouts.
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Minute, // Reasonable for LLM requests
		Transport: &http.Transport{
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}
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

	client := &Client{
		config:  config,
		secrets: secrets,
	}

	// Set default HTTP client if not provided
	if client.HTTPClient == nil {
		client.HTTPClient = defaultHTTPClient()
	}

	return client, nil
}

// Complete sends a completion request using the specified profile.
// If profileName is empty, the default profile is used.
// ctx is used for cancellation and timeout control.
func (c *Client) Complete(ctx context.Context, profileName string, req Request) (*Response, error) {
	providerReq, err := c.buildProviderRequest(profileName, req)
	if err != nil {
		return nil, err
	}

	profile, _ := c.config.GetProfile(profileName)
	provider, err := providers.Get(profile.Provider)
	if err != nil {
		return nil, err
	}

	providerResp, err := provider.Complete(ctx, providerReq)
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
// ctx is used for cancellation; allows stopping streaming mid-request.
func (c *Client) CompleteStream(ctx context.Context, profileName string, req Request) (<-chan Chunk, error) {
	providerReq, err := c.buildProviderRequest(profileName, req)
	if err != nil {
		return nil, err
	}

	profile, _ := c.config.GetProfile(profileName)
	provider, err := providers.Get(profile.Provider)
	if err != nil {
		return nil, err
	}

	providerCh, err := provider.CompleteStream(ctx, providerReq)
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

	// Get API key from secrets (new format: provider:account:field)
	apiKey := c.secrets[profile.Provider+":"+profile.Account+":api_key"]

	// Get base_url from config (non-secret field)
	var baseURL string
	if providerConfig, ok := c.config.Providers[profile.Provider]; ok {
		if accountFields, ok := providerConfig.Accounts[profile.Account]; ok {
			baseURL = accountFields["base_url"]
		}
	}

	return providers.Request{
		Model:      profile.Model,
		System:     req.System,
		Prompt:     req.Prompt,
		MaxTokens:  req.MaxTokens,
		APIKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: c.HTTPClient,
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

// AddProviderAccount adds a provider account with the specified fields.
// Fields should include all required fields for the provider (e.g., "api_key", "base_url").
// Use GetProviderFields to discover what fields a provider requires.
func (c *Client) AddProviderAccount(providerName, account string, fields map[string]string) error {
	// Validate provider exists
	if !providers.Exists(providerName) {
		return fmt.Errorf("unknown provider: %s", providerName)
	}

	// Get provider to access field definitions
	provider, _ := providers.Get(providerName)
	fieldDefs := provider.Fields()

	// Validate required fields are present
	for _, f := range fieldDefs {
		if f.Required {
			val, ok := fields[f.Key]
			if !ok || val == "" {
				return fmt.Errorf("required field %q missing for provider %s", f.Key, providerName)
			}
		}
	}

	// Apply defaults for missing optional fields
	fieldsWithDefaults := make(map[string]string)
	for k, v := range fields {
		fieldsWithDefaults[k] = v
	}
	for _, f := range fieldDefs {
		if f.Default != "" {
			if _, ok := fieldsWithDefaults[f.Key]; !ok {
				fieldsWithDefaults[f.Key] = f.Default
			}
		}
	}

	// Convert provider fields to FieldDef for storage
	storageDefs := make([]FieldDef, len(fieldDefs))
	for i, f := range fieldDefs {
		storageDefs[i] = FieldDef{Key: f.Key, Secret: f.Secret}
	}

	// Store using new format
	if err := c.config.SetAccountFields(providerName, account, fieldsWithDefaults, storageDefs); err != nil {
		return err
	}

	// Update in-memory secrets cache for secret fields
	for _, f := range fieldDefs {
		if f.Secret {
			if val, ok := fieldsWithDefaults[f.Key]; ok {
				c.secrets[providerName+":"+account+":"+f.Key] = val
			}
		}
	}

	return nil
}

// RemoveProviderAccount removes a provider account and all its secrets.
func (c *Client) RemoveProviderAccount(providerName, account string) error {
	// Remove from config
	if err := c.config.RemoveAccount(providerName, account); err != nil {
		return err
	}

	// Remove all secrets for this account
	if err := DeleteAccountSecrets(providerName, account); err != nil {
		return err
	}

	// Update in-memory secrets cache (remove all keys with this prefix)
	prefix := providerName + ":" + account + ":"
	for key := range c.secrets {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			delete(c.secrets, key)
		}
	}

	return nil
}

// ListProviders returns all configured providers with their accounts.
func (c *Client) ListProviders() []ProviderInfo {
	infos := make([]ProviderInfo, 0, len(c.config.Providers))
	for name, config := range c.config.Providers {
		// Extract account names from the map
		accounts := make([]string, 0, len(config.Accounts))
		for accountName := range config.Accounts {
			accounts = append(accounts, accountName)
		}
		sort.Strings(accounts)

		infos = append(infos, ProviderInfo{
			Name:     name,
			Accounts: accounts,
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
	return c.config.HasAccount(providerName, account)
}

// --- Model Discovery ---

// ModelInfo describes an available model.
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListModels returns available models from a provider.
// If account is empty, uses the first configured account.
// ctx is used for cancellation and timeout control.
func (c *Client) ListModels(ctx context.Context, providerName, account string) ([]ModelInfo, error) {
	provider, err := providers.Get(providerName)
	if err != nil {
		return nil, err
	}

	// Get account fields
	var apiKey, baseURL string
	providerConfig, ok := c.config.Providers[providerName]
	if ok {
		// Use specified account or first available
		if account == "" && len(providerConfig.Accounts) > 0 {
			// Get first account name
			for name := range providerConfig.Accounts {
				account = name
				break
			}
		}
		if account != "" {
			// Get API key from secrets (new format)
			apiKey = c.secrets[providerName+":"+account+":api_key"]
			// Get base_url from config
			if accountFields, ok := providerConfig.Accounts[account]; ok {
				baseURL = accountFields["base_url"]
			}
		}
	}

	providerModels, err := provider.ListModels(ctx, apiKey, baseURL)
	if err != nil {
		return nil, err
	}

	// Convert provider models to sage models
	models := make([]ModelInfo, len(providerModels))
	for i, m := range providerModels {
		models[i] = ModelInfo{
			ID:          m.ID,
			Name:        m.Name,
			Description: m.Description,
		}
	}

	return models, nil
}

// ListAvailableProviders returns all provider names that sage supports.
func ListAvailableProviders() []string {
	return providers.List()
}

// ProviderField re-exports the providers.ProviderField type for public API.
type ProviderField = providers.ProviderField

// GetProviderFields returns the field requirements for a provider.
// This is a package-level function since it doesn't require a configured client.
// Use this to discover what fields are needed before calling AddProviderAccount.
func GetProviderFields(providerName string) ([]ProviderField, error) {
	p, err := providers.Get(providerName)
	if err != nil {
		return nil, err
	}
	return p.Fields(), nil
}
