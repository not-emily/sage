# Phase 3: Client API

> **Depends on:** Phase 1 (Provider Fields), Phase 2 (Storage Format)
> **Enables:** Phase 4 (CLI Update)
>
> See: [Full Plan](../plan.md)

## Goal

Update the public Client API to accept dynamic fields and expose provider field metadata.

## Key Deliverables

- Updated `AddProviderAccount` signature to accept fields map
- New `GetProviderFields` function to query field requirements
- Updated provider initialization to use dynamic fields
- Breaking change to public API (v0.3.0)

## Files to Modify

- `pkg/sage/client.go` — Update AddProviderAccount, add GetProviderFields
- `pkg/sage/providers/registry.go` — Ensure provider lookup works for field queries

## Dependencies

**Internal:**
- ProviderField type from Phase 1
- Storage format from Phase 2

**External:** None

## Implementation Notes

### AddProviderAccount Signature Change

**Before:**
```go
func (c *Client) AddProviderAccount(provider, account, apiKey string) error
```

**After:**
```go
func (c *Client) AddProviderAccount(provider, account string, fields map[string]string) error
```

### AddProviderAccount Implementation

```go
func (c *Client) AddProviderAccount(provider, account string, fields map[string]string) error {
    // 1. Get provider to access field definitions
    p, err := GetProvider(provider)
    if err != nil {
        return err
    }

    // 2. Validate required fields are present
    fieldDefs := p.Fields()
    for _, f := range fieldDefs {
        if f.Required {
            if val, ok := fields[f.Key]; !ok || val == "" {
                return fmt.Errorf("required field %q missing", f.Key)
            }
        }
    }

    // 3. Store fields (secrets vs config based on field definition)
    return c.config.setAccountFields(provider, account, fields, fieldDefs)
}
```

### GetProviderFields Function

```go
// GetProviderFields returns the field requirements for a provider.
// This is a package-level function since it doesn't require a configured client.
func GetProviderFields(provider string) ([]ProviderField, error) {
    p, err := GetProvider(provider)
    if err != nil {
        return nil, err
    }
    return p.Fields(), nil
}
```

### Provider Initialization Update

Update how providers are initialized with credentials when making requests:

```go
// Internal: get credentials for a request
func (c *Client) getProviderCredentials(provider, account string) (apiKey, baseURL string, err error) {
    fields, err := c.config.getAccountFields(provider, account)
    if err != nil {
        return "", "", err
    }
    return fields["api_key"], fields["base_url"], nil
}
```

### ListModels Update

The `ListModels` signature currently takes `apiKey, baseURL string`. This can remain unchanged since it's used internally - we just need to extract these from the fields map before calling.

## Validation

How do we know this phase is complete?

- [ ] AddProviderAccount accepts fields map instead of apiKey string
- [ ] GetProviderFields returns field definitions for any provider
- [ ] Required field validation works correctly
- [ ] Secret fields stored in secrets, non-secret in config
- [ ] Provider requests work with new credential retrieval
- [ ] Unit tests for AddProviderAccount with various field combinations
- [ ] Unit tests for GetProviderFields
- [ ] Integration test: add account → make request → verify works
