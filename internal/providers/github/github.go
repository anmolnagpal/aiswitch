// Package github applies GitHub account settings so the gh CLI (and therefore
// GitHub Copilot) uses the selected profile's credentials.
//
// Strategy:
//  1. Write env.sh and env.ps1 — sets GITHUB_TOKEN / GH_TOKEN in the shell.
//  2. Patch ~/.config/gh/hosts.yml (Linux/macOS) or
//     %APPDATA%\GitHub CLI\hosts.yml (Windows) to set the active user.
//  3. Optionally update the global git config user.name/email.
package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/anmolnagpal/aiswitch/internal/config"
)

// Apply writes all GitHub-related credential files for the given config.
func Apply(cfg config.GitHubConfig, paths config.EnvPaths) error {
	if err := writeShFragment(cfg, paths.Sh); err != nil {
		return err
	}
	if err := writePS1Fragment(cfg, paths.PS1); err != nil {
		return err
	}
	if err := patchGHHosts(cfg); err != nil {
		return fmt.Errorf("patching gh hosts: %w", err)
	}
	if cfg.Username != "" || cfg.Email != "" {
		_ = setGitIdentity(cfg.Username, cfg.Email)
	}
	return nil
}

// Detect returns the currently active GitHub username from the gh CLI, or "".
func Detect() string {
	out, err := exec.Command("gh", "auth", "status", "--hostname", "github.com").CombinedOutput()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Logged in to github.com account ") {
			parts := strings.Fields(line)
			if len(parts) >= 7 {
				return parts[6]
			}
		}
	}
	return ""
}

// ─── gh CLI hosts.yml ────────────────────────────────────────────────────────

type ghHosts map[string]ghHost

type ghHost struct {
	Users       map[string]ghUser `yaml:"users,omitempty"`
	GitProtocol string            `yaml:"git_protocol,omitempty"`
	User        string            `yaml:"user,omitempty"`
}

type ghUser struct {
	OAuthToken  string `yaml:"oauth_token,omitempty"`
	GitProtocol string `yaml:"git_protocol,omitempty"`
}

// ghHostsPath returns the OS-appropriate path to the gh CLI hosts.yml file.
//
//   - Windows: %APPDATA%\GitHub CLI\hosts.yml
//   - macOS / Linux: ~/.config/gh/hosts.yml
func ghHostsPath() (string, error) {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("%%APPDATA%% is not set")
		}
		return filepath.Join(appData, "GitHub CLI", "hosts.yml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gh", "hosts.yml"), nil
}

func patchGHHosts(cfg config.GitHubConfig) error {
	path, err := ghHostsPath()
	if err != nil {
		return err
	}

	hosts := ghHosts{}
	data, err := os.ReadFile(path)
	if err == nil {
		_ = yaml.Unmarshal(data, &hosts)
	}

	host := hosts["github.com"]
	if host.Users == nil {
		host.Users = map[string]ghUser{}
	}
	host.Users[cfg.Username] = ghUser{
		OAuthToken:  cfg.Token,
		GitProtocol: "https",
	}
	host.User = cfg.Username
	if host.GitProtocol == "" {
		host.GitProtocol = "https"
	}
	hosts["github.com"] = host

	out, err := yaml.Marshal(hosts)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o600)
}

// ─── env.sh (bash/zsh/fish) ───────────────────────────────────────────────────

func writeShFragment(cfg config.GitHubConfig, path string) error {
	block := fmt.Sprintf("export GITHUB_TOKEN=%q\n", cfg.Token)
	block += fmt.Sprintf("export GH_TOKEN=%q\n", cfg.Token)
	if cfg.Username != "" {
		block += fmt.Sprintf("export GITHUB_USER=%q\n", cfg.Username)
	}
	return mergeIntoFile(path, "# aiswitch:github", "# /aiswitch:github", block,
		"# aiswitch env — source this file or add it to your shell profile\n")
}

// ─── env.ps1 (PowerShell) ─────────────────────────────────────────────────────

func writePS1Fragment(cfg config.GitHubConfig, path string) error {
	block := fmt.Sprintf("$env:GITHUB_TOKEN = %q\n", cfg.Token)
	block += fmt.Sprintf("$env:GH_TOKEN = %q\n", cfg.Token)
	if cfg.Username != "" {
		block += fmt.Sprintf("$env:GITHUB_USER = %q\n", cfg.Username)
	}
	return mergeIntoFile(path, "# aiswitch:github", "# /aiswitch:github", block,
		"# aiswitch env — dot-source this file: . ~/.aiswitch/env.ps1\n")
}

// ─── git identity ─────────────────────────────────────────────────────────────

func setGitIdentity(name, email string) error {
	if name != "" {
		if err := exec.Command("git", "config", "--global", "user.name", name).Run(); err != nil {
			return err
		}
	}
	if email != "" {
		return exec.Command("git", "config", "--global", "user.email", email).Run()
	}
	return nil
}

// ─── shared env file helpers ──────────────────────────────────────────────────

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func mergeIntoFile(path, startMarker, endMarker, block, header string) error {
	existing := readFile(path)
	if existing == "" {
		existing = header
	}
	start := startMarker + "\n"
	startIdx := indexOf(existing, start)
	endIdx := indexOf(existing, endMarker)

	var result string
	if startIdx == -1 || endIdx == -1 {
		sep := ""
		if len(existing) > 0 && existing[len(existing)-1] != '\n' {
			sep = "\n"
		}
		result = existing + sep + start + block + endMarker + "\n"
	} else {
		before := existing[:startIdx]
		after := existing[endIdx+len(endMarker):]
		if len(after) > 0 && after[0] == '\n' {
			after = after[1:]
		}
		result = before + start + block + endMarker + "\n" + after
	}
	return os.WriteFile(path, []byte(result), 0o600)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
