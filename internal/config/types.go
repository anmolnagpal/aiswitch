package config

import "strings"

// Config is the root configuration stored at ~/.aiswitch/config.json.
type Config struct {
	ActiveProfile string             `json:"active_profile"`
	Profiles      map[string]Profile `json:"profiles"`
}

// Profile holds configuration for a single named profile.
type Profile struct {
	Description string        `json:"description,omitempty"`
	Claude      *ClaudeConfig `json:"claude,omitempty"`
	OpenAI      *OpenAIConfig `json:"openai,omitempty"`
	Gemini      *GeminiConfig `json:"gemini,omitempty"`
	GitHub      *GitHubConfig `json:"github,omitempty"`
	IDE         *IDEConfig    `json:"ide,omitempty"`
}

// ClaudeConfig holds Anthropic / Claude credentials.
type ClaudeConfig struct {
	APIKey       string `json:"api_key"`
	DefaultModel string `json:"default_model,omitempty"`
}

// GitHubConfig holds GitHub credentials used by Copilot and the gh CLI.
type GitHubConfig struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

// IDEConfig controls which IDEs aiswitch patches with API keys on profile switch.
// Each enabled IDE has its settings.json updated with the active provider keys
// so inline AI features (Claude, Copilot, GPT-4o…) automatically use the right account.
type IDEConfig struct {
	// Cursor patches ~/Library/Application Support/Cursor/User/settings.json
	// (macOS) or the platform-equivalent path on Linux/Windows.
	Cursor bool `json:"cursor,omitempty"`
	// Windsurf patches ~/Library/Application Support/Windsurf/User/settings.json.
	Windsurf bool `json:"windsurf,omitempty"`
}

// OpenAIConfig holds OpenAI credentials.
type OpenAIConfig struct {
	APIKey string `json:"api_key"`
	// OrgID is the OpenAI organisation ID (optional).
	OrgID        string `json:"org_id,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
}

// GeminiConfig holds Google Gemini / Google AI credentials.
type GeminiConfig struct {
	APIKey string `json:"api_key"`
	// ProjectID is used for Vertex AI (optional; leave blank for AI Studio).
	ProjectID    string `json:"project_id,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
}

// Services returns a human-readable summary of which services this profile configures.
func (p Profile) Services() string {
	var parts []string
	if p.Claude != nil {
		parts = append(parts, "Claude")
	}
	if p.OpenAI != nil {
		parts = append(parts, "OpenAI")
	}
	if p.Gemini != nil {
		parts = append(parts, "Gemini")
	}
	if p.GitHub != nil {
		parts = append(parts, "GitHub")
	}
	if p.IDE != nil {
		var ides []string
		if p.IDE.Cursor {
			ides = append(ides, "Cursor")
		}
		if p.IDE.Windsurf {
			ides = append(ides, "Windsurf")
		}
		if len(ides) > 0 {
			parts = append(parts, strings.Join(ides, "/"))
		}
	}
	if len(parts) == 0 {
		return "empty"
	}
	return strings.Join(parts, " + ")
}
