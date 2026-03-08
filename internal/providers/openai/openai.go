// Package openai applies OpenAI profile settings to the local system.
//
// Strategy:
//  1. Write env.sh and env.ps1 — sets OPENAI_API_KEY (+ OPENAI_ORG_ID and
//     OPENAI_MODEL when configured).
//  2. Write ~/.config/openai/api_key — read by the openai CLI tool.
package openai

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/anmolnagpal/aiswitch/internal/config"
)

// Apply writes all OpenAI-related credential files for the given config.
func Apply(cfg config.OpenAIConfig, paths config.EnvPaths) error {
	if err := writeShFragment(cfg, paths.Sh); err != nil {
		return err
	}
	if err := writePS1Fragment(cfg, paths.PS1); err != nil {
		return err
	}
	return writeConfigFile(cfg.APIKey)
}

// Detect returns the API key currently in use, or an empty string if unknown.
func Detect() string {
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return key
	}
	p, err := configFilePath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(data)
}

// ─── env.sh ───────────────────────────────────────────────────────────────────

func writeShFragment(cfg config.OpenAIConfig, path string) error {
	block := fmt.Sprintf("export OPENAI_API_KEY=%q\n", cfg.APIKey)
	if cfg.OrgID != "" {
		block += fmt.Sprintf("export OPENAI_ORG_ID=%q\n", cfg.OrgID)
	}
	if cfg.DefaultModel != "" {
		block += fmt.Sprintf("export OPENAI_MODEL=%q\n", cfg.DefaultModel)
	}
	return mergeIntoFile(path, "# aiswitch:openai", "# /aiswitch:openai", block,
		"# aiswitch env — source this file or add it to your shell profile\n")
}

// ─── env.ps1 ──────────────────────────────────────────────────────────────────

func writePS1Fragment(cfg config.OpenAIConfig, path string) error {
	block := fmt.Sprintf("$env:OPENAI_API_KEY = %q\n", cfg.APIKey)
	if cfg.OrgID != "" {
		block += fmt.Sprintf("$env:OPENAI_ORG_ID = %q\n", cfg.OrgID)
	}
	if cfg.DefaultModel != "" {
		block += fmt.Sprintf("$env:OPENAI_MODEL = %q\n", cfg.DefaultModel)
	}
	return mergeIntoFile(path, "# aiswitch:openai", "# /aiswitch:openai", block,
		"# aiswitch env — dot-source this file: . ~/.aiswitch/env.ps1\n")
}

// ─── ~/.config/openai/api_key ─────────────────────────────────────────────────

func configFilePath() (string, error) {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("%%APPDATA%% not set")
		}
		return filepath.Join(appData, "OpenAI", "api_key"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "openai", "api_key"), nil
}

func writeConfigFile(apiKey string) error {
	p, err := configFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(apiKey), 0o600)
}

// ─── shared helpers ───────────────────────────────────────────────────────────

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
