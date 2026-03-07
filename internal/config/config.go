package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// GetEnvPaths returns both env file paths.
func GetEnvPaths() (EnvPaths, error) {
	sh, err := EnvShellPath()
	if err != nil {
		return EnvPaths{}, err
	}
	ps1, err := EnvPS1Path()
	if err != nil {
		return EnvPaths{}, err
	}
	return EnvPaths{Sh: sh, PS1: ps1}, nil
}

// Load reads the config file. If it does not exist an empty Config is returned.
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
	return &cfg, nil
}

// Save writes the config file, creating the directory if needed.
func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	path := filepath.Join(dir, configFileName)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
