package merge

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIntoFile_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env.sh")

	err := IntoFile(path, "# aiswitch:claude", "# /aiswitch:claude",
		"export ANTHROPIC_API_KEY=\"sk-test\"\n",
		"# header\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)

	assertContains(t, got, "# aiswitch:claude\n")
	assertContains(t, got, "export ANTHROPIC_API_KEY=\"sk-test\"\n")
	assertContains(t, got, "# /aiswitch:claude\n")
	assertContains(t, got, "# header\n")
}

func TestIntoFile_UpdatesExistingBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env.sh")

	// Write initial content.
	initial := "# header\n# aiswitch:claude\nexport ANTHROPIC_API_KEY=\"old-key\"\n# /aiswitch:claude\n"
	if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}

	err := IntoFile(path, "# aiswitch:claude", "# /aiswitch:claude",
		"export ANTHROPIC_API_KEY=\"new-key\"\n", "# header\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)

	assertContains(t, got, "new-key")
	assertNotContains(t, got, "old-key")
	// Header should not be duplicated.
	if count(got, "# header") > 1 {
		t.Errorf("header was duplicated:\n%s", got)
	}
}

func TestIntoFile_AppendsWhenNoMarkers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env.sh")

	// Existing content without any markers.
	if err := os.WriteFile(path, []byte("export EXISTING=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := IntoFile(path, "# aiswitch:openai", "# /aiswitch:openai",
		"export OPENAI_API_KEY=\"sk-openai\"\n", "# header\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)

	assertContains(t, got, "export EXISTING=1")
	assertContains(t, got, "export OPENAI_API_KEY=\"sk-openai\"")
}

func TestIntoFile_MultipleProviderBlocks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env.sh")

	// Write two provider blocks.
	if err := IntoFile(path, "# aiswitch:claude", "# /aiswitch:claude",
		"export ANTHROPIC_API_KEY=\"claude-key\"\n", "# header\n"); err != nil {
		t.Fatal(err)
	}
	if err := IntoFile(path, "# aiswitch:openai", "# /aiswitch:openai",
		"export OPENAI_API_KEY=\"openai-key\"\n", "# header\n"); err != nil {
		t.Fatal(err)
	}

	// Update only the Claude block.
	if err := IntoFile(path, "# aiswitch:claude", "# /aiswitch:claude",
		"export ANTHROPIC_API_KEY=\"claude-key-v2\"\n", "# header\n"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)

	assertContains(t, got, "claude-key-v2")
	assertNotContains(t, got, "claude-key\"\n")
	assertContains(t, got, "openai-key")
}

func TestIntoFile_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not enforce Unix-style file permission bits")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "env.sh")

	if err := IntoFile(path, "# s", "# /s", "block\n", ""); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected permissions 0600, got %04o", info.Mode().Perm())
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func assertContains(t *testing.T, s, sub string) {
	t.Helper()
	if !contains(s, sub) {
		t.Errorf("expected to contain %q, got:\n%s", sub, s)
	}
}

func assertNotContains(t *testing.T, s, sub string) {
	t.Helper()
	if contains(s, sub) {
		t.Errorf("expected NOT to contain %q, got:\n%s", sub, s)
	}
}

func contains(s, sub string) bool {
	return indexOf(s, sub) != -1
}

func count(s, sub string) int {
	n := 0
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			n++
			i += len(sub) - 1
		}
	}
	return n
}
