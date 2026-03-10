package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var (
	RiffDir     string
	ProjectsDir string
	CdPathFile  string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	RiffDir = filepath.Join(home, ".riff")
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
	"react":  {Command: "bunx create-vite . --template react-ts"},
	"python": {Command: "uv init"},
	"rust":   {Command: "cargo init ."},
	"node":   {Command: "npm init -y"},
	"dotnet": {Command: "dotnet new console"},
	"go":     {Command: "go mod init temp"},
	"next":   {Command: "bunx create-next-app . --ts --eslint --app --use-bun"},
}

// UserConfig represents the optional ~/.riff/config.json file.
type UserConfig struct {
	Templates  map[string]Template `json:"templates"`
	AIProvider string              `json:"ai_provider"`
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
