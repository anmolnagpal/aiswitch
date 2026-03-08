<div align="center">

<h1>⚡ aiswitch</h1>

<p><strong>Switch between Claude, OpenAI, Gemini, and GitHub Copilot accounts in one command.<br/>
Works with Cursor, Windsurf, and every terminal tool — on macOS, Linux, and Windows.</strong></p>

<p>Like <a href="https://github.com/warrensbox/terraform-switcher">tfswitch</a> for Terraform versions, or <a href="https://github.com/nvm-sh/nvm">nvm</a> for Node — but for your AI accounts.</p>

[![CI](https://github.com/anmolnagpal/aiswitch/actions/workflows/ci.yml/badge.svg)](https://github.com/anmolnagpal/aiswitch/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/anmolnagpal/aiswitch)](https://goreportcard.com/report/github.com/anmolnagpal/aiswitch)
[![GitHub Release](https://img.shields.io/github/v/release/anmolnagpal/aiswitch?include_prereleases&sort=semver)](https://github.com/anmolnagpal/aiswitch/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/anmolnagpal/aiswitch)](go.mod)
[![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](#installation)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

</div>

---

```
$ aiswitch

  Switch AI Profile

  ▶ ● work          Claude + OpenAI + GitHub + Cursor/Windsurf   (active)
    ○ personal      Claude + Gemini + GitHub + Cursor
    ○ client-x      OpenAI only
    ○ open-source   GitHub only

  ↑/↓ navigate  •  enter select  •  / filter  •  q quit
```

---

## Table of Contents

- [Why aiswitch?](#why-aiswitch)
- [Features](#features)
- [Installation](#installation)
- [Shell integration](#shell-integration)
- [Quick start](#quick-start)
- [Per-project `.aiswitch` file](#per-project-aiswitch-file)
- [IDE integration](#ide-integration)
- [Commands](#commands)
- [Configuration reference](#configuration-reference)
- [Security](#security)
- [Contributing](#contributing)
- [Roadmap](#roadmap)
- [Related projects](#related-projects)
- [License](#license)

---

## Why aiswitch?

Most developers today juggle **multiple AI accounts** at the same time:

| Account | Why you have more than one |
|---|---|
| Claude (Anthropic) | Work team plan + personal account |
| OpenAI | Different org keys per client project |
| Gemini | AI Studio for experiments, Vertex AI for prod |
| GitHub Copilot | Separate GitHub accounts for work vs. open-source |

Switching between them manually means:

- Editing `~/.bashrc` to swap `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY` …
- Re-entering keys in **Cursor → Settings → Models** or **Windsurf Preferences**
- Running `gh auth switch` and hoping your IDE picks it up
- Burning quota on the wrong account because you forgot to switch

**aiswitch** fixes this the same way `tfswitch` fixed Terraform versions:  
one command to switch, a `.aiswitch` file in each repo to switch automatically on `cd`.

---

## Features

| Category | What it does |
|---|---|
| **Interactive TUI** | Fuzzy-searchable profile list with arrow-key navigation |
| **Direct switch** | `aiswitch use work` — switches in under 100 ms |
| **Per-project pinning** | Commit a `.aiswitch` file; profile switches automatically on `cd` |
| **Auto cd hook** | zsh, bash, fish, PowerShell |
| **Claude** | Sets `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL` · writes `~/.anthropic/api_key` · patches Claude Code credentials |
| **OpenAI** | Sets `OPENAI_API_KEY`, `OPENAI_ORG_ID`, `OPENAI_MODEL` · writes `~/.config/openai/api_key` |
| **Gemini** | Sets `GEMINI_API_KEY`, `GOOGLE_API_KEY`, `GEMINI_MODEL`, `GOOGLE_CLOUD_PROJECT` · writes `~/.config/gemini/api_key` |
| **GitHub Copilot** | Sets `GITHUB_TOKEN`, `GH_TOKEN` · updates `~/.config/gh/hosts.yml` · optionally sets `git config user.email` |
| **Cursor IDE** | Patches `settings.json` — `anthropic.apiKey`, `openai.apiKey`, `googleGenerativeAI.apiKey` |
| **Windsurf IDE** | Same `settings.json` patching (Codeium/Windsurf) |
| **Cross-platform** | macOS (Intel + Apple Silicon), Linux (amd64 + arm64), Windows (amd64) |
| **Zero runtime deps** | Single static binary, ~5 MB |

---

## Installation

### Option 1 — `go install` (recommended)

```bash
go install github.com/anmolnagpal/aiswitch@latest
```

Requires Go 1.22+. The binary is placed in `$(go env GOPATH)/bin` — make sure that's on your `$PATH`.

### Option 2 — Pre-built binary

Download the latest binary for your platform from [Releases](https://github.com/anmolnagpal/aiswitch/releases):

```bash
# macOS Apple Silicon
curl -L https://github.com/anmolnagpal/aiswitch/releases/latest/download/aiswitch_darwin_arm64.tar.gz | tar xz
sudo mv aiswitch /usr/local/bin/

# macOS Intel
curl -L https://github.com/anmolnagpal/aiswitch/releases/latest/download/aiswitch_darwin_amd64.tar.gz | tar xz
sudo mv aiswitch /usr/local/bin/

# Linux amd64
curl -L https://github.com/anmolnagpal/aiswitch/releases/latest/download/aiswitch_linux_amd64.tar.gz | tar xz
sudo mv aiswitch /usr/local/bin/
```

### Option 3 — Build from source

```bash
git clone https://github.com/anmolnagpal/aiswitch.git
cd aiswitch
make install   # builds and copies to /usr/local/bin/
```

### Verify

```bash
aiswitch --version
```

---

## Shell integration

Because a child process cannot modify the parent shell's environment, `aiswitch` writes env vars to `~/.aiswitch/env.sh` (or `env.ps1` on Windows). The shell integration is a thin wrapper that sources that file automatically — so keys like `ANTHROPIC_API_KEY` are live in your **current** session right after every switch.

It also installs a **`cd` hook** that auto-switches profiles when you enter a directory containing a `.aiswitch` file.

### Automatic setup (recommended)

Run **once** after installing:

```bash
aiswitch setup
```

```
✓ Shell integration written to /Users/you/.zshrc

  Activate now without restarting your shell:

    source ~/.zshrc

  Future shells will load it automatically.
```

Options:

```bash
aiswitch setup --shell bash     # target a specific shell (auto-detected by default)
aiswitch setup --dry-run        # preview the line that would be added
aiswitch setup --force          # replace an existing block (after upgrading)
```

### Manual setup

| Shell | Profile file | Line to add |
|---|---|---|
| **Zsh** | `~/.zshrc` | `eval "$(aiswitch shell-init --shell zsh)"` |
| **Bash** (Linux) | `~/.bashrc` | `eval "$(aiswitch shell-init --shell bash)"` |
| **Bash** (macOS) | `~/.bash_profile` | `eval "$(aiswitch shell-init --shell bash)"` |
| **Fish** | `~/.config/fish/config.fish` | `aiswitch shell-init --shell fish \| source` |
| **PowerShell** | `$PROFILE` | `Invoke-Expression (aiswitch shell-init --shell powershell \| Out-String)` |

### How the integration works

```
aiswitch use work
  → writes ANTHROPIC_API_KEY, OPENAI_API_KEY … to ~/.aiswitch/env.sh
  → shell wrapper sources env.sh → keys live in current session ✓

cd ~/work-project            # directory has a .aiswitch file
  → cd hook calls `aiswitch detect`
  → profile switches automatically ✓
```

---

## Quick start

```bash
# 1. Install shell integration (once per machine)
aiswitch setup
source ~/.zshrc          # or restart your terminal

# 2. Add your accounts
aiswitch add             # interactive form — name, providers, IDE options

# 3. Switch profiles
aiswitch use work        # direct switch
aiswitch                 # interactive TUI picker

# 4. Check what's active
aiswitch current
```

---

## Per-project `.aiswitch` file

Works exactly like `.terraform-version` or `.nvmrc`.  
Commit a `.aiswitch` file to each repo so the right AI account activates on `cd`.

### Create

```bash
cd ~/my-work-project
aiswitch init            # interactive form — pick profile + optional model overrides
```

This creates a file that **contains no secrets — safe to commit**:

```yaml
# .aiswitch
profile: work

# Optional per-project model overrides (layer on top of the profile default)
claude:
  model: claude-opus-4-5
openai:
  model: gpt-4o
gemini:
  model: gemini-2.0-flash
github:
  email: me@company.com  # overrides git commit email in this repo
```

Plain-text shorthand also works (just the profile name):

```
work
```

### Auto-switch on `cd`

```
~/personal-project  $ echo $ANTHROPIC_API_KEY
sk-ant-personal-...

~/personal-project  $ cd ~/work-project
⬡ aiswitch → work

~/work-project      $ echo $ANTHROPIC_API_KEY
sk-ant-work-...
```

The hook is **silent** when there's no `.aiswitch` file, **skips** if already on the correct profile, and prints one line only when it actually switches.

### Apply manually

```bash
aiswitch detect           # apply nearest .aiswitch, show what changed
aiswitch detect --quiet   # same, but only prints the one-line indicator
```

---

## IDE integration

aiswitch patches `settings.json` for **Cursor** and **Windsurf** on every profile switch.  
No IDE restart required — VS Code-based editors hot-reload `settings.json`.

### Keys written per provider

| Provider | Key written to `settings.json` |
|---|---|
| Claude | `anthropic.apiKey`, `anthropic.defaultModel` |
| OpenAI | `openai.apiKey`, `openai.organization`, `openai.defaultModel` |
| Gemini | `googleGenerativeAI.apiKey`, `googleGenerativeAI.defaultModel` |

All other existing settings are preserved.

### Settings file locations

| IDE | macOS | Linux | Windows |
|---|---|---|---|
| Cursor | `~/Library/Application Support/Cursor/User/settings.json` | `~/.config/Cursor/User/settings.json` | `%APPDATA%\Cursor\User\settings.json` |
| Windsurf | `~/Library/Application Support/Windsurf/User/settings.json` | `~/.config/Windsurf/User/settings.json` | `%APPDATA%\Windsurf\User\settings.json` |

### Enable

During `aiswitch add`, the wizard detects installed IDEs and shows a multi-select:

```
? Patch IDE settings.json with API keys?
  [x] Cursor IDE  ✓ installed
  [ ] Windsurf IDE
```

Or enable manually in `~/.aiswitch/config.json`:

```json
"ide": { "cursor": true, "windsurf": true }
```

### Verify

```bash
aiswitch current
# Cursor    settings.json patched · sk-a...ey12
# Windsurf  not installed or not yet patched
```

---

## Commands

| Command | Description |
|---|---|
| `aiswitch` | Open the interactive profile picker |
| `aiswitch use <profile>` | Switch to a named profile |
| `aiswitch add [name]` | Add or update a profile (guided form) |
| `aiswitch list` | List all profiles in a table |
| `aiswitch remove <profile>` | Delete a profile |
| `aiswitch current` | Show the active profile and live system state |
| `aiswitch init` | Create a `.aiswitch` file in the current directory |
| `aiswitch detect [--quiet]` | Find and apply the nearest `.aiswitch` file |
| `aiswitch setup [--shell] [--dry-run] [--force]` | Write shell integration into your profile file |
| `aiswitch shell-init [--shell]` | Print shell integration code (manual setup) |

---

## Configuration reference

### Global config — `~/.aiswitch/config.json`

```json
{
  "active_profile": "work",
  "profiles": {
    "work": {
      "description": "Day-job accounts",
      "claude": {
        "api_key": "sk-ant-...",
        "default_model": "claude-opus-4-5"
      },
      "openai": {
        "api_key": "sk-proj-...",
        "org_id": "org-...",
        "default_model": "gpt-4o"
      },
      "gemini": {
        "api_key": "AIza...",
        "default_model": "gemini-2.0-flash",
        "project_id": "my-gcp-project"
      },
      "github": {
        "token": "ghp_...",
        "username": "work-octocat",
        "email": "me@company.com"
      },
      "ide": {
        "cursor": true,
        "windsurf": false
      }
    }
  }
}
```

The file is created and managed by the `add` wizard. You rarely need to edit it by hand.

### What each provider writes

| Provider | Env vars (`~/.aiswitch/env.sh`) | Files written |
|---|---|---|
| **Claude** | `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL` | `~/.anthropic/api_key`, `~/.claude/.credentials.json` |
| **OpenAI** | `OPENAI_API_KEY`, `OPENAI_ORG_ID`, `OPENAI_MODEL` | `~/.config/openai/api_key` |
| **Gemini** | `GEMINI_API_KEY`, `GOOGLE_API_KEY`, `GEMINI_MODEL`, `GOOGLE_CLOUD_PROJECT` | `~/.config/gemini/api_key` |
| **GitHub** | `GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_USER` | `~/.config/gh/hosts.yml`, `git config --global user.email` |
| **Cursor/Windsurf** | — | `settings.json` (`anthropic.apiKey`, `openai.apiKey`, `googleGenerativeAI.apiKey`) |

### Per-project `.aiswitch` file reference

| Key | Type | Description |
|---|---|---|
| `profile` | string | **Required.** Global profile to activate. |
| `claude.model` | string | Override `default_model` for Claude in this directory. |
| `openai.model` | string | Override `default_model` for OpenAI in this directory. |
| `gemini.model` | string | Override `default_model` for Gemini in this directory. |
| `github.email` | string | Override `git config user.email` in this directory. |

### Getting a GitHub token

[github.com/settings/tokens](https://github.com/settings/tokens) — create a classic PAT with:

- `repo` — git operations
- `read:user` — identity
- `copilot` — Copilot API (if your plan exposes it)

After switching, VS Code / Cursor / Windsurf Copilot picks up the new token on the next extension reload. For an immediate refresh: `Cmd/Ctrl+Shift+P` → **GitHub: Sign Out**, then sign back in.

---

## Security

API keys and tokens are stored in `~/.aiswitch/config.json` with mode `0600` (readable only by you). This is the same model used by the `gh` CLI.

**What to be careful about:**

- Do **not** commit `~/.aiswitch/config.json` — it contains secrets
- The per-project `.aiswitch` file **does not contain any secrets** — safe to commit
- `~/.aiswitch/env.sh` is also secret (it contains the active key) — it is in `.gitignore`

**Found a vulnerability?** Please do not open a public issue. Instead, email the maintainer directly or use [GitHub private vulnerability reporting](https://github.com/anmolnagpal/aiswitch/security/advisories/new).

OS keychain integration (`99designs/keyring`) is on the [roadmap](#roadmap).

---

## Contributing

Contributions of all kinds are welcome — bug reports, feature requests, documentation improvements, and code.

### First-time setup

```bash
git clone https://github.com/anmolnagpal/aiswitch.git
cd aiswitch
go mod download

# Install git hooks — blocks commits that fail lint or format checks
make hooks
```

### Development workflow

```bash
make build          # build ./aiswitch binary
make run ARGS="add" # run any command locally
make fmt            # auto-format all Go files
make lint           # run golangci-lint
make build-all      # cross-compile for all platforms
make hooks-check    # verify hooks are active
```

### Commit messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(openai):   add org_id support
fix(github):    correct hosts.yml path on Windows
docs:           improve shell integration examples
refactor:       extract provider helpers into shared package
chore:          update dependencies
ci:             pin golangci-lint to v2
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide, including how to skip a hook in an emergency.

### Good first issues

Look for issues labelled [`good first issue`](https://github.com/anmolnagpal/aiswitch/issues?q=label%3A%22good+first+issue%22) on GitHub — these are intentionally scoped for newcomers.

### Opening a PR

1. Fork the repo and create a branch: `git checkout -b feat/my-feature`
2. Make your changes, run `make lint` and `make build-all`
3. Push and open a PR against `main`
4. A maintainer will review within a few days

---

## Roadmap

Planned features, roughly in priority order:

- [ ] OS keychain integration — store secrets in macOS Keychain / Linux Secret Service / Windows Credential Manager (`99designs/keyring`)
- [ ] `brew install aiswitch` — Homebrew tap
- [ ] More providers — Ollama (local models), Azure OpenAI, AWS Bedrock
- [ ] `aiswitch remove-hook` — clean uninstall of shell integration
- [ ] Shell completion — `aiswitch use <TAB>` profile names

Have an idea? [Open a discussion](https://github.com/anmolnagpal/aiswitch/discussions) or a feature request issue.

---

## Related projects

| Project | What it does |
|---|---|
| [tfswitch](https://github.com/warrensbox/terraform-switcher) | Terraform version switcher — the original inspiration |
| [nvm](https://github.com/nvm-sh/nvm) | Node.js version switcher — same `.nvmrc` auto-switch pattern |
| [gh](https://github.com/cli/cli) | GitHub CLI — aiswitch manages its `hosts.yml` for multi-account |
| [direnv](https://github.com/direnv/direnv) | General-purpose per-directory env vars (complementary) |

---

## License

MIT © [Anmol Nagpal](https://github.com/anmolnagpal)

---

<div align="center">

If aiswitch saves you time, please consider giving it a ⭐ — it helps others discover the project.

[Report a bug](https://github.com/anmolnagpal/aiswitch/issues/new?template=bug_report.md) · [Request a feature](https://github.com/anmolnagpal/aiswitch/issues/new?template=feature_request.md) · [Start a discussion](https://github.com/anmolnagpal/aiswitch/discussions)

</div>
