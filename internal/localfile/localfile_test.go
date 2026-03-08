package localfile

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// ─── Read ─────────────────────────────────────────────────────────────────────

func TestRead_PlainText(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, FileName, "work\n")

	cfg, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "work" {
		t.Errorf("Profile = %q, want %q", cfg.Profile, "work")
	}
}

func TestRead_PlainTextWithComment(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, FileName, "# my profile\npersonal\n")

	cfg, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "personal" {
		t.Errorf("Profile = %q, want %q", cfg.Profile, "personal")
	}
}

func TestRead_YAML_ProfileOnly(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, FileName, "profile: work\n")

	cfg, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "work" {
		t.Errorf("Profile = %q, want %q", cfg.Profile, "work")
	}
}

func TestRead_YAML_WithOverrides(t *testing.T) {
	dir := t.TempDir()
	content := `profile: work
claude:
  model: claude-opus-4-5
openai:
  model: gpt-4o
gemini:
  model: gemini-2.0-flash
github:
  email: me@company.com
`
	path := writeFile(t, dir, FileName, content)

	cfg, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "work" {
		t.Errorf("Profile = %q, want %q", cfg.Profile, "work")
	}
	if cfg.Claude == nil || cfg.Claude.Model != "claude-opus-4-5" {
		t.Errorf("Claude.Model = %v, want claude-opus-4-5", cfg.Claude)
	}
	if cfg.OpenAI == nil || cfg.OpenAI.Model != "gpt-4o" {
		t.Errorf("OpenAI.Model = %v, want gpt-4o", cfg.OpenAI)
	}
	if cfg.Gemini == nil || cfg.Gemini.Model != "gemini-2.0-flash" {
		t.Errorf("Gemini.Model = %v, want gemini-2.0-flash", cfg.Gemini)
	}
	if cfg.GitHub == nil || cfg.GitHub.Email != "me@company.com" {
		t.Errorf("GitHub.Email = %v, want me@company.com", cfg.GitHub)
	}
}

func TestRead_YAML_MissingProfile(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, FileName, "claude:\n  model: claude-opus-4-5\n")

	_, err := Read(path)
	if err == nil {
		t.Error("expected error for missing 'profile' field, got nil")
	}
}

func TestRead_YAML_TypoInKey(t *testing.T) {
	dir := t.TempDir()
	// 'profle' is a typo — profile field should be empty → error
	path := writeFile(t, dir, FileName, "profle: work\n")

	_, err := Read(path)
	if err == nil {
		t.Error("expected error for typo'd profile key, got nil")
	}
}

func TestRead_FileNotFound(t *testing.T) {
	_, err := Read("/nonexistent/.aiswitch")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// ─── Find ─────────────────────────────────────────────────────────────────────

func TestFind_InCurrentDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, FileName, "work")

	got := Find(dir)
	if got == "" {
		t.Error("Find() returned empty string, expected a path")
	}
}

func TestFind_InParentDir(t *testing.T) {
	parent := t.TempDir()
	writeFile(t, parent, FileName, "work")
	child := filepath.Join(parent, "subdir")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}

	got := Find(child)
	if got == "" {
		t.Error("Find() returned empty string, expected to find parent .aiswitch")
	}
}

func TestFind_NotFound(t *testing.T) {
	dir := t.TempDir()
	got := Find(dir)
	if got != "" {
		t.Errorf("Find() = %q, want empty string", got)
	}
}

// ─── Write ────────────────────────────────────────────────────────────────────

func TestWrite_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	original := &LocalConfig{
		Profile: "work",
		Claude:  &ClaudeOverride{Model: "claude-opus-4-5"},
	}

	if err := Write(path, original); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	loaded, err := Read(path)
	if err != nil {
		t.Fatalf("Read() after Write() error: %v", err)
	}

	if loaded.Profile != original.Profile {
		t.Errorf("Profile: got %q, want %q", loaded.Profile, original.Profile)
	}
	if loaded.Claude == nil || loaded.Claude.Model != original.Claude.Model {
		t.Errorf("Claude.Model: got %v, want %q", loaded.Claude, original.Claude.Model)
	}
}
