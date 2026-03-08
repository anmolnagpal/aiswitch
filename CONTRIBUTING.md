# Contributing to aiswitch

Thank you for your interest in contributing!

## Setup

```bash
git clone https://github.com/anmolnagpal/aiswitch.git
cd aiswitch
go mod download

# Activate git hooks (required — blocks commits that fail lint)
make hooks
```

`make hooks` sets `core.hooksPath = .githooks` so two hooks run automatically:

| Hook | What it checks |
|---|---|
| `pre-commit` | `gofmt`, `go vet`, `go build`, `golangci-lint` |
| `commit-msg` | [Conventional Commits](https://www.conventionalcommits.org/) format |

## Workflow

```bash
make build          # build ./bin/aiswitch
make run ARGS="--help"
make fmt            # auto-format all Go files
make lint           # run golangci-lint
make hooks-check    # verify hooks are active
```

## Commit message format

```
<type>(<scope>): <short description>

feat(cli):     new feature
fix(github):   bug fix
docs:          documentation only
refactor:      no behaviour change
test:          tests only
chore:         build / tooling
ci:            GitHub Actions
```

Examples:

```
feat(localfile): support plain-text .aiswitch shorthand
fix(github): use correct hosts.yml path on Windows
docs: add PowerShell integration example to README
```

## Before opening a PR

- `make lint` passes with zero issues
- `make build-all` cross-compiles cleanly for all platforms
- New behaviour has a comment explaining *why*, not *what*

## Skipping a hook (emergencies only)

```bash
git commit --no-verify
```
