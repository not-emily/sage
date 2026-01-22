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

### Download binary (recommended)

```bash
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
```

### From source (requires Go 1.21+)

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

### Basic Example

```go
import (
    "context"
    "fmt"

    "github.com/not-emily/sage/pkg/sage"
)

func main() {
    // Create client
    client, _ := sage.NewClient()

    // Add provider account
    client.AddProviderAccount("openai", "default", map[string]string{
        "api_key": "sk-...",
    })

    // Create profile
    client.AddProfile("myprofile", sage.Profile{
        Provider: "openai",
        Account:  "default",
        Model:    "gpt-4o-mini",
    })

    // Make completion with context
    ctx := context.Background()
    resp, err := client.Complete(ctx, "myprofile", sage.Request{
        Prompt: "Hello!",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(resp.Content)
}
```

### Timeout Example

```go
import (
    "context"
    "time"
)

// Set 30-second timeout for completion
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.Complete(ctx, "myprofile", sage.Request{
    Prompt: "Generate a long essay...",
})
if err != nil {
    // Handle timeout or other errors
    panic(err)
}
```

### Cancellation Example

```go
import (
    "context"
    "time"
)

ctx, cancel := context.WithCancel(context.Background())

// Cancel after 10 seconds
go func() {
    time.Sleep(10 * time.Second)
    cancel()
}()

resp, err := client.Complete(ctx, "myprofile", sage.Request{
    Prompt: "This might take a while...",
})
if err == context.Canceled {
    fmt.Println("Request was canceled")
}
```

### Custom HTTP Client

```go
import (
    "net/http"
    "time"
)

// Configure custom HTTP client with specific timeouts
client, _ := sage.NewClient()
client.HTTPClient = &http.Client{
    Timeout: 2 * time.Minute,
    Transport: &http.Transport{
        MaxIdleConns:        50,
        IdleConnTimeout:     60 * time.Second,
        TLSHandshakeTimeout: 5 * time.Second,
    },
}

ctx := context.Background()
resp, err := client.Complete(ctx, "myprofile", sage.Request{
    Prompt: "Hello!",
})
```

### Streaming Example

```go
import (
    "context"
    "fmt"
)

ctx := context.Background()
chunks, err := client.CompleteStream(ctx, "myprofile", sage.Request{
    Prompt: "Write a story...",
})
if err != nil {
    panic(err)
}

for chunk := range chunks {
    if chunk.Error != nil {
        panic(chunk.Error)
    }
    if chunk.Done {
        break
    }
    fmt.Print(chunk.Content)
}
fmt.Println()
```

### Provider Field Requirements

```go
// Query provider field requirements
fields, _ := sage.GetProviderFields("openai")
for _, f := range fields {
    fmt.Printf("%s (required=%v, secret=%v, default=%s)\n",
        f.Label, f.Required, f.Secret, f.Default)
}
// Output:
// API Key (required=true, secret=true, default=)
// Base URL (required=false, secret=false, default=https://api.openai.com/v1)
```

## Documentation

- [Installation](docs/installation.md)
- [CLI Usage](docs/cli-usage.md)
- [Library Usage](docs/library-usage.md)
- [Changelog](CHANGELOG.md)

## License

MIT
