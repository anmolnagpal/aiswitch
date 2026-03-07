package config

// Config is the root configuration stored at ~/.aiswitch/config.json.
type Config struct {
	ActiveProfile string             `json:"active_profile"`
	Profiles      map[string]Profile `json:"profiles"`
}

// Profile holds configuration for a single named profile.
type Profile struct {
	Description string        `json:"description,omitempty"`
	Claude      *ClaudeConfig `json:"claude,omitempty"`
	GitHub      *GitHubConfig `json:"github,omitempty"`
}

// ClaudeConfig holds Anthropic / Claude credentials.
type ClaudeConfig struct {
	APIKey string `json:"api_key"`
	// Optional – overrides ANTHROPIC_MODEL when set.
	DefaultModel string `json:"default_model,omitempty"`
}

// GitHubConfig holds GitHub credentials used by Copilot and the gh CLI.
type GitHubConfig struct {
	// Personal access token or OAuth token with the right scopes.
	Token    string `json:"token"`
	Username string `json:"username"`
	// Email wired into local git config when switching.
	Email string `json:"email,omitempty"`
}

// Services returns a human-readable summary of which services this profile configures.
func (p Profile) Services() string {
	switch {
	case p.Claude != nil && p.GitHub != nil:
		return "Claude + GitHub"
	case p.Claude != nil:
		return "Claude only"
	case p.GitHub != nil:
		return "GitHub only"
	default:
		return "empty"
	}
}
