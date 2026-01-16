# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-01-15

### Breaking Changes

- **Config format changed**: Provider accounts now store per-account fields instead of a simple account list. Old config files are incompatible - delete `~/.config/sage/config.json` and re-add providers.
- **Secrets key format changed**: Now uses `provider:account:field` format (was `provider:account`). Old secrets will not be found - delete `~/.config/sage/secrets.enc` and re-add providers.
- **AddProviderAccount signature changed**: Now accepts `fields map[string]string` instead of `apiKey string`.

### Added

- `ProviderField` type for declaring provider configuration requirements
- `Provider.Fields()` method - providers now declare their own field requirements
- `GetProviderFields(provider)` function to query field requirements
- Per-account field storage - different accounts can have different configurations
- Dynamic CLI prompting based on provider field requirements
- Old config format detection with helpful migration message

### Changed

- OpenAI provider: API Key (required), Base URL (optional, default: `https://api.openai.com/v1`)
- Anthropic provider: API Key (required)
- Ollama provider: Base URL (required, default: `http://localhost:11434`), API Key (optional)
- CLI prompts show field requirements: required fields first, defaults in brackets

## [0.2.0] - 2026-01-05

### Added

- `ListModels()` method to Provider interface
- `sage provider models <provider>` CLI command
- Auto-initialize master key on first use (no longer requires `sage init`)

### Fixed

- OpenAI: Use `max_completion_tokens` for newer models (o1, o3, gpt-4o, gpt-5)

## [0.1.0] - 2026-01-04

### Added

- Initial release
- Providers: OpenAI, Anthropic, Ollama
- Encrypted secrets storage (AES-256-GCM)
- Multi-account support per provider
- Profile management
- Streaming completions
- Go library API
