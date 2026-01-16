# Sage

Unified CLI and Go library for LLM providers.

Sage provides a single interface for working with multiple LLM providers (OpenAI, Anthropic, Ollama), with secure credential storage and user-defined profiles.

## Quick Start

```bash
# Add a provider (you'll be prompted for required fields)
sage provider add openai
API Key: ****
Base URL (optional) [https://api.openai.com/v1]:

# Add Ollama (different fields)
sage provider add ollama
Base URL [http://localhost:11434]:
API Key (optional):

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
- **Dynamic configuration**: Each provider declares its own field requirements
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

## Library Usage

```go
import "github.com/not-emily/sage/pkg/sage"

// Query provider field requirements
fields, _ := sage.GetProviderFields("openai")
for _, f := range fields {
    fmt.Printf("%s (required=%v, secret=%v)\n", f.Label, f.Required, f.Secret)
}

// Create client
client, _ := sage.NewClient()

// Add provider with fields
client.AddProviderAccount("openai", "default", map[string]string{
    "api_key": "sk-...",
})

// Add Ollama (different required fields)
client.AddProviderAccount("ollama", "local", map[string]string{
    "base_url": "http://localhost:11434",
})

// Create profile and make completion
client.AddProfile("myprofile", sage.Profile{
    Provider: "openai",
    Account:  "default",
    Model:    "gpt-4o-mini",
})

resp, _ := client.Complete("myprofile", sage.Request{
    Prompt: "Hello!",
})
fmt.Println(resp.Content)
```

## Documentation

- [Installation](docs/installation.md)
- [CLI Usage](docs/cli-usage.md)
- [Library Usage](docs/library-usage.md)
- [Changelog](CHANGELOG.md)

## License

MIT
