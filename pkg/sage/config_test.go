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

	// Create a config
	original := &Config{
		Providers: map[string]ProviderConfig{
			"openai": {
				Accounts: []string{"default", "work"},
				BaseURL:  "",
			},
			"ollama": {
				Accounts: []string{"local"},
				BaseURL:  "http://localhost:11434",
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
