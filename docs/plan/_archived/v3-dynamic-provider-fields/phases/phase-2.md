# Phase 2: Storage Format

> **Depends on:** Phase 1 (Provider Fields)
> **Enables:** Phase 3 (Client API)
>
> See: [Full Plan](../plan.md)

## Goal

Update config and secrets storage to support per-account dynamic fields instead of a single api_key per account.

## Key Deliverables

- Updated ProviderConfig struct with per-account field storage
- New secrets key format: `provider:account:field`
- Old config format detection with helpful error message
- Updated internal methods that read/write provider config

## Files to Modify

- `pkg/sage/config.go` — Update ProviderConfig struct, add format detection
- `pkg/sage/secrets.go` — Update key format helpers

## Dependencies

**Internal:** ProviderField type from Phase 1

**External:** None

## Implementation Notes

### New Config Structure

**Before:**
```go
type ProviderConfig struct {
    Accounts map[string]string `json:"accounts"` // account -> api_key reference
}
```

**After:**
```go
type ProviderConfig struct {
    Accounts map[string]map[string]string `json:"accounts"` // account -> field -> value
}
```

### Config File Format Change

**Before (config.json):**
```json
{
  "providers": {
    "openai": {
      "accounts": {
        "default": "openai:default"
      }
    }
  }
}
```

**After (config.json):**
```json
{
  "providers": {
    "openai": {
      "accounts": {
        "default": {
          "base_url": "https://api.openai.com/v1"
        }
      }
    }
  }
}
```

Note: Secret fields are NOT stored in config.json - only non-secret fields.

### Secrets Key Format

**Before:** `provider:account` (e.g., `openai:default`)

**After:** `provider:account:field` (e.g., `openai:default:api_key`)

### Old Format Detection

When loading config, detect old format (where account value is a string instead of a map) and return a clear error:

```go
func (c *Config) validateProviderFormat() error {
    // Check if any provider account has the old string format
    // Return: "Config format has changed in v0.3.0. Please remove ~/.sage/config.json
    //          and re-add your providers with 'sage provider add <provider>'"
}
```

### Helper Functions

Add/update these internal helpers:

```go
// Get secret key for a specific field
func secretKey(provider, account, field string) string {
    return fmt.Sprintf("%s:%s:%s", provider, account, field)
}

// Get all fields for an account (merges config + secrets)
func (c *Config) getAccountFields(provider, account string) (map[string]string, error)

// Set fields for an account (splits to config + secrets)
func (c *Config) setAccountFields(provider, account string, fields map[string]string, fieldDefs []ProviderField) error
```

## Validation

How do we know this phase is complete?

- [ ] ProviderConfig uses new nested map structure
- [ ] Secrets use new `provider:account:field` key format
- [ ] Loading old config format returns helpful error message
- [ ] Internal field getter/setter helpers work correctly
- [ ] Unit tests for format detection
- [ ] Unit tests for field storage/retrieval
