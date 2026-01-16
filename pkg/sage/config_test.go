package sage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NoFile(t *testing.T) {
	// Use temp dir to avoid touching real config
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Providers == nil {
		t.Error("Providers map should be initialized")
	}
	if cfg.Profiles == nil {
		t.Error("Profiles map should be initialized")
	}
	if cfg.DefaultProfile != "" {
		t.Errorf("DefaultProfile = %q, want empty", cfg.DefaultProfile)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Create a config with new format (accounts as nested maps)
	original := &Config{
		Providers: map[string]ProviderConfig{
			"openai": {
				Accounts: map[string]map[string]string{
					"default": {"base_url": "https://api.openai.com/v1"},
					"work":    {"base_url": "https://api.openai.com/v1"},
				},
			},
			"ollama": {
				Accounts: map[string]map[string]string{
					"local": {"base_url": "http://localhost:11434"},
				},
			},
		},
		Profiles: map[string]Profile{
			"small_brain": {
				Provider: "openai",
				Account:  "default",
				Model:    "gpt-4o-mini",
			},
			"big_brain": {
				Provider: "openai",
				Account:  "work",
				Model:    "gpt-4o",
			},
		},
		DefaultProfile: "small_brain",
	}

	// Save it
	if err := original.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	path, _ := ConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load it back
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify round-trip
	if loaded.DefaultProfile != original.DefaultProfile {
		t.Errorf("DefaultProfile = %q, want %q", loaded.DefaultProfile, original.DefaultProfile)
	}

	if len(loaded.Providers) != len(original.Providers) {
		t.Errorf("Providers count = %d, want %d", len(loaded.Providers), len(original.Providers))
	}

	if len(loaded.Profiles) != len(original.Profiles) {
		t.Errorf("Profiles count = %d, want %d", len(loaded.Profiles), len(original.Profiles))
	}

	// Check specific profile
	profile, ok := loaded.Profiles["small_brain"]
	if !ok {
		t.Fatal("small_brain profile not found")
	}
	if profile.Model != "gpt-4o-mini" {
		t.Errorf("small_brain.Model = %q, want %q", profile.Model, "gpt-4o-mini")
	}
}

func TestConfigDir_CreatesDirectory(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}

	expected := filepath.Join(tmp, ".config", "sage")
	if dir != expected {
		t.Errorf("ConfigDir() = %q, want %q", dir, expected)
	}

	// Verify directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("path is not a directory")
	}
}

func TestConfig_GetProfile(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]Profile{
			"test": {
				Provider: "openai",
				Account:  "default",
				Model:    "gpt-4o",
			},
		},
		DefaultProfile: "test",
	}

	// Get by name
	p, err := cfg.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile(test) error = %v", err)
	}
	if p.Model != "gpt-4o" {
		t.Errorf("Model = %q, want %q", p.Model, "gpt-4o")
	}
	if p.Name != "test" {
		t.Errorf("Name = %q, want %q", p.Name, "test")
	}

	// Get default (empty name)
	p, err = cfg.GetProfile("")
	if err != nil {
		t.Fatalf("GetProfile('') error = %v", err)
	}
	if p.Name != "test" {
		t.Errorf("default profile Name = %q, want %q", p.Name, "test")
	}

	// Get non-existent
	_, err = cfg.GetProfile("nonexistent")
	if err == nil {
		t.Error("GetProfile(nonexistent) should return error")
	}
}

func TestConfig_GetProfile_NoDefault(t *testing.T) {
	cfg := &Config{
		Profiles:       map[string]Profile{},
		DefaultProfile: "",
	}

	_, err := cfg.GetProfile("")
	if err == nil {
		t.Error("GetProfile('') with no default should return error")
	}
}

func TestLoadConfig_OldFormat(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Create config directory
	configDir := filepath.Join(tmp, ".config", "sage")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("mkdir error = %v", err)
	}

	// Write old format config (accounts as array, base_url at provider level)
	oldConfig := `{
  "providers": {
    "openai": {
      "accounts": ["default"],
      "base_url": ""
    }
  },
  "profiles": {},
  "default_profile": ""
}`
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte(oldConfig), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	// Loading should error with helpful message
	_, err := LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig() should error on old format")
	}
	if !contains(err.Error(), "v0.3.0") {
		t.Errorf("Error should mention v0.3.0, got: %v", err)
	}
}

func TestIsOldConfigFormat(t *testing.T) {
	tests := []struct {
		name string
		json string
		want bool
	}{
		{
			name: "old format with array accounts",
			json: `{"providers": {"openai": {"accounts": ["default"]}}}`,
			want: true,
		},
		{
			name: "old format with base_url at provider level",
			json: `{"providers": {"openai": {"accounts": [], "base_url": "http://example.com"}}}`,
			want: true,
		},
		{
			name: "new format with nested map accounts",
			json: `{"providers": {"openai": {"accounts": {"default": {"base_url": "http://example.com"}}}}}`,
			want: false,
		},
		{
			name: "new format empty accounts",
			json: `{"providers": {"openai": {"accounts": {}}}}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOldConfigFormat([]byte(tt.json))
			if got != tt.want {
				t.Errorf("isOldConfigFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
