package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	RiffDir     string
	ProjectsDir string
	CdPathFile  string
)

func init() {
	InitPaths()
}

// InitPaths sets RiffDir, ProjectsDir and CdPathFile based on the RIFF_HOME
// environment variable. When RIFF_HOME is unset it defaults to ~/.riff.
// It is called automatically at init time and can be called again in tests
// after changing the environment.
func InitPaths() {
	dir := os.Getenv("RIFF_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		dir = filepath.Join(home, ".riff")
	}
	RiffDir = dir
	ProjectsDir = filepath.Join(RiffDir, "projects")
	CdPathFile = filepath.Join(RiffDir, ".cd-path")
}

// HasShellWrapper reports whether the shell wrapper is active.
// The wrapper sets RIFF_WRAPPER=1 before invoking the binary.
func HasShellWrapper() bool {
	return os.Getenv("RIFF_WRAPPER") == "1"
}

const MetadataFile = ".riff.json"

// Template defines a built-in or user-configured project template.
type Template struct {
	Command string `json:"command"`
}

// BuiltinTemplates are always available, compiled into the binary.
var BuiltinTemplates = map[string]Template{
	"bun":    {Command: "bun init -y"},
	"react":  {Command: "pnpx create-vite . --template react-ts"},
	"python": {Command: "uv init"},
	"rust":   {Command: "cargo init ."},
	"node":   {Command: "npm init -y"},
	"dotnet": {Command: "dotnet new console"},
	"go":     {Command: "go mod init temp"},
	"next":   {Command: "pnpx create-next-app . --ts --eslint --app"},
}

// UserConfig represents the optional ~/.riff/config.json file.
type UserConfig struct {
	Schema     string              `json:"$schema,omitempty"`
	Templates  map[string]Template `json:"templates,omitempty"`
	AIProvider string              `json:"ai_provider,omitempty"`
}

// ConfigPath returns the path to the config file (~/.riff/config.json).
func ConfigPath() string {
	return filepath.Join(RiffDir, "config.json")
}

// ConfiguredAIProvider returns the AI provider set in ~/.riff/config.json,
// or an empty string when no preference is configured.
func ConfiguredAIProvider() string {
	cfg := LoadConfig()
	return cfg.AIProvider
}

// LoadConfig reads ~/.riff/config.json if it exists, returns empty config otherwise.
func LoadConfig() UserConfig {
	configPath := filepath.Join(RiffDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return UserConfig{}
	}

	var cfg UserConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return UserConfig{}
	}
	return cfg
}

// DetectShell returns the shell name from the SHELL env var (e.g. "bash", "zsh", "fish").
// Returns empty string if SHELL is unset or unrecognised.
func DetectShell() string {
	shell := filepath.Base(os.Getenv("SHELL"))
	switch shell {
	case "bash", "zsh", "fish":
		return shell
	}
	return ""
}

// EnsureShellWrapper checks whether the shell wrapper is active and, if not,
// appends the appropriate `riff init` line to the user's shell config file.
func EnsureShellWrapper(projectPath string) {
	if HasShellWrapper() {
		return
	}

	shell := DetectShell()
	if shell == "" {
		fmt.Printf("\n  %s %s\n",
			Dim("To auto-cd into projects, add to your shell config:"),
			Cyan("eval \"$(riff init)\""),
		)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	var configFile string
	var initLine string
	switch shell {
	case "bash":
		configFile = filepath.Join(home, ".bashrc")
		initLine = `eval "$(riff init)"`
	case "zsh":
		configFile = filepath.Join(home, ".zshrc")
		initLine = `eval "$(riff init)"`
	case "fish":
		configFile = filepath.Join(home, ".config", "fish", "config.fish")
		initLine = "riff init fish | source"
	}

	data, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	relPath := "~/" + strings.TrimPrefix(configFile, home+"/")

	if strings.Contains(string(data), "riff init") {
		// Already installed but not active in current session.
		fmt.Printf("\n  %s\n",
			Dim("To activate auto-cd, run: "+Cyan("source "+relPath+" && cd "+projectPath)),
		)
		return
	}

	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	line := "\n" + initLine + "\n"
	if _, err := f.WriteString(line); err != nil {
		return
	}

	fmt.Printf("\n  %s Added %s to %s\n",
		Green("+"),
		Cyan(initLine),
		Bold(relPath),
	)
	fmt.Printf("  %s\n",
		Dim("To activate now, run: "+Cyan("source "+relPath+" && cd "+projectPath)),
	)
}

// GetTemplates returns built-in templates merged with user overrides.
func GetTemplates() map[string]Template {
	templates := make(map[string]Template)
	for k, v := range BuiltinTemplates {
		templates[k] = v
	}

	cfg := LoadConfig()
	for k, v := range cfg.Templates {
		templates[k] = v
	}

	return templates
}
