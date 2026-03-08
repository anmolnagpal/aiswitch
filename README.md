<div align="center">

<h1>⚡ aiswitch</h1>

<p><strong>Switch between Claude, OpenAI, Gemini, and GitHub Copilot accounts in one command.<br/>
Works with Cursor, Windsurf, and any terminal tool.</strong><br/>
Like <a href="https://github.com/warrensbox/terraform-switcher">tfswitch</a>, but for AI.</p>

[![CI](https://github.com/anmolnagpal/aiswitch/actions/workflows/ci.yml/badge.svg)](https://github.com/anmolnagpal/aiswitch/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/anmolnagpal/aiswitch)](https://goreportcard.com/report/github.com/anmolnagpal/aiswitch)
[![GitHub Release](https://img.shields.io/github/v/release/anmolnagpal/aiswitch?include_prereleases&sort=semver)](https://github.com/anmolnagpal/aiswitch/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/anmolnagpal/aiswitch)](go.mod)
[![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](#installation)

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

## Why aiswitch?

Most developers juggle **multiple AI accounts** — a work Anthropic account on a premium plan, a personal Claude account, an OpenAI key for GPT-4o, a Gemini key for AI Studio experiments, and two or three GitHub accounts each with a different Copilot subscription. Switching between them today means:

- Manually editing `~/.bashrc` to swap `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY` …
- Running `gh auth switch` and hoping VS Code / Cursor / Windsurf picks it up
- Digging into **Cursor → Settings → Models** or **Windsurf Preferences** to re-enter keys by hand
- Forgetting which account is active mid-session and burning API quota on the wrong key

**aiswitch** solves this the same way `tfswitch` solved Terraform versions and `nvm` solved Node versions: one command to switch, a `.aiswitch` file in each project to make it automatic.

---

## Features

- **Interactive TUI** — fuzzy-searchable profile list with arrow-key navigation
- **Instant switching** — `aiswitch use work` switches in under 100ms
- **Per-project pinning** — commit a `.aiswitch` file; the profile switches automatically on `cd`
- **Auto-detect cd hook** — works with zsh, bash, fish, and PowerShell
- **Claude** — sets `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL`, writes `~/.anthropic/api_key`, patches Claude Code credentials
- **OpenAI** — sets `OPENAI_API_KEY`, `OPENAI_ORG_ID`, `OPENAI_MODEL`, writes `~/.config/openai/api_key`
- **Gemini** — sets `GEMINI_API_KEY`, `GOOGLE_API_KEY`, `GEMINI_MODEL`, `GOOGLE_CLOUD_PROJECT`, writes `~/.config/gemini/api_key`
- **GitHub Copilot** — updates `~/.config/gh/hosts.yml`, sets `GITHUB_TOKEN` / `GH_TOKEN`, optionally updates `git config user.email`
- **Cursor IDE** — patches `settings.json` with `anthropic.apiKey`, `openai.apiKey`, `googleGenerativeAI.apiKey` — no restart needed
- **Windsurf IDE** — same `settings.json` patching for Windsurf (Codeium)
- **Cross-platform** — macOS (Intel + Apple Silicon), Linux (amd64 + arm64), Windows (amd64)
- **Zero runtime deps** — single self-contained binary, ~5 MB

---

## Installation

### Recommended — `go install`

```bash
go install github.com/anmolnagpal/aiswitch@latest
```

### Download pre-built binary

Head to [Releases](https://github.com/anmolnagpal/aiswitch/releases) and grab the archive for your platform, then move the binary onto your `$PATH`.

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

### Build from source

```bash
git clone https://github.com/anmolnagpal/aiswitch.git
cd aiswitch
make install   # builds and copies to /usr/local/bin/
```

---

## Shell integration

Because a child process cannot modify the parent shell's environment, `aiswitch` writes env vars to `~/.aiswitch/env.sh` (or `env.ps1` on Windows). The shell integration is a thin wrapper that sources that file automatically — so `ANTHROPIC_API_KEY` is live in your current session right after every switch.

It also installs a **`cd` hook** that auto-switches profiles when you enter a directory containing a `.aiswitch` file.

### Automatic setup (recommended)

Run this **once** after installing. It detects your shell and appends the one-liner to the right profile file:

```bash
aiswitch setup
```

Output:

```
✓ Shell integration written to /Users/you/.zshrc

  Activate now without restarting your shell:

    source ~/.zshrc

  Future shells will load it automatically.
```

Options:

```bash
aiswitch setup --shell bash        # target a specific shell
aiswitch setup --dry-run           # preview without writing
aiswitch setup --force             # remove old block and re-install
```

### Manual setup

If you prefer to control exactly what goes in your profile, add the one-liner yourself:

| Shell | File | Line to add |
|---|---|---|
| **Zsh** | `~/.zshrc` | `eval "$(aiswitch shell-init --shell zsh)"` |
| **Bash** (Linux) | `~/.bashrc` | `eval "$(aiswitch shell-init --shell bash)"` |
| **Bash** (macOS) | `~/.bash_profile` | `eval "$(aiswitch shell-init --shell bash)"` |
| **Fish** | `~/.config/fish/config.fish` | `aiswitch shell-init --shell fish \| source` |
| **PowerShell** | `$PROFILE` | `Invoke-Expression (aiswitch shell-init --shell powershell \| Out-String)` |

### What the integration does

```bash
# 1. Wrapper function — env vars apply in the CURRENT session after every switch:
aiswitch use work   # → runs binary, then sources ~/.aiswitch/env.sh automatically

# 2. cd hook — auto-switches when you enter a project directory:
cd ~/work-project   # → aiswitch detect runs, reads .aiswitch, applies profile
```

---

## Quick start

```bash
# 1. Install shell integration (once)
aiswitch setup          # auto-detects your shell, appends to ~/.zshrc / ~/.bashrc / etc.
source ~/.zshrc         # activate immediately (or open a new tab)

# 2. Add profiles for your accounts
aiswitch add            # guided interactive form

# 3. Switch to a profile
aiswitch use work       # direct
aiswitch                # interactive TUI picker

# 4. Verify what's active
aiswitch current
```

---

## Per-project `.aiswitch` file

Works exactly like `.terraform-version` (tfswitch) or `.nvmrc` (nvm). Commit a `.aiswitch` file to each project so the right AI account activates automatically.

### Create it

```bash
cd ~/my-work-project
aiswitch init           # interactive form — pick profile + optional overrides
```

This writes a `.aiswitch` file. **It contains no secrets — safe to commit.**

```yaml
# .aiswitch
# aiswitch project config — safe to commit, contains no secrets
profile: work

claude:
  model: claude-opus-4-5       # optional: pin a model for this project

openai:
  model: gpt-4o                # optional: pin an OpenAI model

gemini:
  model: gemini-2.0-flash      # optional: pin a Gemini model

github:
  email: me@company.com        # optional: override git commit email
```

Minimal plain-text form also works:

```
work
```

### Auto-switch on `cd`

With shell integration active, aiswitch switches the moment you enter the directory:

```
~/personal-project  $ echo $ANTHROPIC_API_KEY
sk-ant-personal-...

~/personal-project  $ cd ~/work-project
⬡ aiswitch → work

~/work-project      $ echo $ANTHROPIC_API_KEY
sk-ant-work-...
```

The hook is **silent** when there's no `.aiswitch` file, **skips** if already on the right profile, and shows a one-liner only when it actually switches.

### Apply manually

```bash
aiswitch detect           # verbose output
aiswitch detect --quiet   # one-line indicator (same as what the hook shows)
```

---

## Commands

| Command | Description |
|---|---|
| `aiswitch` | Open the interactive profile selector |
| `aiswitch use <profile>` | Switch to a profile directly |
| `aiswitch add [name]` | Add or update a profile (guided form) |
| `aiswitch list` | List all profiles in a table |
| `aiswitch remove <profile>` | Delete a profile |
| `aiswitch current` | Show the active profile and live system state |
| `aiswitch init` | Create a `.aiswitch` file in the current directory |
| `aiswitch detect [--quiet]` | Find and apply the nearest `.aiswitch` file |
| `aiswitch setup [--shell] [--dry-run] [--force]` | Install shell integration into your profile file |
| `aiswitch shell-init [--shell]` | Print shell integration code (for manual setup) |

---

## Configuration reference

Global profiles are stored in `~/.aiswitch/config.json` (mode `0600`).

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
      "github": {
        "token": "ghp_...",
        "username": "work-octocat",
        "email": "me@company.com"
      }
    },
    "personal": {
      "claude": {
        "api_key": "sk-ant-..."
      },
      "github": {
        "token": "ghp_...",
        "username": "personal-octocat"
      }
    }
  }
}
```

### What each provider touches

| Provider | Files / env vars written |
|---|---|
| **Claude** | `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL` → `~/.aiswitch/env.sh` · `~/.anthropic/api_key` · `~/.claude/.credentials.json` (Claude Code) |
| **GitHub** | `GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_USER` → `~/.aiswitch/env.sh` · `~/.config/gh/hosts.yml` (active user for `gh` CLI) · `git config --global user.email` |

> **Security note:** API keys and tokens are stored in `~/.aiswitch/config.json` with mode `0600` (owner-readable only). This is the same approach used by the `gh` CLI. OS keychain integration is on the roadmap.

### GitHub Copilot path by OS

| OS | hosts.yml location |
|---|---|
| macOS / Linux | `~/.config/gh/hosts.yml` |
| Windows | `%APPDATA%\GitHub CLI\hosts.yml` |

After switching, VS Code Copilot picks up the new token on the next window launch. If you need it immediately: `Cmd/Ctrl+Shift+P` → **GitHub: Sign Out**, then sign back in.

### Getting a GitHub token

[github.com/settings/tokens](https://github.com/settings/tokens) — create a classic PAT with:

- `repo` — git operations
- `read:user` — identity
- `copilot` — Copilot API (if your plan exposes it)

---

## IDE integration

aiswitch can patch the `settings.json` of **Cursor** and **Windsurf** on every profile switch, so the IDE's built-in AI features automatically use the right API keys without any manual reconfiguration.

### How it works

When you enable an IDE in your profile, aiswitch writes these keys into the IDE's `settings.json`:

| Provider | Setting key written |
|---|---|
| Claude | `anthropic.apiKey`, `anthropic.defaultModel` |
| OpenAI | `openai.apiKey`, `openai.organization`, `openai.defaultModel` |
| Gemini | `googleGenerativeAI.apiKey`, `googleGenerativeAI.defaultModel` |

All other existing settings are preserved. The IDE does **not** need to be restarted — VS Code-based editors hot-reload `settings.json`.

### Settings file locations

| IDE | macOS | Linux | Windows |
|---|---|---|---|
| Cursor | `~/Library/Application Support/Cursor/User/settings.json` | `~/.config/Cursor/User/settings.json` | `%APPDATA%\Cursor\User\settings.json` |
| Windsurf | `~/Library/Application Support/Windsurf/User/settings.json` | `~/.config/Windsurf/User/settings.json` | `%APPDATA%\Windsurf\User\settings.json` |

### Enable during `aiswitch add`

The `add` wizard automatically detects installed IDEs (marked **✓ installed**) and presents them as a multi-select:

```
? Patch IDE settings.json with API keys?
  [x] Cursor IDE  ✓ installed
  [ ] Windsurf IDE
```

### Enable manually

Edit `~/.aiswitch/config.json` and add an `ide` block to any profile:

```json
{
  "active_profile": "work",
  "profiles": {
    "work": {
      "claude": { "api_key": "sk-ant-..." },
      "openai": { "api_key": "sk-proj-..." },
      "ide": {
        "cursor": true,
        "windsurf": true
      }
    }
  }
}
```

### Verify

```bash
aiswitch current
# …
# Cursor    settings.json patched · sk-a...ey12
# Windsurf  settings.json patched · sk-p...5678
```

---

## Contributing

Contributions are welcome! Please open an issue first for large changes.

```bash
git clone https://github.com/anmolnagpal/aiswitch.git
cd aiswitch
go mod download

# Activate git hooks — blocks commits that fail lint (run this once after cloning)
make hooks

make build          # build ./bin/aiswitch
make fmt            # auto-format all Go files
make lint           # run golangci-lint
make run ARGS="--help"
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for commit message format and full workflow.

### Roadmap

- [ ] OS keychain integration for secrets (`99designs/keyring`)
- [ ] `brew install aiswitch` (Homebrew tap)
- [ ] More providers: Ollama, Azure OpenAI, Bedrock

---

## Related projects

| Project | Does what |
|---|---|
| [tfswitch](https://github.com/warrensbox/terraform-switcher) | Terraform version switcher — the original inspiration |
| [nvm](https://github.com/nvm-sh/nvm) | Node.js version switcher — same `.nvmrc` pattern |
| [gh](https://github.com/cli/cli) | GitHub CLI — aiswitch manages its `hosts.yml` |

---

## License

MIT © [Anmol Nagpal](https://github.com/anmolnagpal)
