package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	configDirName  = ".aiswitch"
	configFileName = "config.json"
	// EnvShellFile is sourced by bash/zsh/fish shell wrappers.
	EnvShellFile = "env.sh"
	// EnvPS1File is dot-sourced by PowerShell wrappers on Windows.
	EnvPS1File = "env.ps1"
)

// EnvPaths holds the paths to both shell env files.
type EnvPaths struct {
	Sh  string // bash / zsh / fish
	PS1 string // PowerShell
}

// Dir returns the path to ~/.aiswitch.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, configDirName), nil
}

// Path returns the path to ~/.aiswitch/config.json.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// EnvShellPath returns the path to ~/.aiswitch/env.sh.
func EnvShellPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, EnvShellFile), nil
}

// EnvPS1Path returns the path to ~/.aiswitch/env.ps1.
func EnvPS1Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, EnvPS1File), nil
}

// GetEnvPaths returns the env file paths for the current OS.
// PS1 is only populated on Windows; it is empty on macOS and Linux.
func GetEnvPaths() (EnvPaths, error) {
	sh, err := EnvShellPath()
	if err != nil {
		return EnvPaths{}, err
	}
	var ps1 string
	if runtime.GOOS == "windows" {
		ps1, err = EnvPS1Path()
		if err != nil {
			return EnvPaths{}, err
		}
	}
	return EnvPaths{Sh: sh, PS1: ps1}, nil
}

// Load reads the config and secrets files.
// If config.json does not exist an empty Config is returned.
// Secrets from secrets.json are merged in (and take precedence over any
// legacy plaintext api_key values still present in config.json).
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{Profiles: map[string]Profile{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}

	// Merge secrets from the separate secrets file.
	s, err := loadSecrets()
	if err != nil {
		return nil, err
	}
	mergeSecrets(&cfg, s)

	return &cfg, nil
}

// Save writes config.json (no secrets) and secrets.json (credentials only).
// Any existing plaintext api_key/token values in cfg are automatically
// migrated to secrets.json and removed from config.json.
func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	// Extract secrets before marshalling — this zeroes the key fields in cfg.
	s := extractSecrets(cfg)
	if err := saveSecrets(s); err != nil {
		return err
	}

	// cfg.Profiles now has empty api_key/token fields — safe to write.
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	path := filepath.Join(dir, configFileName)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	// Restore the zeroed fields so the in-memory cfg remains usable after Save.
	mergeSecrets(cfg, s)

	return nil
}
