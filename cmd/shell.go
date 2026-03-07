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
	Long: `Print a shell wrapper so env vars are applied in the current shell
session automatically after every profile switch.

Bash / Zsh / Fish (add to ~/.zshrc or ~/.bashrc):
  eval "$(aiswitch shell-init)"

PowerShell (add to $PROFILE):
  Invoke-Expression (aiswitch shell-init --shell powershell | Out-String)

Fish (add to ~/.config/fish/config.fish):
  aiswitch shell-init --shell fish | source
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
		default: // bash, zsh, posix sh
			printBash(paths.Sh)
		}
		return nil
	},
}

func init() {
	shellInitCmd.Flags().StringVar(&shellFlag, "shell", "auto",
		`Shell to target: auto, bash, zsh, fish, powershell`)
}

// detectShell guesses the running shell from environment variables.
func detectShell() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	// $SHELL is set by most Unix login shells.
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


func printBash(envPath string) {
	fmt.Printf(`# aiswitch shell integration — paste into ~/.zshrc or ~/.bashrc
aiswitch() {
  command aiswitch "$@"
  local _exit=$?
  if [ -f %q ]; then
    # shellcheck source=/dev/null
    . %q
  fi
  return $_exit
}
`, envPath, envPath)
}

func printFish(envPath string) {
	fmt.Printf(`# aiswitch shell integration — paste into ~/.config/fish/config.fish
function aiswitch
  command aiswitch $argv
  if test -f %q
    source %q
  end
end
`, envPath, envPath)
}

func printPowerShell(ps1Path string) {
	fmt.Printf(`# aiswitch shell integration — paste into $PROFILE
function aiswitch {
  & (Get-Command aiswitch -CommandType Application).Source @args
  if (Test-Path %q) {
    . %q
  }
}
`, ps1Path, ps1Path)
}
