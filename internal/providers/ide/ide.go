// Package ide patches the settings.json of Cursor and Windsurf with the API
// keys from the active profile, so that built-in AI features (Claude inline
// edit, Copilot, etc.) automatically use the correct account after a profile
// switch — without any manual IDE restart.
//
// Settings file locations by platform:
//
//	macOS   ~/Library/Application Support/{IDE}/User/settings.json
//	Linux   ~/.config/{IDE}/User/settings.json
//	Windows %APPDATA%\{IDE}\User\settings.json
//
// The file is patched in-place: all existing keys are preserved; only the
// AI-provider keys managed by aiswitch are added or updated.
package ide

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/anmolnagpal/aiswitch/internal/config"
	"github.com/anmolnagpal/aiswitch/internal/ui"
)

// Apply patches each enabled IDE's settings.json with the API keys held in
// profile. It is a no-op for any IDE whose flag is false.
func Apply(cfg config.IDEConfig, profile config.Profile) error {
	if cfg.Cursor {
		if err := applyIDE("Cursor", profile); err != nil {
			// Non-fatal: IDE may not be installed.
			fmt.Fprintf(os.Stderr, "aiswitch: warning: Cursor settings: %v\n", err)
		}
	}
	if cfg.Windsurf {
		if err := applyIDE("Windsurf", profile); err != nil {
			fmt.Fprintf(os.Stderr, "aiswitch: warning: Windsurf settings: %v\n", err)
		}
	}
	return nil
}

// InstalledIDEs returns the names of IDEs that appear to be installed on this
// machine (i.e. their settings directory exists). Useful for the `add` wizard.
func InstalledIDEs() []string {
	candidates := []string{"Cursor", "Windsurf"}
	var found []string
	for _, name := range candidates {
		p, err := settingsPath(name)
		if err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Dir(p)); err == nil {
			found = append(found, name)
		}
	}
	return found
}

// DetectIDE returns a short status string for the named IDE:
// the value of "anthropic.apiKey" (masked) if aiswitch has patched the file,
// or an empty string if the IDE is not installed / not yet patched.
func DetectIDE(ideName string) string {
	p, err := settingsPath(ideName)
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return ""
	}
	// Check for any key we manage as evidence this IDE has been patched.
	for _, key := range []string{"anthropic.apiKey", "openai.apiKey", "googleGenerativeAI.apiKey"} {
		if v, ok := settings[key].(string); ok && v != "" {
			return ui.MaskSecret(v)
		}
	}
	return ""
}

// ─── internal ─────────────────────────────────────────────────────────────────

func applyIDE(name string, profile config.Profile) error {
	path, err := settingsPath(name)
	if err != nil {
		return err
	}

	// Load existing settings (ignore missing file — we will create it).
	settings := map[string]interface{}{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &settings)
	}

	patch(settings, profile)

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating settings directory: %w", err)
	}
	data, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

// patch writes provider API-key settings into a VS Code–style JSON map.
//
// Key names follow the conventions used by popular VS Code extensions:
//
//	Claude  → anthropic.apiKey, anthropic.defaultModel
//	OpenAI  → openai.apiKey, openai.organization, openai.defaultModel
//	Gemini  → googleGenerativeAI.apiKey, googleGenerativeAI.defaultModel
func patch(settings map[string]interface{}, profile config.Profile) {
	if profile.Claude != nil {
		settings["anthropic.apiKey"] = profile.Claude.APIKey
		if profile.Claude.DefaultModel != "" {
			settings["anthropic.defaultModel"] = profile.Claude.DefaultModel
		}
	}
	if profile.OpenAI != nil {
		settings["openai.apiKey"] = profile.OpenAI.APIKey
		if profile.OpenAI.OrgID != "" {
			settings["openai.organization"] = profile.OpenAI.OrgID
		}
		if profile.OpenAI.DefaultModel != "" {
			settings["openai.defaultModel"] = profile.OpenAI.DefaultModel
		}
	}
	if profile.Gemini != nil {
		settings["googleGenerativeAI.apiKey"] = profile.Gemini.APIKey
		if profile.Gemini.DefaultModel != "" {
			settings["googleGenerativeAI.defaultModel"] = profile.Gemini.DefaultModel
		}
	}
}

func settingsPath(ideName string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", ideName, "User", "settings.json"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("%%APPDATA%% not set")
		}
		return filepath.Join(appData, ideName, "User", "settings.json"), nil
	default: // Linux and BSD
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config", ideName, "User", "settings.json"), nil
	}
}
