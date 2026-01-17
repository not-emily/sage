# Phase 4: Documentation

> **Depends on:** Phase 2 (CLI JSON Coverage)
> **Enables:** Hub-core developers can use sage CLI correctly
>
> See: [Full Plan](../plan.md)

## Goal

Document the CLI JSON output format so hub-core (and other consumers) know exactly what to expect from each command.

## Key Deliverables

- CLI reference documentation with JSON examples
- JSON schema documentation for each command
- Installation instructions for pre-compiled binaries

## Files to Create/Modify

- `docs/cli-reference.md` — New or update existing CLI docs
- `README.md` — Add installation section for binaries

## Dependencies

**Internal:** Phase 2 completes JSON implementation

**External:** None

## Implementation Notes

**CLI Reference structure:**

```markdown
# Sage CLI Reference

## Installation

### Pre-compiled binaries (recommended)

curl -L https://github.com/not-emily/sage/releases/latest/download/sage-linux-amd64 -o /usr/local/bin/sage
chmod +x /usr/local/bin/sage

### From source

go install github.com/not-emily/sage/cmd/sage@latest

## Global Flags

- `--json` — Output in JSON format (available on all commands)

## Commands

### sage complete

Send a completion request to an LLM.

**Usage:**
sage complete [flags] <prompt>

**Flags:**
- `--profile` — Profile to use (default: default profile)
- `--system` — System message
- `--max-tokens` — Maximum tokens to generate
- `--json` — Output JSON instead of streaming text
- `--stream` — Stream output (use with --json for NDJSON)

**Examples:**

# Plain text streaming (default)
sage complete "Hello"

# Buffered JSON
sage complete "Hello" --json

# NDJSON streaming
sage complete "Hello" --json --stream

**JSON Output (buffered):**
{
  "content": "Hello! How can I help you?",
  "model": "gpt-4o",
  "usage": {
    "prompt_tokens": 5,
    "completion_tokens": 8
  }
}

**NDJSON Output (streaming):**
{"content":"Hello","done":false}
{"content":"!","done":false}
{"content":"","done":true,"model":"gpt-4o","usage":{"prompt_tokens":5,"completion_tokens":8}}
```

**Document each command:**
- sage version
- sage complete
- sage provider list/add/remove/models
- sage profile list/add/remove/default

**For each command include:**
1. Description
2. Usage syntax
3. Flags
4. Plain output example
5. JSON output example with schema

**README installation section:**

```markdown
## Installation

### Download binary (recommended)

# Linux (amd64)
curl -L https://github.com/not-emily/sage/releases/latest/download/sage-linux-amd64 \
  -o /usr/local/bin/sage && chmod +x /usr/local/bin/sage

# Linux (arm64)
curl -L https://github.com/not-emily/sage/releases/latest/download/sage-linux-arm64 \
  -o /usr/local/bin/sage && chmod +x /usr/local/bin/sage

# macOS (Apple Silicon)
curl -L https://github.com/not-emily/sage/releases/latest/download/sage-darwin-arm64 \
  -o /usr/local/bin/sage && chmod +x /usr/local/bin/sage

# macOS (Intel)
curl -L https://github.com/not-emily/sage/releases/latest/download/sage-darwin-amd64 \
  -o /usr/local/bin/sage && chmod +x /usr/local/bin/sage

### From source (requires Go 1.22+)

go install github.com/not-emily/sage/cmd/sage@latest
```

## Validation

- [ ] CLI reference documents all commands
- [ ] Each command has JSON output example
- [ ] JSON schemas are accurate (match actual output)
- [ ] README has binary installation instructions
- [ ] Installation commands work as documented
- [ ] Examples in docs actually work when run
