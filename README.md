# aiswitch

> Switch between Claude and GitHub Copilot accounts instantly — like `tfswitch`, but for AI.

If you have multiple Anthropic accounts (personal, work, client), or multiple GitHub accounts with different Copilot subscriptions, `aiswitch` lets you flip between them with a single command.

```
$ aiswitch
┌─────────────────────────────────────────────────┐
│  Switch AI Profile                              │
└─────────────────────────────────────────────────┘
  ▶ ● work         Claude + GitHub   (active)
    ○ personal     Claude + GitHub
    ○ client-x     Claude only
    ○ open-source  GitHub only

  ↑/↓ navigate  •  enter select  •  / filter  •  q quit
```

---

## What it does

| Service | What switches |
|---|---|
| **Claude / Anthropic** | Sets `ANTHROPIC_API_KEY` (+ optionally `ANTHROPIC_MODEL`), writes `~/.anthropic/api_key`, patches `~/.claude/.credentials.json` (Claude Code) |
| **GitHub / Copilot** | Sets `GITHUB_TOKEN` / `GH_TOKEN`, updates `~/.config/gh/hosts.yml` active user (used by `gh` CLI and VS Code GitHub extension), optionally updates global `git config user.email` |

---

## Installation

### Build from source

```bash
git clone https://github.com/anmolnagpal/aiswitch
cd aiswitch
make install          # builds and copies to /usr/local/bin/aiswitch
```

### Cross-platform release builds

```bash
make build-all        # macOS (intel + apple silicon), Linux (amd64 + arm64), Windows (amd64)
make build-mac        # macOS only
make build-linux      # Linux only
make build-windows    # Windows only
```

Binaries appear in `./bin/`.

---

## Shell integration (required for env vars)

Because a child process cannot set env vars in the parent shell, `aiswitch` writes
env vars to `~/.aiswitch/env.sh` (or `env.ps1` on Windows). The shell integration
wraps the binary so that file is sourced automatically after every switch.

### Bash / Zsh (macOS, Linux)

Add to `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(aiswitch shell-init)"
```

Then reload: `source ~/.zshrc`

### Fish

Add to `~/.config/fish/config.fish`:

```fish
aiswitch shell-init --shell fish | source
```

### PowerShell (Windows)

Add to `$PROFILE` (run `echo $PROFILE` to find the path):

```powershell
Invoke-Expression (aiswitch shell-init --shell powershell | Out-String)
```

### Manual fallback (any shell)

```bash
# source after each switch yourself
source ~/.aiswitch/env.sh          # bash/zsh/fish
. $HOME/.aiswitch/env.ps1          # PowerShell
```

---

## Quick start

```bash
# 1. Add your first profile
aiswitch add

# 2. Switch to it
aiswitch use work

# 3. Check what's active
aiswitch current
```

---

## Commands

| Command | Description |
|---|---|
| `aiswitch` | Interactive profile selector (arrow keys + enter) |
| `aiswitch use <profile>` | Switch directly without the interactive UI |
| `aiswitch add [name]` | Add or update a profile (guided form) |
| `aiswitch list` | List all profiles in a table |
| `aiswitch remove <profile>` | Delete a profile |
| `aiswitch current` | Show the active profile and live system state |
| `aiswitch shell-init` | Print shell integration code |

---

## Configuration

Profiles are stored in `~/.aiswitch/config.json`.

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
        "username": "work-user",
        "email": "me@company.com"
      }
    },
    "personal": {
      "claude": {
        "api_key": "sk-ant-..."
      },
      "github": {
        "token": "ghp_...",
        "username": "personal-user"
      }
    }
  }
}
```

> **Security note:** API keys and tokens are stored in plaintext in `~/.aiswitch/config.json` (mode `0600`, readable only by your user). This is the same approach used by the `gh` CLI and many other developer tools. Keyring integration is planned for a future release.

---

## GitHub Copilot notes

`aiswitch` switches the active user in the gh CLI config file:

| OS | Path |
|---|---|
| macOS / Linux | `~/.config/gh/hosts.yml` |
| Windows | `%APPDATA%\GitHub CLI\hosts.yml` |

This affects:

- The `gh` CLI (runs as the new user immediately)
- **VS Code GitHub Copilot** — VS Code reads its GitHub session from the OS keychain. After switching you may need to sign out and back in once inside VS Code (`Cmd+Shift+P → GitHub: Sign Out`), or simply open a new VS Code window (VS Code re-reads the gh token on startup).

### Getting a GitHub token

Go to [github.com/settings/tokens](https://github.com/settings/tokens) and create a fine-grained or classic PAT with these scopes:

- `repo` — for git operations
- `read:user` — for user identity
- `copilot` — for Copilot API access (if available on your plan)

---

## License

MIT
