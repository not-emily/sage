# Installation

## Prerequisites

- Go 1.21+ (for building from source or library usage)
- No external dependencies (stdlib only)

## Install from Source

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
