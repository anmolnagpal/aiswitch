// Package localfile handles reading and writing the per-project .aiswitch file.
//
// The .aiswitch file works exactly like .terraform-version or .nvmrc: you
// commit it to your repo so that everyone (and every shell session) uses the
// correct AI account when working in that directory.
//
// File format (YAML):
//
//	# .aiswitch
//	profile: work
//
//	claude:
//	  model: claude-opus-4-5        # optional — overrides profile default
//
//	openai:
//	  model: gpt-4o                 # optional — overrides profile default
//
//	gemini:
//	  model: gemini-2.0-flash       # optional — overrides profile default
//
//	github:
//	  email: me@company.com         # optional — overrides git commit email
//
// Plain-text shorthand (just a profile name) is also supported:
//
//	work
package localfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileName is the name of the project-level config file.
const FileName = ".aiswitch"

// LocalConfig is the structure of a project-level .aiswitch file.
type LocalConfig struct {
	// Profile is the global aiswitch profile to activate in this directory.
	Profile string `yaml:"profile"`

	// Per-provider overrides — all fields are optional.
	Claude *ClaudeOverride `yaml:"claude,omitempty"`
	OpenAI *OpenAIOverride `yaml:"openai,omitempty"`
	Gemini *GeminiOverride `yaml:"gemini,omitempty"`
	GitHub *GitHubOverride `yaml:"github,omitempty"`
}

// ClaudeOverride lets a project pin a specific Claude model.
type ClaudeOverride struct {
	Model string `yaml:"model,omitempty"`
}

// OpenAIOverride lets a project pin a specific OpenAI model.
type OpenAIOverride struct {
	Model string `yaml:"model,omitempty"`
}

// GeminiOverride lets a project pin a specific Gemini model.
type GeminiOverride struct {
	Model string `yaml:"model,omitempty"`
}

// GitHubOverride lets a project override the git commit email.
type GitHubOverride struct {
	Email string `yaml:"email,omitempty"`
}

// Find walks up from dir (inclusive) looking for a .aiswitch file.
// Returns the full path if found, empty string if not found anywhere.
func Find(dir string) string {
	for {
		path := filepath.Join(dir, FileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "" // reached filesystem root
		}
		dir = parent
	}
}

// Read parses a .aiswitch file at path and returns its config.
//
// Supports two formats:
//  1. Plain text — a single line containing only the profile name.
//  2. YAML — full structured format with optional overrides.
func Read(path string) (*LocalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(string(data))

	// Strip comment lines for the plain-text detection heuristic.
	nonComment := ""
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			nonComment = trimmed
			break
		}
	}

	// Plain-text shorthand: single token with no YAML keys.
	if nonComment != "" &&
		!strings.Contains(nonComment, ":") &&
		!strings.Contains(nonComment, " ") {
		return &LocalConfig{Profile: nonComment}, nil
	}

	var cfg LocalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if cfg.Profile == "" {
		return nil, fmt.Errorf("%s: 'profile' field is required", path)
	}
	return &cfg, nil
}

// Write serialises cfg to path with a helpful header comment.
func Write(path string, cfg *LocalConfig) error {
	body, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	header := "# aiswitch project config — safe to commit, contains no secrets\n" +
		"# Run `aiswitch init` to update  |  Run `aiswitch detect` to apply now\n"
	return os.WriteFile(path, append([]byte(header), body...), 0o644)
}
