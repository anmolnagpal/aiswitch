# Changelog

All notable changes to aiswitch are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] — 2026-03-08 🎉 First release

### ✨ Added

#### Core
- **Interactive TUI** — fuzzy-searchable profile picker built with Bubble Tea and Lip Gloss
- **Direct switching** — `aiswitch use <profile>` switches in under 100 ms
- **Global config** — profiles stored in `~/.aiswitch/config.json` (mode `0600`)
- **`aiswitch add`** — guided interactive form to create / update profiles
- **`aiswitch list`** — tabular profile overview
- **`aiswitch remove`** — profile deletion
- **`aiswitch current`** — show active profile and live system state (masked keys)
- **Version info** — `aiswitch --version` shows version, commit, and build date

#### Providers
- **Claude (Anthropic)** — sets `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL` · writes `~/.anthropic/api_key` · patches Claude Code credentials (`~/.claude/.credentials.json`)
- **OpenAI** — sets `OPENAI_API_KEY`, `OPENAI_ORG_ID`, `OPENAI_MODEL` · writes `~/.config/openai/api_key`
- **Gemini (Google AI)** — sets `GEMINI_API_KEY`, `GOOGLE_API_KEY`, `GEMINI_MODEL`, `GOOGLE_CLOUD_PROJECT` · writes `~/.config/gemini/api_key`
- **GitHub Copilot** — sets `GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_USER` · updates `~/.config/gh/hosts.yml` · optionally sets `git config --global user.email`
- All providers write to both `~/.aiswitch/env.sh` (Unix) and `~/.aiswitch/env.ps1` (Windows)

#### IDE integration
- **Cursor IDE** — patches `settings.json` with `anthropic.apiKey`, `openai.apiKey`, `googleGenerativeAI.apiKey` — no restart needed
- **Windsurf IDE** — same `settings.json` patching
- Auto-detects installed IDEs during `aiswitch add`

#### Shell integration
- **`aiswitch setup`** — one command to install shell integration; auto-detects shell; idempotent (guarded markers); `--dry-run`, `--force`, `--shell` flags
- **`aiswitch shell-init`** — prints shell integration code for manual setup
- Supports **zsh** (`add-zsh-hook chpwd`), **bash** (`PROMPT_COMMAND`), **fish** (`--on-variable PWD`), **PowerShell** (`Set-Location` override)
- Wrapper function sources `env.sh` after every `aiswitch` call so env vars are live in the current session

#### Per-project `.aiswitch` file
- **`aiswitch init`** — interactive form to create a project-level `.aiswitch` file (no secrets, safe to commit)
- **`aiswitch detect`** — walks up the directory tree, applies the nearest `.aiswitch` profile with optional per-provider model overrides
- Supports YAML format and plain-text shorthand (just the profile name)
- Per-project model overrides for Claude, OpenAI, Gemini, and email override for GitHub

#### Platform support
- macOS (Apple Silicon arm64 + Intel amd64)
- Linux (amd64 + arm64)
- Windows (amd64)
- Zero runtime dependencies — single static binary (~5 MB)

#### Developer experience
- Git hooks (`.githooks/`) — `pre-commit`: gofmt, go vet, go build, golangci-lint; `commit-msg`: Conventional Commits
- `make hooks` — one command to activate hooks for new contributors
- GitHub Actions — CI (build, test, lint, cross-compile) + GoReleaser release pipeline
- Dependabot — weekly updates for Go modules and GitHub Actions

---

[0.1.0]: https://github.com/anmolnagpal/aiswitch/releases/tag/v0.1.0
