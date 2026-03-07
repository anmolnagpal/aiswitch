// Package claude applies Claude / Anthropic profile settings to the local system.
//
// Strategy (applied in order):
//  1. Write ~/.aiswitch/env.sh and env.ps1 — sourced by the shell wrapper.
//  2. Write ~/.anthropic/api_key — read by the Anthropic Python/Node SDKs.
//  3. Patch ~/.claude/.credentials.json — read by Claude Code (claude.ai CLI).
package claude

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/anmolnagpal/aiswitch/internal/config"
)

// Apply writes all Claude-related credential files for the given config.
func Apply(cfg config.ClaudeConfig, paths config.EnvPaths) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := writeShFragment(cfg, paths.Sh); err != nil {
		return err
	}
	if err := writePS1Fragment(cfg, paths.PS1); err != nil {
		return err
	}
	if err := writeAnthropicAPIKey(home, cfg.APIKey); err != nil {
		return err
	}
	if err := patchClaudeCredentials(home, cfg.APIKey); err != nil {
		return err
	}
	return nil
}

// Detect returns the API key currently in use, or an empty string if unknown.
func Detect() string {
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key
	}
	home, _ := os.UserHomeDir()
	if home == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, ".anthropic", "api_key"))
	if err == nil {
		return string(data)
	}
	return ""
}

// ─── env.sh (bash/zsh/fish) ───────────────────────────────────────────────────

func writeShFragment(cfg config.ClaudeConfig, path string) error {
	block := fmt.Sprintf("export ANTHROPIC_API_KEY=%q\n", cfg.APIKey)
	if cfg.DefaultModel != "" {
		block += fmt.Sprintf("export ANTHROPIC_MODEL=%q\n", cfg.DefaultModel)
	}
	return mergeIntoFile(path, "# aiswitch:claude", "# /aiswitch:claude", block,
		"# aiswitch env — source this file or add it to your shell profile\n")
}

// ─── env.ps1 (PowerShell) ─────────────────────────────────────────────────────

func writePS1Fragment(cfg config.ClaudeConfig, path string) error {
	block := fmt.Sprintf("$env:ANTHROPIC_API_KEY = %q\n", cfg.APIKey)
	if cfg.DefaultModel != "" {
		block += fmt.Sprintf("$env:ANTHROPIC_MODEL = %q\n", cfg.DefaultModel)
	}
	return mergeIntoFile(path, "# aiswitch:claude", "# /aiswitch:claude", block,
		"# aiswitch env — dot-source this file: . ~/.aiswitch/env.ps1\n")
}

// ─── ~/.anthropic/api_key ─────────────────────────────────────────────────────

func writeAnthropicAPIKey(home, key string) error {
	dir := filepath.Join(home, ".anthropic")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir .anthropic: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "api_key"), []byte(key), 0o600)
}

// ─── ~/.claude/.credentials.json ─────────────────────────────────────────────

type claudeCredentials struct {
	ClaudeAI map[string]claudeAPIEntry `json:"claudeAI,omitempty"`
}

type claudeAPIEntry struct {
	APIKey string `json:"apiKey"`
}

// patchClaudeCredentials updates ~/.claude/.credentials.json if it exists.
// On Windows the path is %APPDATA%\Claude\.credentials.json.
func patchClaudeCredentials(home, apiKey string) error {
	credPath := claudeCredentialsPath(home)

	data, err := os.ReadFile(credPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil // Claude Code not installed; skip.
	}
	if err != nil {
		return fmt.Errorf("reading .credentials.json: %w", err)
	}

	var creds claudeCredentials
	_ = json.Unmarshal(data, &creds)
	if creds.ClaudeAI == nil {
		creds.ClaudeAI = map[string]claudeAPIEntry{}
	}
	creds.ClaudeAI["default"] = claudeAPIEntry{APIKey: apiKey}

	out, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(credPath, out, 0o600)
}

// claudeCredentialsPath returns the OS-appropriate path for the credentials file.
func claudeCredentialsPath(home string) string {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "Claude", ".credentials.json")
		}
	}
	return filepath.Join(home, ".claude", ".credentials.json")
}

// ─── shared env file helpers ──────────────────────────────────────────────────

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// mergeIntoFile replaces the region between startMarker/endMarker in path
// with block, or appends a new region if the markers are absent.
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
