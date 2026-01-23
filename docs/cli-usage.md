# CLI Usage

## Overview

```
sage <command> [flags]

Commands:
  init        Initialize sage (create config, generate master key)
  complete    Send a completion request
  provider    Manage provider accounts
  profile     Manage profiles
  version     Show version
  help        Show help
```

## Global Flags

All commands support `--json` for machine-readable output. When `--json` is used, errors are also returned as JSON to stdout.

## Init Command

Initialize sage configuration and encryption key.

```bash
sage init
```

Creates:
- `~/.config/sage/config.json`
- `~/.config/sage/master.key` (chmod 600)

Run this once before using other commands.

## Complete Command

Send a completion request to an LLM.

```bash
sage complete [flags] <prompt>
```

### Flags

| Flag | Description |
|------|-------------|
| `--profile` | Profile to use (default: configured default) |
| `--system` | System message |
| `--max-tokens` | Maximum tokens to generate |
| `--json` | Output full response as JSON instead of streaming |
| `--stream` | Stream output (use with `--json` for NDJSON) |

### Examples

```bash
# Basic completion (streams response)
sage complete "Explain recursion in one sentence"

# Use specific profile
sage complete --profile=claude "Write a haiku about Go"

# JSON output (for scripting)
sage complete --json "What is 2+2?"

# NDJSON streaming (for real-time machine parsing)
sage complete --json --stream "Hello"

# Read prompt from stdin
echo "Translate to French: Hello" | sage complete

# Multi-line prompt from stdin
cat << 'EOF' | sage complete
Summarize this code:

func main() {
    fmt.Println("Hello")
}
EOF
```

### Output Modes

**Streaming (default)**: Text streams to stdout as it's generated.

**JSON mode** (`--json`): Buffers and returns full response:
```json
{
  "content": "The answer is 4.",
  "model": "gpt-4o-mini",
  "usage": {
    "prompt_tokens": 12,
    "completion_tokens": 5
  }
}
```

**NDJSON streaming** (`--json --stream`): Real-time JSON streaming, one object per line:
```json
{"content":"Hello","done":false}
{"content":"!","done":false}
{"content":" How","done":false}
{"content":" can","done":false}
{"content":" I","done":false}
{"content":" help","done":false}
{"content":"?","done":false}
{"content":"","done":true,"model":"gpt-4o"}
```

Each line is valid JSON. The final chunk has `"done":true` and includes the model name.

## Provider Commands

Manage provider accounts and API keys.

```bash
sage provider <command> [flags]

Commands:
  list      List configured providers and accounts
  add       Add a provider account
  remove    Remove a provider account
  models    List available models from a provider
  fields    List configuration fields for a provider
```

### Supported Providers

- `openai` — OpenAI API
- `anthropic` — Anthropic Claude API
- `ollama` — Local Ollama instance

### provider list

```bash
sage provider list
```

Output:
```
openai:
  - default
  - work
anthropic:
  - default
```

**JSON output** (`--json`):
```json
{
  "providers": [
    {"name": "openai", "accounts": ["default", "work"]},
    {"name": "anthropic", "accounts": ["default"]}
  ]
}
```

### provider add

```bash
sage provider add <provider> [flags]
```

| Flag | Description |
|------|-------------|
| `--account` | Account name (default: "default") |
| `--api-key-env` | Environment variable containing API key |
| `--base-url` | Custom base URL (for proxies or compatible APIs) |
| `--stdin` | Read fields as JSON from stdin (for scripting) |

Examples:

```bash
# Interactive prompt for API key
sage provider add openai

# Multiple accounts
sage provider add openai --account=work

# From environment variable (for CI/CD)
export OPENAI_API_KEY="sk-..."
sage provider add openai --api-key-env=OPENAI_API_KEY

# Custom endpoint (OpenAI-compatible)
sage provider add openai --base-url=https://api.myproxy.com/v1

# Ollama (typically no API key needed)
sage provider add ollama

# Remote Ollama
sage provider add ollama --base-url=http://server:11434
```

#### Scripting with --stdin

For programmatic use, pass fields as JSON via stdin. This keeps secrets out of bash history and process arguments.

```bash
# Discover required fields first
sage provider fields openai --json

# Add provider with fields from stdin
echo '{"api_key":"sk-..."}' | sage provider add openai --stdin --json

# With custom base URL
echo '{"api_key":"sk-...","base_url":"https://api.myproxy.com/v1"}' | \
  sage provider add openai --stdin --json

# Flags override stdin values
echo '{"api_key":"sk-..."}' | \
  sage provider add openai --stdin --base-url=https://custom.url --json
```

**JSON output** (`--json`):
```json
{
  "success": true,
  "provider": "openai",
  "account": "default"
}
```

### provider remove

```bash
sage provider remove <provider> [flags]
```

| Flag | Description |
|------|-------------|
| `--account` | Account name to remove (default: "default") |

Examples:

```bash
sage provider remove openai --account=work
```

**JSON output** (`--json`):
```json
{
  "success": true,
  "provider": "openai",
  "account": "work"
}
```

### provider models

List available models from a provider.

```bash
sage provider models <provider> [--account=NAME]
```

Examples:

```bash
sage provider models openai
sage provider models ollama
sage provider models openai --account=work
```

**JSON output** (`--json`):
```json
{
  "provider": "openai",
  "models": [
    {"id": "gpt-4o", "name": "gpt-4o"},
    {"id": "gpt-4o-mini", "name": "gpt-4o-mini"}
  ]
}
```

### provider fields

List configuration fields required by a provider. Useful for scripting `provider add`.

```bash
sage provider fields <provider>
```

Examples:

```bash
sage provider fields openai
sage provider fields anthropic
sage provider fields ollama
```

Output:
```
Fields for openai:

  api_key: API Key (required)
  base_url: Base URL [default: https://api.openai.com/v1]
```

**JSON output** (`--json`):
```json
{
  "provider": "openai",
  "fields": [
    {"key": "api_key", "label": "API Key", "required": true, "secret": true},
    {"key": "base_url", "label": "Base URL", "required": false, "secret": false, "default": "https://api.openai.com/v1"}
  ]
}
```

## Profile Commands

Manage profiles that bind provider accounts to models.

```bash
sage profile <command> [flags]

Commands:
  list        List configured profiles
  add         Add a profile
  remove      Remove a profile
  set-default Set the default profile
```

### profile list

```bash
sage profile list
```

Output:
```
default
  provider: openai
  account:  default
  model:    gpt-4o-mini
fast (default)
  provider: openai
  account:  default
  model:    gpt-4o-mini
smart
  provider: anthropic
  account:  default
  model:    claude-sonnet-4-20250514
```

**JSON output** (`--json`):
```json
{
  "default": "fast",
  "profiles": [
    {"name": "default", "provider": "openai", "account": "default", "model": "gpt-4o-mini", "is_default": false},
    {"name": "fast", "provider": "openai", "account": "default", "model": "gpt-4o-mini", "is_default": true},
    {"name": "smart", "provider": "anthropic", "account": "default", "model": "claude-sonnet-4-20250514", "is_default": false}
  ]
}
```

### profile add

```bash
sage profile add <name> --provider=X --model=Y [--account=Z]
```

| Flag | Description |
|------|-------------|
| `--provider` | Provider name (required) |
| `--model` | Model name (required) |
| `--account` | Provider account (default: "default") |

Examples:

```bash
# Basic profile
sage profile add default --provider=openai --model=gpt-4o-mini

# Using specific account
sage profile add work --provider=openai --model=gpt-4o --account=work

# Anthropic profile
sage profile add claude --provider=anthropic --model=claude-sonnet-4-20250514

# Local Ollama
sage profile add local --provider=ollama --model=llama3.2
```

**JSON output** (`--json`):
```json
{
  "success": true,
  "profile": "default"
}
```

### profile remove

```bash
sage profile remove <name>
```

Note: Cannot remove the default profile. Set a different default first.

**JSON output** (`--json`):
```json
{
  "success": true,
  "profile": "myprofile"
}
```

### profile set-default

```bash
sage profile set-default <name>
```

Sets which profile is used when `--profile` is not specified.

**JSON output** (`--json`):
```json
{
  "default": "fast"
}
```

## Version Command

```bash
sage version [--json]
```

**JSON output** (`--json`):
```json
{
  "version": "v0.4.0"
}
```

## Error Handling

When `--json` is used and an error occurs, the error is returned as JSON to stdout:

```json
{
  "error": "profile not found: nonexistent"
}
```

The exit code will still be non-zero (1) for errors.

## Environment Variables

For CI/CD or scripting, you can pass API keys via environment variables:

```bash
export OPENAI_API_KEY="sk-..."
sage provider add openai --api-key-env=OPENAI_API_KEY
```

## Configuration Files

All configuration is stored in `~/.config/sage/`:

| File | Purpose |
|------|---------|
| `config.json` | Providers, profiles, default profile |
| `master.key` | Encryption key (chmod 600) |
| `secrets.enc` | Encrypted API keys |

### config.json structure

```json
{
  "providers": {
    "openai": {
      "accounts": ["default", "work"],
      "base_url": ""
    }
  },
  "profiles": {
    "default": {
      "provider": "openai",
      "account": "default",
      "model": "gpt-4o-mini"
    }
  },
  "default_profile": "default"
}
```

API keys are stored separately in `secrets.enc`, encrypted with the master key.
