package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitPaths_Default(t *testing.T) {
	t.Setenv("RIFF_HOME", "")
	InitPaths()

	if !strings.HasSuffix(RiffDir, ".riff") {
		t.Errorf("RiffDir = %q, want suffix .riff", RiffDir)
	}
	if !strings.HasSuffix(ProjectsDir, filepath.Join(".riff", "projects")) {
		t.Errorf("ProjectsDir = %q, want suffix .riff/projects", ProjectsDir)
	}
	if !strings.HasSuffix(CdPathFile, filepath.Join(".riff", ".cd-path")) {
		t.Errorf("CdPathFile = %q, want suffix .riff/.cd-path", CdPathFile)
	}
}

func TestInitPaths_RIFF_HOME(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RIFF_HOME", dir)
	InitPaths()

	if RiffDir != dir {
		t.Errorf("RiffDir = %q, want %q", RiffDir, dir)
	}
	if ProjectsDir != filepath.Join(dir, "projects") {
		t.Errorf("ProjectsDir = %q, want %q", ProjectsDir, filepath.Join(dir, "projects"))
	}
	if CdPathFile != filepath.Join(dir, ".cd-path") {
		t.Errorf("CdPathFile = %q, want %q", CdPathFile, filepath.Join(dir, ".cd-path"))
	}
}

func TestConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RIFF_HOME", dir)
	InitPaths()

	got := ConfigPath()
	want := filepath.Join(dir, "config.json")
	if got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RIFF_HOME", dir)
	InitPaths()

	cfg := LoadConfig()
	if cfg.AIProvider != "" {
		t.Errorf("AIProvider = %q, want empty", cfg.AIProvider)
	}
	if len(cfg.Templates) != 0 {
		t.Errorf("Templates = %v, want empty", cfg.Templates)
	}
}

func TestLoadConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RIFF_HOME", dir)
	InitPaths()

	cfg := UserConfig{
		AIProvider: "claude",
		Templates: map[string]Template{
			"custom": {Command: "echo hello"},
		},
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	got := LoadConfig()
	if got.AIProvider != "claude" {
		t.Errorf("AIProvider = %q, want %q", got.AIProvider, "claude")
	}
	if tmpl, ok := got.Templates["custom"]; !ok || tmpl.Command != "echo hello" {
		t.Errorf("Templates[custom] = %v, want command 'echo hello'", got.Templates["custom"])
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RIFF_HOME", dir)
	InitPaths()

	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{broken"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := LoadConfig()
	if cfg.AIProvider != "" {
		t.Errorf("AIProvider = %q, want empty for invalid JSON", cfg.AIProvider)
	}
}

func TestGetTemplates_BuiltinsOnly(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RIFF_HOME", dir)
	InitPaths()

	templates := GetTemplates()

	// Should have all builtins.
	for name := range BuiltinTemplates {
		if _, ok := templates[name]; !ok {
			t.Errorf("missing builtin template %q", name)
		}
	}
}

func TestGetTemplates_UserOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RIFF_HOME", dir)
	InitPaths()

	// Write a config that overrides "go" and adds "custom".
	cfg := UserConfig{
		Templates: map[string]Template{
			"go":     {Command: "go mod init mymod"},
			"custom": {Command: "echo custom"},
		},
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	templates := GetTemplates()

	if templates["go"].Command != "go mod init mymod" {
		t.Errorf("go template = %q, want user override", templates["go"].Command)
	}
	if templates["custom"].Command != "echo custom" {
		t.Errorf("custom template = %q, want 'echo custom'", templates["custom"].Command)
	}
	// Builtins that weren't overridden should still be there.
	if _, ok := templates["bun"]; !ok {
		t.Error("missing builtin template 'bun'")
	}
}
