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
	return mergeIntoFile(path, "# aiswitch:gemini", "# /aiswitch:gemini", block,
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
	return mergeIntoFile(path, "# aiswitch:gemini", "# /aiswitch:gemini", block,
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
