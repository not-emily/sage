# Sage

Unified CLI and Go library for LLM providers.

Sage provides a single interface for working with multiple LLM providers (OpenAI, Anthropic, Ollama), with secure credential storage and user-defined profiles.

## Quick Start

```bash
# Initialize (creates config, generates encryption key)
sage init

# Add a provider
sage provider add openai
Enter API key: ****

# Create a profile
sage profile add default --provider=openai --model=gpt-4o-mini

# Set as default
sage profile set-default default

# Use it
sage complete "Hello, world!"
```

## Features

- **Multiple providers**: OpenAI, Anthropic, Ollama
- **Secure credentials**: API keys encrypted at rest (AES-256-GCM)
- **Profiles**: Name your configurations (fast, smart, local, etc.)
- **Streaming**: Real-time response output
- **Library**: Import in your Go projects

## Installation

### From source

```bash
go install github.com/not-emily/sage/cmd/sage@latest
```

### Build locally

```bash
git clone https://github.com/not-emily/sage.git
cd sage
./scripts/build.sh
./bin/sage version
```

## Documentation

- [Installation](docs/installation.md)
- [CLI Usage](docs/cli-usage.md)
- [Library Usage](docs/library-usage.md)

## License

MIT
