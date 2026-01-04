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
| `--json` | Output full response as JSON instead of streaming |

### Examples

```bash
# Basic completion (streams response)
sage complete "Explain recursion in one sentence"

# Use specific profile
sage complete --profile=claude "Write a haiku about Go"

# JSON output (for scripting)
sage complete --json "What is 2+2?"

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

**JSON mode** (`--json`): Returns full response:
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

## Provider Commands

Manage provider accounts and API keys.

```bash
sage provider <command> [flags]

Commands:
  list      List configured providers and accounts
  add       Add a provider account
  remove    Remove a provider account
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

### provider add

```bash
sage provider add <provider> [flags]
```

| Flag | Description |
|------|-------------|
| `--account` | Account name (default: "default") |
| `--api-key-env` | Environment variable containing API key |
| `--base-url` | Custom base URL (for proxies or compatible APIs) |

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

### profile remove

```bash
sage profile remove <name>
```

Note: Cannot remove the default profile. Set a different default first.

### profile set-default

```bash
sage profile set-default <name>
```

Sets which profile is used when `--profile` is not specified.

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
