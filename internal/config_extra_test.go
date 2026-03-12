package internal

import (
	"os"
	"testing"
)

func TestHasShellWrapper_Set(t *testing.T) {
	t.Setenv("RIFF_WRAPPER", "1")
	if !HasShellWrapper() {
		t.Error("HasShellWrapper() = false, want true when RIFF_WRAPPER=1")
	}
}

func TestHasShellWrapper_Unset(t *testing.T) {
	t.Setenv("RIFF_WRAPPER", "")
	if HasShellWrapper() {
		t.Error("HasShellWrapper() = true, want false when RIFF_WRAPPER is empty")
	}
}

func TestHasShellWrapper_WrongValue(t *testing.T) {
	t.Setenv("RIFF_WRAPPER", "yes")
	if HasShellWrapper() {
		t.Error("HasShellWrapper() = true, want false when RIFF_WRAPPER=yes (not '1')")
	}
}

func TestDetectShell_Bash(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	if got := DetectShell(); got != "bash" {
		t.Errorf("DetectShell() = %q, want %q", got, "bash")
	}
}

func TestDetectShell_Zsh(t *testing.T) {
	t.Setenv("SHELL", "/usr/bin/zsh")
	if got := DetectShell(); got != "zsh" {
		t.Errorf("DetectShell() = %q, want %q", got, "zsh")
	}
}

func TestDetectShell_Fish(t *testing.T) {
	t.Setenv("SHELL", "/usr/local/bin/fish")
	if got := DetectShell(); got != "fish" {
		t.Errorf("DetectShell() = %q, want %q", got, "fish")
	}
}

func TestDetectShell_Unknown(t *testing.T) {
	t.Setenv("SHELL", "/bin/tcsh")
	if got := DetectShell(); got != "" {
		t.Errorf("DetectShell() = %q, want empty for unsupported shell", got)
	}
}

func TestDetectShell_Empty(t *testing.T) {
	t.Setenv("SHELL", "")
	if got := DetectShell(); got != "" {
		t.Errorf("DetectShell() = %q, want empty when SHELL is unset", got)
	}
}

func TestConfiguredAIProvider_Empty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RIFF_HOME", dir)
	InitPaths()

	// No config file exists.
	if got := ConfiguredAIProvider(); got != "" {
		t.Errorf("ConfiguredAIProvider() = %q, want empty when no config exists", got)
	}
}

func TestConfiguredAIProvider_Set(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RIFF_HOME", dir)
	InitPaths()

	if err := os.WriteFile(ConfigPath(), []byte(`{"ai_provider":"copilot"}`), 0644); err != nil {
		t.Fatal(err)
	}

	if got := ConfiguredAIProvider(); got != "copilot" {
		t.Errorf("ConfiguredAIProvider() = %q, want %q", got, "copilot")
	}
}
