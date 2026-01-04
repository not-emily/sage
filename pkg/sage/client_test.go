package sage

import (
	"testing"
)

func setupTestClient(t *testing.T) *Client {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Initialize secrets
	if err := InitSecrets(); err != nil {
		t.Fatalf("InitSecrets() error = %v", err)
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

func TestNewClient(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Initialize secrets first
	if err := InitSecrets(); err != nil {
		t.Fatalf("InitSecrets() error = %v", err)
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.config == nil {
		t.Error("client.config is nil")
	}

	if client.secrets == nil {
		t.Error("client.secrets is nil")
	}
}

func TestClient_ProfileManagement(t *testing.T) {
	client := setupTestClient(t)

	// Add a profile
	profile := Profile{
		Provider: "openai",
		Account:  "default",
		Model:    "gpt-4o-mini",
	}

	if err := client.AddProfile("test", profile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Get it back
	got, err := client.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if got.Model != "gpt-4o-mini" {
		t.Errorf("Model = %q, want %q", got.Model, "gpt-4o-mini")
	}

	// List profiles
	profiles := client.ListProfiles()
	if len(profiles) != 1 {
		t.Errorf("ListProfiles() count = %d, want 1", len(profiles))
	}

	// Set as default
	if err := client.SetDefaultProfile("test"); err != nil {
		t.Fatalf("SetDefaultProfile() error = %v", err)
	}

	if client.GetDefaultProfile() != "test" {
		t.Errorf("GetDefaultProfile() = %q, want %q", client.GetDefaultProfile(), "test")
	}

	// Can't remove default profile
	if err := client.RemoveProfile("test"); err == nil {
		t.Error("RemoveProfile() should error for default profile")
	}

	// Add another and remove first
	if err := client.AddProfile("other", profile); err != nil {
		t.Fatalf("AddProfile(other) error = %v", err)
	}
	if err := client.SetDefaultProfile("other"); err != nil {
		t.Fatalf("SetDefaultProfile(other) error = %v", err)
	}

	// Now we can remove test
	if err := client.RemoveProfile("test"); err != nil {
		t.Fatalf("RemoveProfile() error = %v", err)
	}

	profiles = client.ListProfiles()
	if len(profiles) != 1 {
		t.Errorf("ListProfiles() after remove = %d, want 1", len(profiles))
	}
}

func TestClient_ProviderAccountManagement(t *testing.T) {
	client := setupTestClient(t)

	// Add a provider account
	if err := client.AddProviderAccount("openai", "work", "sk-test-key"); err != nil {
		t.Fatalf("AddProviderAccount() error = %v", err)
	}

	// Check it exists
	if !client.HasProviderAccount("openai", "work") {
		t.Error("HasProviderAccount() should return true")
	}

	// List providers
	providers := client.ListProviders()
	if len(providers) != 1 {
		t.Fatalf("ListProviders() count = %d, want 1", len(providers))
	}

	if providers[0].Name != "openai" {
		t.Errorf("Provider name = %q, want %q", providers[0].Name, "openai")
	}

	if len(providers[0].Accounts) != 1 || providers[0].Accounts[0] != "work" {
		t.Errorf("Provider accounts = %v, want [work]", providers[0].Accounts)
	}

	// Add another account to same provider
	if err := client.AddProviderAccount("openai", "personal", "sk-personal-key"); err != nil {
		t.Fatalf("AddProviderAccount(personal) error = %v", err)
	}

	providers = client.ListProviders()
	if len(providers[0].Accounts) != 2 {
		t.Errorf("Provider accounts count = %d, want 2", len(providers[0].Accounts))
	}

	// Remove an account
	if err := client.RemoveProviderAccount("openai", "work"); err != nil {
		t.Fatalf("RemoveProviderAccount() error = %v", err)
	}

	if client.HasProviderAccount("openai", "work") {
		t.Error("HasProviderAccount() should return false after remove")
	}
}

func TestClient_AddProfile_InvalidProvider(t *testing.T) {
	client := setupTestClient(t)

	profile := Profile{
		Provider: "invalid-provider",
		Account:  "default",
		Model:    "model",
	}

	err := client.AddProfile("test", profile)
	if err == nil {
		t.Error("AddProfile() with invalid provider should error")
	}
}

func TestClient_AddProviderAccount_InvalidProvider(t *testing.T) {
	client := setupTestClient(t)

	err := client.AddProviderAccount("invalid-provider", "default", "key")
	if err == nil {
		t.Error("AddProviderAccount() with invalid provider should error")
	}
}

func TestClient_GetProfile_UsesDefault(t *testing.T) {
	client := setupTestClient(t)

	// Add and set default
	profile := Profile{
		Provider: "openai",
		Account:  "default",
		Model:    "gpt-4o",
	}
	client.AddProfile("myprofile", profile)
	client.SetDefaultProfile("myprofile")

	// Get with empty name should return default
	got, err := client.GetProfile("")
	if err != nil {
		t.Fatalf("GetProfile('') error = %v", err)
	}

	if got.Name != "myprofile" {
		t.Errorf("GetProfile('').Name = %q, want %q", got.Name, "myprofile")
	}
}

func TestClient_UpdateExistingAccount(t *testing.T) {
	client := setupTestClient(t)

	// Add account
	if err := client.AddProviderAccount("openai", "work", "old-key"); err != nil {
		t.Fatalf("AddProviderAccount() error = %v", err)
	}

	// Update with new key (same account name)
	if err := client.AddProviderAccount("openai", "work", "new-key"); err != nil {
		t.Fatalf("AddProviderAccount() update error = %v", err)
	}

	// Should still only have one account
	providers := client.ListProviders()
	if len(providers[0].Accounts) != 1 {
		t.Errorf("Accounts count = %d, want 1 (should update, not duplicate)", len(providers[0].Accounts))
	}
}
