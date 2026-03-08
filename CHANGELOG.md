# Changelog

All notable changes to **aiswitch** are documented here.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)  
Versioning: [Semantic Versioning](https://semver.org/spec/v2.0.0.html)

---

## [0.2.0] — 2026-03-08

### ✨ Added
- **First-run wizard** — running `aiswitch` with no profiles drops straight into `aiswitch add` instead of showing an error

### 🔒 Security
- **Secrets separation** — API keys and tokens moved to `~/.aiswitch/secrets.json` (mode `0600`); `config.json` now contains zero credentials
- **Auto-migration** — legacy `config.json` files with plaintext keys are silently migrated to `secrets.json` on next save
- **Clean provider handoff** — switching away from a profile fully clears that provider's block in `env.sh`, preventing stale keys from persisting

### 🛠 Improvements
- **Profile name validation** — names restricted to letters, digits, hyphens, and underscores
- **Shell detection** — `detectShell()` now falls back to `$ZSH_VERSION` / `$FISH_VERSION` when `$SHELL` is unset
- **PS1 fix** — `env.ps1` is only generated on Windows; non-Windows systems no longer create the file
- **`omitempty` on all sensitive fields** — empty `api_key` fields no longer appear as `""` in `config.json`
- **Shared merge helpers** — provider env-file logic extracted into `internal/providers/merge` package, eliminating duplication across all four providers
- **Consistent secret masking** — three different private `maskSecret()` implementations unified into `internal/ui.MaskSecret()`

### 🧪 Testing
- **Integration test suite** — builds the real binary and exercises all commands against an isolated temp home

---

## [0.1.0] — 2026-03-08 🎉

> First public release of aiswitch — switch between Claude, OpenAI, Gemini,
> and GitHub Copilot accounts in one command. Works with Cursor and Windsurf IDEs.

### ✨ Added

#### Core CLI
- **Interactive TUI** — fuzzy-searchable profile picker built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **`aiswitch use <profile>`** — direct switch in under 100 ms
- **`aiswitch add [name]`** — guided multi-step interactive form; pre-fills when updating an existing profile
- **`aiswitch list`** — tabular profile overview with active indicator
- **`aiswitch remove <profile>`** — profile deletion
- **`aiswitch current`** — show active profile with live system state (masked keys, IDE patch status)
- **`aiswitch --version`** — prints version, commit hash, and build date injected by GoReleaser
- **Profile name validation** — rejects names with invalid characters at creation time

#### AI Providers
| Provider | What gets applied |
|---|---|
| **Claude** (Anthropic) | `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL` → `env.sh`/`env.ps1` · `~/.anthropic/api_key` · `~/.claude/.credentials.json` (Claude Code) |
| **OpenAI** | `OPENAI_API_KEY`, `OPENAI_ORG_ID`, `OPENAI_MODEL` → `env.sh`/`env.ps1` · `~/.config/openai/api_key` |
| **Gemini** (Google AI) | `GEMINI_API_KEY`, `GOOGLE_API_KEY`, `GEMINI_MODEL`, `GOOGLE_CLOUD_PROJECT` → `env.sh`/`env.ps1` · `~/.config/gemini/api_key` |
| **GitHub Copilot** | `GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_USER` → `env.sh`/`env.ps1` · `~/.config/gh/hosts.yml` · `git config --global user.email` |

- All providers write to both `~/.aiswitch/env.sh` (Unix) and `~/.aiswitch/env.ps1` (Windows)
- **Clean provider handoff** — stale env blocks from previous profiles are cleared (`ClearBlock`) so no leftover keys bleed across switches
- Shared `internal/providers/merge` package — deduplicates env-file merge logic across all providers

#### IDE Integration
- **Cursor IDE** — patches `settings.json` with `anthropic.apiKey`, `openai.apiKey`, `googleGenerativeAI.apiKey`; no restart needed (VS Code hot-reloads)
- **Windsurf IDE** — same `settings.json` patching
- `InstalledIDEs()` — auto-detects which IDEs are present and marks them **✓ installed** in the `add` wizard

#### Shell Integration
- **`aiswitch setup`** — one command to write shell integration into the correct profile file; idempotent (guarded markers); `--shell`, `--dry-run`, `--force` flags
- **`aiswitch shell-init [--shell]`** — prints integration code for manual setup
- Supports **zsh** (`add-zsh-hook chpwd`), **bash** (`PROMPT_COMMAND`), **fish** (`--on-variable PWD`), **PowerShell** (`Set-Location` override)
- Shell wrapper sources `env.sh` after every `aiswitch` call so env vars are live in the current session immediately
- Improved `detectShell()` — uses `$SHELL`, `$ZSH_VERSION`, `$FISH_VERSION`, and OS as fallbacks

#### Per-project `.aiswitch` file
- **`aiswitch init`** — interactive form to create a per-project `.aiswitch` file (no secrets — safe to commit)
- **`aiswitch detect [--quiet]`** — walks up the directory tree, finds the nearest `.aiswitch`, applies the profile with optional overrides
- YAML format and plain-text shorthand (just the profile name) both supported
- Per-project model overrides: `claude.model`, `openai.model`, `gemini.model`, `github.email`

#### Security
- **Secrets separated** — API keys and tokens stored in `~/.aiswitch/secrets.json` (mode `0600`), separate from non-sensitive `config.json`
- `omitempty` on all sensitive fields — keys never appear as empty strings in stored JSON
- `internal/ui.MaskSecret()` — consolidated secret masking (replaces three independent copies)

#### Testing
- Unit tests: `internal/config`, `internal/localfile`, `internal/providers/merge`, `internal/ui`
- Integration test suite (`integration_test.go`) — builds and exercises the real binary end-to-end

#### Platform & Distribution
- macOS Apple Silicon (arm64), macOS Intel (amd64)
- Linux amd64, Linux arm64
- Windows amd64
- Zero runtime dependencies — single static binary (~5 MB)
- GoReleaser pipeline — cross-compiled binaries, `.tar.gz` / `.zip` archives, `checksums.txt`
- Dependabot — weekly updates for Go modules and GitHub Actions

#### Developer Experience
- Git hooks (`.githooks/`) — `pre-commit`: gofmt, go vet, go build, golangci-lint; `commit-msg`: Conventional Commits
- `make hooks` — one command to activate hooks after cloning
- GitHub Actions CI — build + test + lint + cross-compile on every push and PR
- `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, GitHub issue + PR templates

---

[0.1.0]: https://github.com/anmolnagpal/aiswitch/releases/tag/v0.1.0
