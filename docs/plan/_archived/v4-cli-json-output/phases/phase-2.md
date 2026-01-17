# Phase 2: CLI JSON Coverage

> **Depends on:** Phase 1 (JSON Streaming)
> **Enables:** Phase 4 (Documentation)
>
> See: [Full Plan](../plan.md)

## Goal

Ensure all sage CLI commands support `--json` flag for machine-readable output, enabling hub-core to interact with sage entirely through structured data.

## Key Deliverables

- `--json` flag on all subcommands
- Consistent JSON output schema per command
- Error output as JSON when `--json` flag is set

## Commands to Update

| Command | Current Output | JSON Output |
|---------|----------------|-------------|
| `sage version` | `0.3.1` | `{"version":"0.3.1"}` |
| `sage provider list` | Table format | `{"providers":[...]}` |
| `sage provider add` | Success message | `{"success":true,"provider":"..."}` |
| `sage provider remove` | Success message | `{"success":true}` |
| `sage provider models` | List | `{"models":[...]}` |
| `sage profile list` | Table format | `{"profiles":[...]}` |
| `sage profile add` | Success message | `{"success":true,"profile":"..."}` |
| `sage profile remove` | Success message | `{"success":true}` |
| `sage profile default` | Success/current | `{"default":"..."}` |

## Files to Modify

- `internal/cli/root.go` — Add --json to version
- `internal/cli/provider.go` — Add --json to all provider subcommands
- `internal/cli/profile.go` — Add --json to all profile subcommands

## Dependencies

**Internal:** Phase 1 establishes the --json pattern

**External:** None

## Implementation Notes

**Consistent pattern for each command:**

```go
func runProviderList(args []string) error {
    fs := flag.NewFlagSet("provider list", flag.ExitOnError)
    jsonOutput := fs.Bool("json", false, "output JSON")
    fs.Parse(args)

    // ... get data ...

    if *jsonOutput {
        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        return enc.Encode(map[string]interface{}{
            "providers": providers,
        })
    }

    // ... existing table output ...
}
```

**Error handling with --json:**

When `--json` is set, errors should also be JSON:

```go
func jsonError(err error) {
    json.NewEncoder(os.Stderr).Encode(map[string]interface{}{
        "error": err.Error(),
    })
}
```

Consider a helper function or wrapper to handle this consistently.

**Provider list JSON schema:**

```json
{
  "providers": [
    {
      "name": "openai",
      "display_name": "OpenAI",
      "accounts": ["default", "work"]
    }
  ]
}
```

**Profile list JSON schema:**

```json
{
  "profiles": [
    {
      "name": "fast",
      "provider": "openai",
      "account": "default",
      "model": "gpt-4o-mini",
      "is_default": true
    }
  ],
  "default": "fast"
}
```

**Models list JSON schema:**

```json
{
  "provider": "openai",
  "models": [
    {
      "id": "gpt-4o",
      "name": "GPT-4o",
      "description": "Most capable model"
    }
  ]
}
```

## Validation

- [ ] `sage version --json` outputs valid JSON with version
- [ ] `sage provider list --json` outputs provider array
- [ ] `sage provider models openai --json` outputs model array
- [ ] `sage profile list --json` outputs profile array with default
- [ ] `sage profile add ... --json` outputs success confirmation
- [ ] Errors with `--json` flag output JSON to stderr
- [ ] All commands without `--json` behave unchanged
