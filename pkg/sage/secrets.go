package sage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	keySize   = 32 // AES-256
	nonceSize = 12 // GCM standard nonce size
)

// MasterKeyPath returns the path to master.key.
func MasterKeyPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "master.key"), nil
}

// SecretsPath returns the path to secrets.enc.
func SecretsPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "secrets.enc"), nil
}

// InitSecrets creates master.key if it doesn't exist.
// The key file is created with mode 0600 (owner read/write only).
func InitSecrets() error {
	keyPath, err := MasterKeyPath()
	if err != nil {
		return err
	}

	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		return nil // Already exists
	}

	// Generate random key
	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return fmt.Errorf("cannot generate random key: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return fmt.Errorf("cannot write master key: %w", err)
	}

	return nil
}

// loadMasterKey reads and validates the master key.
func loadMasterKey() ([]byte, error) {
	keyPath, err := MasterKeyPath()
	if err != nil {
		return nil, err
	}

	// Check key exists, auto-initialize if not
	info, err := os.Stat(keyPath)
	if os.IsNotExist(err) {
		// Auto-initialize on first use
		if err := InitSecrets(); err != nil {
			return nil, fmt.Errorf("failed to auto-initialize: %w", err)
		}
		// Re-stat after initialization
		info, err = os.Stat(keyPath)
		if err != nil {
			return nil, fmt.Errorf("cannot stat master key after init: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("cannot stat master key: %w", err)
	}

	// Check permissions (should be 0600)
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		return nil, fmt.Errorf("master key has insecure permissions %o (should be 600)", mode)
	}

	// Read key
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read master key: %w", err)
	}

	if len(key) != keySize {
		return nil, fmt.Errorf("invalid master key size: got %d, want %d", len(key), keySize)
	}

	return key, nil
}

// encrypt encrypts plaintext using AES-256-GCM.
// Returns: nonce (12 bytes) || ciphertext
func encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return append(nonce, ciphertext...), nil
}

// decrypt decrypts data encrypted with encrypt().
// Expects: nonce (12 bytes) || ciphertext
func decrypt(key, data []byte) ([]byte, error) {
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// LoadSecrets decrypts and returns the secrets map.
// Returns empty map if secrets file doesn't exist.
func LoadSecrets() (map[string]string, error) {
	key, err := loadMasterKey()
	if err != nil {
		return nil, err
	}

	secretsPath, err := SecretsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(secretsPath)
	if os.IsNotExist(err) {
		return make(map[string]string), nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read secrets file: %w", err)
	}

	plaintext, err := decrypt(key, data)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt secrets: %w", err)
	}

	var secrets map[string]string
	if err := json.Unmarshal(plaintext, &secrets); err != nil {
		return nil, fmt.Errorf("invalid secrets format: %w", err)
	}

	return secrets, nil
}

// SaveSecrets encrypts and saves the secrets map.
func SaveSecrets(secrets map[string]string) error {
	key, err := loadMasterKey()
	if err != nil {
		return err
	}

	plaintext, err := json.Marshal(secrets)
	if err != nil {
		return fmt.Errorf("cannot marshal secrets: %w", err)
	}

	ciphertext, err := encrypt(key, plaintext)
	if err != nil {
		return fmt.Errorf("cannot encrypt secrets: %w", err)
	}

	secretsPath, err := SecretsPath()
	if err != nil {
		return err
	}

	if err := os.WriteFile(secretsPath, ciphertext, 0600); err != nil {
		return fmt.Errorf("cannot write secrets file: %w", err)
	}

	return nil
}

// secretKey returns the map key for a provider:account:field triplet.
func secretKey(provider, account, field string) string {
	return provider + ":" + account + ":" + field
}

// GetSecretField returns a decrypted secret field value.
func GetSecretField(provider, account, field string) (string, error) {
	secrets, err := LoadSecrets()
	if err != nil {
		return "", err
	}

	key := secretKey(provider, account, field)
	secret, ok := secrets[key]
	if !ok {
		return "", fmt.Errorf("no secret found for %s", key)
	}

	return secret, nil
}

// SetSecretField encrypts and stores a secret field value.
func SetSecretField(provider, account, field, value string) error {
	secrets, err := LoadSecrets()
	if err != nil {
		return err
	}

	secrets[secretKey(provider, account, field)] = value
	return SaveSecrets(secrets)
}

// DeleteSecretField removes a secret field.
func DeleteSecretField(provider, account, field string) error {
	secrets, err := LoadSecrets()
	if err != nil {
		return err
	}

	key := secretKey(provider, account, field)
	if _, ok := secrets[key]; !ok {
		return fmt.Errorf("no secret found for %s", key)
	}

	delete(secrets, key)
	return SaveSecrets(secrets)
}

// DeleteAccountSecrets removes all secrets for an account.
func DeleteAccountSecrets(provider, account string) error {
	secrets, err := LoadSecrets()
	if err != nil {
		return err
	}

	prefix := provider + ":" + account + ":"
	for key := range secrets {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			delete(secrets, key)
		}
	}

	return SaveSecrets(secrets)
}

// HasSecretField checks if a secret field exists.
func HasSecretField(provider, account, field string) (bool, error) {
	secrets, err := LoadSecrets()
	if err != nil {
		return false, err
	}

	_, ok := secrets[secretKey(provider, account, field)]
	return ok, nil
}
