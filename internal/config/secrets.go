package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// secretsFileName is the filename for credentials, stored inside ~/.aiswitch/.
const secretsFileName = "secrets.json"

// ProfileSecrets holds only the sensitive credential values for one profile.
type ProfileSecrets struct {
	ClaudeAPIKey string `json:"claude_api_key,omitempty"`
	OpenAIAPIKey string `json:"openai_api_key,omitempty"`
	GeminiAPIKey string `json:"gemini_api_key,omitempty"`
	GitHubToken  string `json:"github_token,omitempty"`
}

// secrets maps profile names to their secrets.
type secrets map[string]ProfileSecrets

// secretsPath returns ~/.aiswitch/secrets.json.
func secretsPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, secretsFileName), nil
}

// loadSecrets reads ~/.aiswitch/secrets.json.
// Returns an empty map if the file does not exist.
func loadSecrets() (secrets, error) {
	path, err := secretsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return secrets{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading secrets: %w", err)
	}
	var s secrets
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing secrets: %w", err)
	}
	return s, nil
}

// saveSecrets writes s to ~/.aiswitch/secrets.json with 0o600 permissions.
func saveSecrets(s secrets) error {
	path, err := secretsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding secrets: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing secrets: %w", err)
	}
	return nil
}

// extractSecrets pulls sensitive values out of cfg's profiles into a secrets
// map, zeroing the in-place fields so they are not written to config.json.
func extractSecrets(cfg *Config) secrets {
	s := secrets{}
	for name, p := range cfg.Profiles {
		ps := ProfileSecrets{}
		if p.Claude != nil {
			ps.ClaudeAPIKey = p.Claude.APIKey
			p.Claude.APIKey = ""
		}
		if p.OpenAI != nil {
			ps.OpenAIAPIKey = p.OpenAI.APIKey
			p.OpenAI.APIKey = ""
		}
		if p.Gemini != nil {
			ps.GeminiAPIKey = p.Gemini.APIKey
			p.Gemini.APIKey = ""
		}
		if p.GitHub != nil {
			ps.GitHubToken = p.GitHub.Token
			p.GitHub.Token = ""
		}
		cfg.Profiles[name] = p
		if ps != (ProfileSecrets{}) {
			s[name] = ps
		}
	}
	return s
}

// mergeSecrets writes the values from s back into cfg's profiles.
// Values from s take precedence over any keys already in cfg (e.g. from a
// legacy config.json that still contains plaintext api_key fields).
func mergeSecrets(cfg *Config, s secrets) {
	for name, ps := range s {
		p, ok := cfg.Profiles[name]
		if !ok {
			continue
		}
		if ps.ClaudeAPIKey != "" && p.Claude != nil {
			p.Claude.APIKey = ps.ClaudeAPIKey
		}
		if ps.OpenAIAPIKey != "" && p.OpenAI != nil {
			p.OpenAI.APIKey = ps.OpenAIAPIKey
		}
		if ps.GeminiAPIKey != "" && p.Gemini != nil {
			p.Gemini.APIKey = ps.GeminiAPIKey
		}
		if ps.GitHubToken != "" && p.GitHub != nil {
			p.GitHub.Token = ps.GitHubToken
		}
		cfg.Profiles[name] = p
	}
}
