package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/config"
)

var shellFlag string

var shellInitCmd = &cobra.Command{
	Use:   "shell-init",
	Short: "Print shell integration code to eval in your profile",
	Long: `Print shell code that:
  1. Wraps the aiswitch binary so env vars apply in the current session.
  2. Installs a cd hook that auto-switches profiles when you enter a
     directory containing a .aiswitch file (like tfswitch / nvm).

Bash / Zsh — add to ~/.zshrc or ~/.bashrc:
  eval "$(aiswitch shell-init)"

Fish — add to ~/.config/fish/config.fish:
  aiswitch shell-init --shell fish | source

PowerShell — add to $PROFILE:
  Invoke-Expression (aiswitch shell-init --shell powershell | Out-String)
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sh := shellFlag
		if sh == "auto" || sh == "" {
			sh = detectShell()
		}

		paths, err := config.GetEnvPaths()
		if err != nil {
			return err
		}

		switch sh {
		case "powershell", "pwsh":
			printPowerShell(paths.PS1)
		case "fish":
			printFish(paths.Sh)
		case "zsh":
			printZsh(paths.Sh)
		default: // bash, posix sh
			printBash(paths.Sh)
		}
		return nil
	},
}

func init() {
	shellInitCmd.Flags().StringVar(&shellFlag, "shell", "auto",
		"Shell to target: auto, bash, zsh, fish, powershell")
}

// detectShell guesses the running shell from $SHELL (Unix) or OS (Windows).
func detectShell() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	shellPath := os.Getenv("SHELL")
	switch {
	case strings.Contains(shellPath, "fish"):
		return "fish"
	case strings.Contains(shellPath, "zsh"):
		return "zsh"
	default:
		return "bash"
	}
}

// ─── Bash ────────────────────────────────────────────────────────────────────

func printBash(envPath string) {
	fmt.Printf(`# aiswitch shell integration — paste into ~/.bashrc
# 1. Wrapper: sources env.sh after every aiswitch call.
# 2. cd hook: auto-detects .aiswitch when you enter a directory.

aiswitch() {
  command aiswitch "$@"
  local _exit=$?
  if [ -f %[1]q ]; then
    # shellcheck source=/dev/null
    . %[1]q
  fi
  return $_exit
}

_aiswitch_hook() {
  command aiswitch detect --quiet
  if [ -f %[1]q ]; then
    # shellcheck source=/dev/null
    . %[1]q
  fi
}

# Run the hook once on shell startup, then on every cd.
_aiswitch_hook
PROMPT_COMMAND="${PROMPT_COMMAND:+${PROMPT_COMMAND}$'\n'}_aiswitch_hook"
`, envPath)
}

// ─── Zsh ─────────────────────────────────────────────────────────────────────

func printZsh(envPath string) {
	fmt.Printf(`# aiswitch shell integration — paste into ~/.zshrc
# 1. Wrapper: sources env.sh after every aiswitch call.
# 2. chpwd hook: auto-detects .aiswitch when you enter a directory.

aiswitch() {
  command aiswitch "$@"
  local _exit=$?
  if [ -f %[1]q ]; then
    # shellcheck source=/dev/null
    . %[1]q
  fi
  return $_exit
}

_aiswitch_hook() {
  command aiswitch detect --quiet
  if [ -f %[1]q ]; then
    . %[1]q
  fi
}

# autoload is needed for add-zsh-hook.
autoload -U add-zsh-hook
# Run on every directory change.
add-zsh-hook chpwd _aiswitch_hook
# Run once when the shell starts.
_aiswitch_hook
`, envPath)
}

// ─── Fish ─────────────────────────────────────────────────────────────────────

func printFish(envPath string) {
	fmt.Printf(`# aiswitch shell integration — paste into ~/.config/fish/config.fish
# 1. Wrapper: sources env.sh after every aiswitch call.
# 2. PWD hook: auto-detects .aiswitch when you enter a directory.

function aiswitch
  command aiswitch $argv
  if test -f %[1]q
    source %[1]q
  end
end

function _aiswitch_hook --on-variable PWD
  command aiswitch detect --quiet
  if test -f %[1]q
    source %[1]q
  end
end

# Run once on shell startup.
_aiswitch_hook
`, envPath)
}

// ─── PowerShell ───────────────────────────────────────────────────────────────

func printPowerShell(ps1Path string) {
	fmt.Printf(`# aiswitch shell integration — paste into $PROFILE
# 1. Wrapper: dot-sources env.ps1 after every aiswitch call.
# 2. Set-Location override: auto-detects .aiswitch when you cd.

function Invoke-AiSwitch {
  & (Get-Command aiswitch -CommandType Application).Source @args
  if (Test-Path %[1]q) {
    . %[1]q
  }
}
Set-Alias -Name aiswitch -Value Invoke-AiSwitch -Force

function Set-Location {
  Microsoft.PowerShell.Management\Set-Location @args
  & (Get-Command aiswitch -CommandType Application) detect --quiet | Out-Null
  if (Test-Path %[1]q) {
    . %[1]q
  }
}

# Run once on shell startup.
& (Get-Command aiswitch -CommandType Application) detect --quiet | Out-Null
if (Test-Path %[1]q) { . %[1]q }
`, ps1Path)
}
