package sage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInitSecrets_CreatesKeyFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := InitSecrets(); err != nil {
		t.Fatalf("InitSecrets() error = %v", err)
	}

	keyPath, _ := MasterKeyPath()

	// Verify key exists
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("master key not created: %v", err)
	}

	// Verify size
	if info.Size() != keySize {
		t.Errorf("key size = %d, want %d", info.Size(), keySize)
	}

	// Verify permissions
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("key permissions = %o, want 0600", mode)
	}
}

func TestInitSecrets_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// First init
	if err := InitSecrets(); err != nil {
		t.Fatalf("first InitSecrets() error = %v", err)
	}

	keyPath, _ := MasterKeyPath()
	originalKey, _ := os.ReadFile(keyPath)

	// Second init should not change the key
	if err := InitSecrets(); err != nil {
		t.Fatalf("second InitSecrets() error = %v", err)
	}

	newKey, _ := os.ReadFile(keyPath)
	if !bytes.Equal(originalKey, newKey) {
		t.Error("InitSecrets() changed existing key")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := make([]byte, keySize)
	for i := range key {
		key[i] = byte(i)
	}

	original := []byte("hello world, this is a test secret!")

	ciphertext, err := encrypt(key, original)
	if err != nil {
		t.Fatalf("encrypt() error = %v", err)
	}

	// Ciphertext should be larger than original (nonce + tag overhead)
	if len(ciphertext) <= len(original) {
		t.Error("ciphertext should be larger than plaintext")
	}

	plaintext, err := decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("decrypt() error = %v", err)
	}

	if !bytes.Equal(plaintext, original) {
		t.Errorf("decrypt() = %q, want %q", plaintext, original)
	}
}

func TestEncrypt_DifferentNonce(t *testing.T) {
	key := make([]byte, keySize)
	original := []byte("same content")

	ct1, _ := encrypt(key, original)
	ct2, _ := encrypt(key, original)

	// Even with same content, ciphertext should differ (different nonce)
	if bytes.Equal(ct1, ct2) {
		t.Error("encrypt() should produce different ciphertext each time")
	}

	// But both should decrypt to the same value
	pt1, _ := decrypt(key, ct1)
	pt2, _ := decrypt(key, ct2)
	if !bytes.Equal(pt1, pt2) {
		t.Error("both ciphertexts should decrypt to same value")
	}
}

func TestLoadSecrets_NoFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Init to create master key
	if err := InitSecrets(); err != nil {
		t.Fatalf("InitSecrets() error = %v", err)
	}

	secrets, err := LoadSecrets()
	if err != nil {
		t.Fatalf("LoadSecrets() error = %v", err)
	}

	if secrets == nil {
		t.Error("LoadSecrets() should return empty map, not nil")
	}
	if len(secrets) != 0 {
		t.Errorf("LoadSecrets() should return empty map, got %d items", len(secrets))
	}
}

func TestSecrets_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := InitSecrets(); err != nil {
		t.Fatalf("InitSecrets() error = %v", err)
	}

	original := map[string]string{
		"openai:default":    "sk-test-key-12345",
		"anthropic:default": "sk-ant-test-key",
	}

	if err := SaveSecrets(original); err != nil {
		t.Fatalf("SaveSecrets() error = %v", err)
	}

	loaded, err := LoadSecrets()
	if err != nil {
		t.Fatalf("LoadSecrets() error = %v", err)
	}

	if len(loaded) != len(original) {
		t.Errorf("loaded secrets count = %d, want %d", len(loaded), len(original))
	}

	for k, v := range original {
		if loaded[k] != v {
			t.Errorf("loaded[%q] = %q, want %q", k, loaded[k], v)
		}
	}
}

func TestSetGetDeleteSecret(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := InitSecrets(); err != nil {
		t.Fatalf("InitSecrets() error = %v", err)
	}

	// Set
	if err := SetSecret("openai", "work", "sk-work-key"); err != nil {
		t.Fatalf("SetSecret() error = %v", err)
	}

	// Get
	secret, err := GetSecret("openai", "work")
	if err != nil {
		t.Fatalf("GetSecret() error = %v", err)
	}
	if secret != "sk-work-key" {
		t.Errorf("GetSecret() = %q, want %q", secret, "sk-work-key")
	}

	// HasSecret
	has, err := HasSecret("openai", "work")
	if err != nil {
		t.Fatalf("HasSecret() error = %v", err)
	}
	if !has {
		t.Error("HasSecret() should return true")
	}

	// Delete
	if err := DeleteSecret("openai", "work"); err != nil {
		t.Fatalf("DeleteSecret() error = %v", err)
	}

	// Verify deleted
	_, err = GetSecret("openai", "work")
	if err == nil {
		t.Error("GetSecret() should error after delete")
	}
}

func TestLoadSecrets_NoMasterKey(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Don't call InitSecrets - no master key

	_, err := LoadSecrets()
	if err == nil {
		t.Error("LoadSecrets() should error when master key doesn't exist")
	}
}

func TestLoadSecrets_InsecurePermissions(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Create config dir
	dir, _ := ConfigDir()
	keyPath := filepath.Join(dir, "master.key")

	// Write key with wrong permissions
	key := make([]byte, keySize)
	os.WriteFile(keyPath, key, 0644) // World-readable - bad!

	_, err := LoadSecrets()
	if err == nil {
		t.Error("LoadSecrets() should error with insecure permissions")
	}
}
