// Package main_test contains end-to-end integration tests.
//
// These tests compile the real aiswitch binary and exercise the full
// user-facing CLI flows: listing, switching, env-file generation, shell-init,
// and auto-detect.  They run against a temporary home directory so they never
// touch the developer's real config.
//
// Run with:
//
//	go test -v -run Integration ./...
//	go test -v -race -count=1 -run Integration ./...
package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ─── test binary setup ────────────────────────────────────────────────────────

// buildBinary compiles the aiswitch binary into a temp dir once per test run
// and returns the path to the executable. It is called from TestMain so that
// subsequent sub-tests can share the same binary.
var integrationBinary string

func TestMain(m *testing.M) {
	// Build the binary into a temp directory.
	dir, err := os.MkdirTemp("", "aiswitch-integration-*")
	if err != nil {
		panic("failed to create temp dir for binary: " + err.Error())
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "aiswitch")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
	if err != nil {
		panic("failed to build binary: " + err.Error() + "\n" + string(out))
	}
	integrationBinary = bin

	os.Exit(m.Run())
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// harness holds the state for one integration test scenario.
type harness struct {
	t    *testing.T
	home string // temp home directory
}

// newHarness creates an isolated home directory and returns a harness.
func newHarness(t *testing.T) *harness {
	t.Helper()
	home := t.TempDir()
	return &harness{t: t, home: home}
}

// run executes the aiswitch binary with the given arguments in a clean
// environment that redirects HOME to the harness's temp directory.
// It returns stdout+stderr combined and the exit error (nil = exit 0).
func (h *harness) run(args ...string) (string, error) {
	h.t.Helper()
	cmd := exec.Command(integrationBinary, args...)
	cmd.Env = append(os.Environ(),
		"HOME="+h.home,
		"USERPROFILE="+h.home, // Windows
		"NO_COLOR=1",          // disable ANSI so assertions are predictable
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// mustRun calls run and fails the test if the command exits non-zero.
func (h *harness) mustRun(args ...string) string {
	h.t.Helper()
	out, err := h.run(args...)
	if err != nil {
		h.t.Fatalf("aiswitch %v failed:\n%s\nerror: %v", args, out, err)
	}
	return out
}

// mustFail calls run and fails the test if the command exits zero.
func (h *harness) mustFail(args ...string) string {
	h.t.Helper()
	out, err := h.run(args...)
	if err == nil {
		h.t.Fatalf("aiswitch %v should have failed but exited 0:\n%s", args, out)
	}
	return out
}

// seedConfig writes a config.json directly into the harness's home dir,
// bypassing the interactive `add` wizard.
func (h *harness) seedConfig(cfg map[string]interface{}) {
	h.t.Helper()
	dir := filepath.Join(h.home, ".aiswitch")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		h.t.Fatal(err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		h.t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		h.t.Fatal(err)
	}
}

// readFile returns the contents of a file relative to the harness home dir.
func (h *harness) readFile(rel string) string {
	h.t.Helper()
	data, err := os.ReadFile(filepath.Join(h.home, rel))
	if err != nil {
		h.t.Fatalf("readFile(%q): %v", rel, err)
	}
	return string(data)
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected output to contain %q\ngot:\n%s", needle, haystack)
	}
}

func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("expected output NOT to contain %q\ngot:\n%s", needle, haystack)
	}
}

// ─── integration tests ────────────────────────────────────────────────────────

func TestIntegration_Version(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("--version")
	assertContains(t, out, "aiswitch")
}

func TestIntegration_ListEmpty(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("list")
	assertContains(t, out, "No profiles yet")
	assertContains(t, out, "aiswitch add")
}

func TestIntegration_UseNonExistentProfile(t *testing.T) {
	h := newHarness(t)
	out := h.mustFail("use", "doesnotexist")
	assertContains(t, out, "doesnotexist")
	assertContains(t, out, "not found")
}

func TestIntegration_UseProfile_WritesEnvFiles(t *testing.T) {
	h := newHarness(t)
	h.seedConfig(map[string]interface{}{
		"active_profile": "",
		"profiles": map[string]interface{}{
			"work": map[string]interface{}{
				"description": "Work account",
				"claude": map[string]interface{}{
					"api_key":       "sk-ant-work1234567890abcdef",
					"default_model": "claude-opus-4-5",
				},
			},
		},
	})

	out := h.mustRun("use", "work")
	assertContains(t, out, "work")

	// env.sh must exist and contain the API key and model.
	envSh := h.readFile(".aiswitch/env.sh")
	assertContains(t, envSh, "ANTHROPIC_API_KEY")
	assertContains(t, envSh, "sk-ant-work1234567890abcdef")
	assertContains(t, envSh, "ANTHROPIC_MODEL")
	assertContains(t, envSh, "claude-opus-4-5")
	assertContains(t, envSh, "# aiswitch:claude")
	assertContains(t, envSh, "# /aiswitch:claude")

	// env.ps1 is only written on Windows.
	if runtime.GOOS == "windows" {
		envPS1 := h.readFile(".aiswitch/env.ps1")
		assertContains(t, envPS1, "ANTHROPIC_API_KEY")
		assertContains(t, envPS1, "sk-ant-work1234567890abcdef")
	} else {
		if _, err := os.Stat(filepath.Join(h.home, ".aiswitch", "env.ps1")); err == nil {
			t.Error("env.ps1 should not be created on non-Windows")
		}
	}
}

func TestIntegration_UseProfile_UpdatesActiveProfile(t *testing.T) {
	h := newHarness(t)
	h.seedConfig(map[string]interface{}{
		"active_profile": "",
		"profiles": map[string]interface{}{
			"work": map[string]interface{}{
				"claude": map[string]interface{}{
					"api_key": "sk-ant-workkey12345678",
				},
			},
			"personal": map[string]interface{}{
				"claude": map[string]interface{}{
					"api_key": "sk-ant-personalkey5678",
				},
			},
		},
	})

	h.mustRun("use", "work")
	// Switch to a different profile.
	h.mustRun("use", "personal")

	// The personal key should be in env.sh now; work key should be gone.
	envSh := h.readFile(".aiswitch/env.sh")
	assertContains(t, envSh, "sk-ant-personalkey5678")
	assertNotContains(t, envSh, "sk-ant-workkey12345678")
}

func TestIntegration_UseProfile_MultiProvider(t *testing.T) {
	h := newHarness(t)
	h.seedConfig(map[string]interface{}{
		"active_profile": "",
		"profiles": map[string]interface{}{
			"full": map[string]interface{}{
				"claude": map[string]interface{}{
					"api_key": "sk-ant-multitest12345678",
				},
				"openai": map[string]interface{}{
					"api_key": "sk-openai-multitest12345",
					"org_id":  "org-testorg",
				},
				"gemini": map[string]interface{}{
					"api_key": "AIza-gemini-multitest12",
				},
			},
		},
	})

	h.mustRun("use", "full")

	envSh := h.readFile(".aiswitch/env.sh")

	// All three provider blocks must be present.
	assertContains(t, envSh, "ANTHROPIC_API_KEY")
	assertContains(t, envSh, "sk-ant-multitest12345678")
	assertContains(t, envSh, "OPENAI_API_KEY")
	assertContains(t, envSh, "sk-openai-multitest12345")
	assertContains(t, envSh, "OPENAI_ORG_ID")
	assertContains(t, envSh, "org-testorg")
	assertContains(t, envSh, "GEMINI_API_KEY")
	assertContains(t, envSh, "AIza-gemini-multitest12")
	assertContains(t, envSh, "GOOGLE_API_KEY")
}

func TestIntegration_List_ShowsProfiles(t *testing.T) {
	h := newHarness(t)
	h.seedConfig(map[string]interface{}{
		"active_profile": "work",
		"profiles": map[string]interface{}{
			"work": map[string]interface{}{
				"description": "Day job",
				"claude": map[string]interface{}{
					"api_key": "sk-ant-worklisttest1234",
				},
			},
			"personal": map[string]interface{}{
				"description": "Side projects",
			},
		},
	})

	out := h.mustRun("list")
	assertContains(t, out, "work")
	assertContains(t, out, "personal")
	assertContains(t, out, "Day job")
	// API key must be masked, not shown in full.
	assertNotContains(t, out, "sk-ant-worklisttest1234")
}

func TestIntegration_Current_ShowsActiveProfile(t *testing.T) {
	h := newHarness(t)
	h.seedConfig(map[string]interface{}{
		"active_profile": "",
		"profiles": map[string]interface{}{
			"work": map[string]interface{}{
				"description": "My work account",
				"claude": map[string]interface{}{
					"api_key": "sk-ant-currenttest12345",
				},
			},
		},
	})

	// Before switching: no active profile.
	out := h.mustRun("current")
	assertContains(t, out, "No active profile")

	// After switching: should show the profile name.
	h.mustRun("use", "work")
	out = h.mustRun("current")
	assertContains(t, out, "work")
	assertContains(t, out, "My work account")
	// API key must be masked.
	assertNotContains(t, out, "sk-ant-currenttest12345")
}

func TestIntegration_ShellInit_Bash(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("shell-init", "--shell", "bash")
	assertContains(t, out, "aiswitch()")
	assertContains(t, out, "env.sh")
	assertContains(t, out, "_aiswitch_hook")
	assertContains(t, out, "PROMPT_COMMAND")
}

func TestIntegration_ShellInit_Zsh(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("shell-init", "--shell", "zsh")
	assertContains(t, out, "aiswitch()")
	assertContains(t, out, "add-zsh-hook")
	assertContains(t, out, "chpwd")
}

func TestIntegration_ShellInit_Fish(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("shell-init", "--shell", "fish")
	assertContains(t, out, "function aiswitch")
	assertContains(t, out, "on-variable PWD")
}

func TestIntegration_ShellInit_PowerShell(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("shell-init", "--shell", "powershell")
	assertContains(t, out, "Invoke-AiSwitch")
	assertContains(t, out, "Set-Location")
	assertContains(t, out, "env.ps1")
}

func TestIntegration_Detect_NoFile(t *testing.T) {
	h := newHarness(t)
	// In a dir with no .aiswitch, --quiet should produce no output.
	cmd := exec.Command(integrationBinary, "detect", "--quiet")
	cmd.Env = append(os.Environ(),
		"HOME="+h.home,
		"USERPROFILE="+h.home,
		"NO_COLOR=1",
	)
	cmd.Dir = h.home
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("detect --quiet failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("detect --quiet in dir without .aiswitch should produce no output, got: %q", out)
	}
}

func TestIntegration_Detect_WithFile(t *testing.T) {
	h := newHarness(t)
	h.seedConfig(map[string]interface{}{
		"active_profile": "",
		"profiles": map[string]interface{}{
			"work": map[string]interface{}{
				"claude": map[string]interface{}{
					"api_key": "sk-ant-detecttest123456",
				},
			},
		},
	})

	// Create a project directory with a .aiswitch file.
	projectDir := filepath.Join(h.home, "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".aiswitch"), []byte("work\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run detect from inside the project dir.
	cmd := exec.Command(integrationBinary, "detect")
	cmd.Env = append(os.Environ(),
		"HOME="+h.home,
		"USERPROFILE="+h.home,
		"NO_COLOR=1",
	)
	cmd.Dir = projectDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("detect failed: %v\n%s", err, out)
	}

	assertContains(t, string(out), "work")

	// env.sh must have been written.
	envSh := h.readFile(".aiswitch/env.sh")
	assertContains(t, envSh, "sk-ant-detecttest123456")
}

func TestIntegration_Detect_ModelOverride(t *testing.T) {
	h := newHarness(t)
	h.seedConfig(map[string]interface{}{
		"active_profile": "",
		"profiles": map[string]interface{}{
			"work": map[string]interface{}{
				"claude": map[string]interface{}{
					"api_key":       "sk-ant-overridetest1234",
					"default_model": "claude-sonnet-4-5",
				},
			},
		},
	})

	projectDir := filepath.Join(h.home, "proj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// .aiswitch with a per-project model override.
	localFile := "profile: work\nclaude:\n  model: claude-opus-4-5\n"
	if err := os.WriteFile(filepath.Join(projectDir, ".aiswitch"), []byte(localFile), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(integrationBinary, "detect")
	cmd.Env = append(os.Environ(),
		"HOME="+h.home,
		"USERPROFILE="+h.home,
		"NO_COLOR=1",
	)
	cmd.Dir = projectDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("detect failed: %v\n%s", err, out)
	}

	// The overridden model should appear in env.sh.
	envSh := h.readFile(".aiswitch/env.sh")
	assertContains(t, envSh, "claude-opus-4-5")
	assertNotContains(t, envSh, "claude-sonnet-4-5")
}

func TestIntegration_SwitchPreservesOtherProviders(t *testing.T) {
	h := newHarness(t)
	h.seedConfig(map[string]interface{}{
		"active_profile": "",
		"profiles": map[string]interface{}{
			"claude-only": map[string]interface{}{
				"claude": map[string]interface{}{
					"api_key": "sk-ant-preserve12345678",
				},
			},
			"openai-only": map[string]interface{}{
				"openai": map[string]interface{}{
					"api_key": "sk-openai-preserve12345",
				},
			},
		},
	})

	// First switch: claude profile.
	h.mustRun("use", "claude-only")
	envSh := h.readFile(".aiswitch/env.sh")
	assertContains(t, envSh, "sk-ant-preserve12345678")

	// Second switch: openai profile. Claude block must remain (IntoFile preserves other blocks).
	h.mustRun("use", "openai-only")
	envSh = h.readFile(".aiswitch/env.sh")
	assertContains(t, envSh, "sk-openai-preserve12345")
	// Claude block should still be in the file from the previous switch.
	assertContains(t, envSh, "aiswitch:claude")
}

func TestIntegration_SwitchClearsDroppedProviders(t *testing.T) {
	h := newHarness(t)
	h.seedConfig(map[string]interface{}{
		"active_profile": "",
		"profiles": map[string]interface{}{
			"with-openai": map[string]interface{}{
				"claude": map[string]interface{}{
					"api_key": "sk-ant-cleartest12345678",
				},
				"openai": map[string]interface{}{
					"api_key": "sk-openai-cleartest12345",
				},
			},
			"claude-only": map[string]interface{}{
				"claude": map[string]interface{}{
					"api_key": "sk-ant-cleartest12345678",
				},
			},
		},
	})

	// First switch: both providers written.
	h.mustRun("use", "with-openai")
	envSh := h.readFile(".aiswitch/env.sh")
	assertContains(t, envSh, "OPENAI_API_KEY")

	// Second switch: profile has no OpenAI — block must be cleared.
	h.mustRun("use", "claude-only")
	envSh = h.readFile(".aiswitch/env.sh")
	assertNotContains(t, envSh, "sk-openai-cleartest12345")
}

func TestIntegration_InvalidProfileNameRejected(t *testing.T) {
	// The `add` command is interactive so we can't test name validation via CLI
	// without a TTY. Verify instead that `use` with a name containing a space
	// (which can't exist in a valid config) returns a useful error.
	h := newHarness(t)
	out := h.mustFail("use", "bad name")
	assertContains(t, out, "not found")
}
