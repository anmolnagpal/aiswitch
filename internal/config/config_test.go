package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// overrideHome redirects os.UserHomeDir to a temp directory for the duration
// of the test. It sets both HOME (Unix) and USERPROFILE (Windows) so that
// os.UserHomeDir() returns the temp dir on all platforms.
func overrideHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir) // Windows
	return dir
}

func TestLoad_EmptyWhenMissing(t *testing.T) {
	overrideHome(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error for missing file: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected empty profiles, got %d", len(cfg.Profiles))
	}
}

func TestSaveAndLoad_Roundtrip(t *testing.T) {
	overrideHome(t)

	original := &Config{
		ActiveProfile: "work",
		Profiles: map[string]Profile{
			"work": {
				Description: "Day job",
				Claude: &ClaudeConfig{
					APIKey:       "sk-ant-test",
					DefaultModel: "claude-opus-4-5",
				},
			},
			"personal": {
				GitHub: &GitHubConfig{
					Token:    "ghp_token",
					Username: "octocat",
					Email:    "me@example.com",
				},
			},
		},
	}

	if err := Save(original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.ActiveProfile != original.ActiveProfile {
		t.Errorf("ActiveProfile: got %q, want %q", loaded.ActiveProfile, original.ActiveProfile)
	}
	if len(loaded.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(loaded.Profiles))
	}

	work, ok := loaded.Profiles["work"]
	if !ok {
		t.Fatal("'work' profile missing after load")
	}
	if work.Claude == nil {
		t.Fatal("work.Claude is nil after load")
	}
	if work.Claude.APIKey != "sk-ant-test" {
		t.Errorf("APIKey: got %q, want %q", work.Claude.APIKey, "sk-ant-test")
	}
	if work.Claude.DefaultModel != "claude-opus-4-5" {
		t.Errorf("DefaultModel: got %q, want %q", work.Claude.DefaultModel, "claude-opus-4-5")
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	home := overrideHome(t)

	dir := filepath.Join(home, configDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte("{bad json"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load() should return an error for corrupt JSON")
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	home := overrideHome(t)

	cfg := &Config{Profiles: map[string]Profile{}}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	dir := filepath.Join(home, configDirName)
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("config directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory")
	}
	// Windows does not enforce Unix-style permission bits.
	if runtime.GOOS != "windows" {
		if info.Mode().Perm() != 0o700 {
			t.Errorf("directory permissions: got %04o, want 0700", info.Mode().Perm())
		}
	}
}

func TestSave_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not enforce Unix-style file permission bits")
	}

	overrideHome(t)

	cfg := &Config{Profiles: map[string]Profile{}}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("file permissions: got %04o, want 0600", info.Mode().Perm())
	}
}

func TestProfile_Services(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		want    string
	}{
		{"empty", Profile{}, "empty"},
		{"claude only", Profile{Claude: &ClaudeConfig{}}, "Claude"},
		{"multiple", Profile{
			Claude: &ClaudeConfig{},
			OpenAI: &OpenAIConfig{},
		}, "Claude + OpenAI"},
		{"with ides", Profile{
			Claude: &ClaudeConfig{},
			IDE:    &IDEConfig{Cursor: true, Windsurf: true},
		}, "Claude + Cursor/Windsurf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.profile.Services()
			if got != tt.want {
				t.Errorf("Services() = %q, want %q", got, tt.want)
			}
		})
	}
}
