# Installation

## Pre-compiled Binaries (Recommended)

Download the latest binary for your platform:

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

Or install to a user directory (no sudo required):

```bash
mkdir -p ~/.local/bin
curl -L https://github.com/not-emily/sage/releases/latest/download/sage-linux-amd64 \
  -o ~/.local/bin/sage && chmod +x ~/.local/bin/sage
```

Ensure `~/.local/bin` is in your PATH.

## From Source (requires Go 1.21+)

```bash
go install github.com/not-emily/sage/cmd/sage@latest
```

This installs the `sage` binary to `$GOPATH/bin` (usually `~/go/bin`).

Ensure `~/go/bin` is in your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Add this line to your `~/.bashrc` or `~/.zshrc` for persistence.

## Build Locally

```bash
git clone https://github.com/not-emily/sage.git
cd sage
./scripts/build.sh
```

The binary is created at `./bin/sage`. To install system-wide:

```bash
# User-only install
mkdir -p ~/.local/bin
cp ./bin/sage ~/.local/bin/

# Or system-wide (requires sudo)
sudo cp ./bin/sage /usr/local/bin/
```

## Verify Installation

```bash
sage version
```

Expected output:
```
sage v0.1.0
```

## Initialize Sage

After installation, initialize sage to create the config directory and encryption key:

```bash
sage init
```

This creates:
- `~/.config/sage/config.json` — Configuration file
- `~/.config/sage/master.key` — Encryption key (chmod 600)
- `~/.config/sage/secrets.enc` — Encrypted credentials (created when you add a provider)

## Next Steps

1. [Add a provider](cli-usage.md#provider-commands)
2. [Create a profile](cli-usage.md#profile-commands)
3. [Run your first completion](cli-usage.md#complete-command)
