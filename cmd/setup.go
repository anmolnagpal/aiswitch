package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/ui"
)

// Markers wrap the aiswitch block so we can detect and re-install cleanly.
const (
	markerStart = "# >>> aiswitch shell integration >>>"
	markerEnd   = "# <<< aiswitch shell integration <<<"
)

var (
	setupDryRun bool
	setupShell  string
	setupForce  bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install shell integration into your shell profile (one-liner)",
	Long: `Append the aiswitch one-liner into your shell profile so that:

  • env vars (ANTHROPIC_API_KEY, OPENAI_API_KEY, …) are applied in the
    current session right after every aiswitch call.
  • Entering a directory with a .aiswitch file automatically switches the
    profile (like tfswitch / nvm's .nvmrc).

Shell profiles written to:

  Zsh         ~/.zshrc
  Bash        ~/.bashrc  (or ~/.bash_profile on macOS)
  Fish        ~/.config/fish/config.fish
  PowerShell  $PROFILE

Run with --dry-run to preview without writing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sh := setupShell
		if sh == "auto" || sh == "" {
			sh = detectShell()
		}

		rcPath, lineToAdd := rcFileAndLine(sh)
		if rcPath == "" {
			return fmt.Errorf("could not determine shell profile path for shell %q", sh)
		}

		// Expand ~ manually since os.WriteFile doesn't do it.
		if strings.HasPrefix(rcPath, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			rcPath = filepath.Join(home, rcPath[2:])
		}

		block := markerStart + "\n" + lineToAdd + "\n" + markerEnd + "\n"

		// ── dry-run ──────────────────────────────────────────────────────────
		if setupDryRun {
			fmt.Println(ui.StyleHint.Render("Dry-run — would append to " + rcPath + ":"))
			fmt.Println()
			fmt.Println(block)
			return nil
		}

		// ── already installed? ────────────────────────────────────────────────
		existing := readFileString(rcPath)
		if strings.Contains(existing, markerStart) && !setupForce {
			fmt.Println(ui.StyleSuccess.Render("✓ aiswitch is already set up in " + rcPath))
			fmt.Println(ui.StyleHint.Render("  Re-run with --force to overwrite, or edit the file manually."))
			return nil
		}

		// Remove old block when --force is used.
		if setupForce && strings.Contains(existing, markerStart) {
			existing = removeBlock(existing, markerStart, markerEnd)
			if err := os.WriteFile(rcPath, []byte(existing), 0o644); err != nil {
				return fmt.Errorf("removing old block from %s: %w", rcPath, err)
			}
		}

		// ── append ───────────────────────────────────────────────────────────
		if err := os.MkdirAll(filepath.Dir(rcPath), 0o755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", rcPath, err)
		}

		f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("opening %s: %w", rcPath, err)
		}
		defer f.Close()

		// Ensure we start on a fresh line.
		if existing != "" && !strings.HasSuffix(existing, "\n") {
			_, _ = fmt.Fprintln(f)
		}
		_, _ = fmt.Fprint(f, "\n"+block)

		fmt.Println(ui.StyleSuccess.Render("✓ Shell integration written to " + rcPath))
		fmt.Println()
		fmt.Println("  Activate now without restarting your shell:")
		fmt.Println()

		switch sh {
		case "fish":
			fmt.Println(ui.StyleHint.Render("    source " + rcPath))
		case "powershell", "pwsh":
			fmt.Println(ui.StyleHint.Render("    . $PROFILE"))
		default:
			fmt.Println(ui.StyleHint.Render("    source " + rcPath))
		}
		fmt.Println()
		fmt.Println(ui.StyleMuted.Render("  Future shells will load it automatically."))
		return nil
	},
}

func init() {
	setupCmd.Flags().StringVar(&setupShell, "shell", "auto",
		"Shell to target: auto, bash, zsh, fish, powershell")
	setupCmd.Flags().BoolVar(&setupDryRun, "dry-run", false,
		"Print what would be written without modifying any file")
	setupCmd.Flags().BoolVar(&setupForce, "force", false,
		"Remove and re-install the integration block even if already present")
}

// rcFileAndLine returns the shell profile path and the single line to insert.
func rcFileAndLine(sh string) (path, line string) {
	switch sh {
	case "zsh":
		return "~/.zshrc",
			`eval "$(aiswitch shell-init --shell zsh)"`

	case "bash":
		if runtime.GOOS == "darwin" {
			// macOS bash login shells source ~/.bash_profile, not ~/.bashrc.
			return "~/.bash_profile",
				`eval "$(aiswitch shell-init --shell bash)"`
		}
		return "~/.bashrc",
			`eval "$(aiswitch shell-init --shell bash)"`

	case "fish":
		return "~/.config/fish/config.fish",
			`aiswitch shell-init --shell fish | source`

	case "powershell", "pwsh":
		// $PROFILE is expanded at runtime; we return the canonical path.
		home, _ := os.UserHomeDir()
		var psProfile string
		if runtime.GOOS == "windows" {
			psProfile = filepath.Join(home, "Documents", "PowerShell",
				"Microsoft.PowerShell_profile.ps1")
		} else {
			psProfile = filepath.Join(home, ".config", "powershell",
				"Microsoft.PowerShell_profile.ps1")
		}
		return psProfile,
			`Invoke-Expression (aiswitch shell-init --shell powershell | Out-String)`

	default:
		return "", ""
	}
}

func readFileString(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// removeBlock strips the lines from startMarker through endMarker (inclusive).
func removeBlock(s, startMarker, endMarker string) string {
	lines := strings.Split(s, "\n")
	var out []string
	skip := false
	for _, line := range lines {
		if strings.TrimSpace(line) == startMarker {
			skip = true
		}
		if !skip {
			out = append(out, line)
		}
		if strings.TrimSpace(line) == endMarker {
			skip = false
		}
	}
	return strings.Join(out, "\n")
}
