// Package gemini applies Google Gemini / Google AI profile settings.
//
// Two flavours are supported:
//   - Google AI Studio  — uses GEMINI_API_KEY / GOOGLE_API_KEY
//   - Google Vertex AI  — uses GOOGLE_CLOUD_PROJECT + Application Default Credentials
//     (ADC). When a project_id is set, GOOGLE_CLOUD_PROJECT is also exported.
//
// Strategy:
//  1. Write env.sh and env.ps1.
//  2. Write ~/.config/gemini/api_key (for tools that look there).
package gemini

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/anmolnagpal/aiswitch/internal/config"
	"github.com/anmolnagpal/aiswitch/internal/providers/merge"
)

// Apply writes all Gemini-related credential files for the given config.
func Apply(cfg config.GeminiConfig, paths config.EnvPaths) error {
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
	// Check both common env var names.
	for _, envKey := range []string{"GEMINI_API_KEY", "GOOGLE_API_KEY"} {
		if key := os.Getenv(envKey); key != "" {
			return key
		}
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

func writeShFragment(cfg config.GeminiConfig, path string) error {
	// Export both names for maximum SDK compatibility.
	block := fmt.Sprintf("export GEMINI_API_KEY=%q\n", cfg.APIKey)
	block += fmt.Sprintf("export GOOGLE_API_KEY=%q\n", cfg.APIKey)
	if cfg.DefaultModel != "" {
		block += fmt.Sprintf("export GEMINI_MODEL=%q\n", cfg.DefaultModel)
	}
	if cfg.ProjectID != "" {
		block += fmt.Sprintf("export GOOGLE_CLOUD_PROJECT=%q\n", cfg.ProjectID)
	}
	return merge.IntoFile(path, "# aiswitch:gemini", "# /aiswitch:gemini", block,
		"# aiswitch env — source this file or add it to your shell profile\n")
}

// ─── env.ps1 ──────────────────────────────────────────────────────────────────

func writePS1Fragment(cfg config.GeminiConfig, path string) error {
	block := fmt.Sprintf("$env:GEMINI_API_KEY = %q\n", cfg.APIKey)
	block += fmt.Sprintf("$env:GOOGLE_API_KEY = %q\n", cfg.APIKey)
	if cfg.DefaultModel != "" {
		block += fmt.Sprintf("$env:GEMINI_MODEL = %q\n", cfg.DefaultModel)
	}
	if cfg.ProjectID != "" {
		block += fmt.Sprintf("$env:GOOGLE_CLOUD_PROJECT = %q\n", cfg.ProjectID)
	}
	return merge.IntoFile(path, "# aiswitch:gemini", "# /aiswitch:gemini", block,
		"# aiswitch env — dot-source this file: . ~/.aiswitch/env.ps1\n")
}

// ─── ~/.config/gemini/api_key ─────────────────────────────────────────────────

func configFilePath() (string, error) {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("%%APPDATA%% not set")
		}
		return filepath.Join(appData, "Google", "Gemini", "api_key"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gemini", "api_key"), nil
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
