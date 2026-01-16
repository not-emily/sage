# Phase 4: CLI Update

> **Depends on:** Phase 3 (Client API)
> **Enables:** Phase 5 (Documentation)
>
> See: [Full Plan](../plan.md)

## Goal

Update the Sage CLI to dynamically prompt for provider fields based on field metadata.

## Key Deliverables

- Dynamic field prompting in `sage provider add`
- Proper handling of secret vs non-secret fields
- Default value display and application
- Field ordering (required first)

## Files to Modify

- `internal/cli/provider.go` — Update add command to use dynamic fields

## Dependencies

**Internal:**
- GetProviderFields from Phase 3
- AddProviderAccount (new signature) from Phase 3

**External:** None

## Implementation Notes

### Current Implementation

```go
func addProvider(args []string) error {
    // Currently hardcoded to prompt for "API Key" for all providers
    fmt.Print("API Key: ")
    apiKey := readSecretInput()
    return client.AddProviderAccount(provider, account, apiKey)
}
```

### New Implementation

```go
func addProvider(args []string) error {
    provider := args[0]
    account := "default"
    if len(args) > 1 {
        account = args[1]
    }

    // Get field requirements
    fields, err := sage.GetProviderFields(provider)
    if err != nil {
        return err
    }

    // Collect field values
    values := make(map[string]string)

    // Sort: required fields first
    sort.Slice(fields, func(i, j int) bool {
        if fields[i].Required != fields[j].Required {
            return fields[i].Required
        }
        return i < j // preserve original order within required/optional
    })

    for _, f := range fields {
        value := promptForField(f)
        if value != "" {
            values[f.Key] = value
        }
    }

    return client.AddProviderAccount(provider, account, values)
}

func promptForField(f sage.ProviderField) string {
    // Build prompt
    prompt := f.Label
    if !f.Required {
        prompt += " (optional)"
    }
    if f.Default != "" {
        prompt += fmt.Sprintf(" [%s]", f.Default)
    }
    prompt += ": "

    fmt.Print(prompt)

    var value string
    if f.Secret {
        value = readSecretInput() // no echo
    } else {
        value = readInput() // normal input
    }

    // Apply default if empty and default exists
    if value == "" && f.Default != "" {
        return f.Default
    }

    return value
}
```

### Example User Experience

**Adding OpenAI:**
```
$ sage provider add openai
API Key: ********
Base URL (optional) [https://api.openai.com/v1]:
Provider openai added with account 'default'
```

**Adding Ollama:**
```
$ sage provider add ollama
Base URL [http://localhost:11434]:
API Key (optional):
Provider ollama added with account 'default'
```

**Adding Ollama with custom URL:**
```
$ sage provider add ollama remote
Base URL [http://localhost:11434]: http://192.168.1.100:11434
API Key (optional): ********
Provider ollama added with account 'remote'
```

### Secret Input Handling

The existing `readSecretInput()` function should be used for secret fields to prevent echo. If it doesn't exist, implement it:

```go
func readSecretInput() string {
    // Disable echo, read input, restore echo
    // Use golang.org/x/term if available, otherwise basic approach
}
```

Note: Since Sage is stdlib-only, we need a stdlib approach for hiding input. On Unix, this can be done with terminal manipulation. For simplicity, we may just accept that secrets are visible during input (many CLIs do this).

## Validation

How do we know this phase is complete?

- [ ] `sage provider add openai` prompts for API Key (required), Base URL (optional)
- [ ] `sage provider add ollama` prompts for Base URL (required), API Key (optional)
- [ ] Default values displayed in prompt and applied when user presses Enter
- [ ] Required fields prompted first
- [ ] Secret fields use appropriate input handling
- [ ] Manual test: add provider → list models → verify connection works
