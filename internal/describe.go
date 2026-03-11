package internal

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const describePrompt = "Describe this project's contents and theme in exactly 7-8 words. Reply with ONLY the description, nothing else."

// ErrNoLLM is returned when no supported LLM CLI tool is installed.
var ErrNoLLM = errors.New("no supported LLM tool found")

var (
	llmProviderOnce  sync.Once
	llmProviderCache string
)

// llmProvider returns the name of the first available LLM CLI tool,
// or an empty string when none is installed. The detection result is
// cached after the first call. Setting RIFF_NO_AI=1 bypasses detection
// and always returns an empty string.
func llmProvider() string {
	if os.Getenv("RIFF_NO_AI") == "1" {
		return ""
	}
	llmProviderOnce.Do(func() {
		// Honour explicit config preference.
		if pref := ConfiguredAIProvider(); pref != "" {
			if _, err := exec.LookPath(pref); err == nil {
				llmProviderCache = pref
				return
			}
		}
		// Auto-detect: prefer Claude Code, then Copilot.
		for _, name := range []string{"claude", "copilot"} {
			if _, err := exec.LookPath(name); err == nil {
				llmProviderCache = name
				return
			}
		}
	})
	return llmProviderCache
}

// HasLLM reports whether any supported LLM CLI tool is available.
func HasLLM() bool {
	return llmProvider() != ""
}

// LLMProvider returns the name of the active LLM provider (e.g. "claude",
// "copilot") or an empty string when none is available.
func LLMProvider() string {
	return llmProvider()
}

// GenerateDescription runs an LLM CLI tool inside projectPath and returns a
// short description of the project. It returns an empty string and a non-nil
// error when generation fails or the result does not pass basic validation.
// It returns ErrNoLLM when no supported LLM tool is installed.
func GenerateDescription(projectPath string) (string, error) {
	provider := llmProvider()
	if provider == "" {
		return "", ErrNoLLM
	}

	var cmd *exec.Cmd
	switch provider {
	case "claude":
		cmd = exec.Command("claude", "-p", "--model", "haiku", "--no-session-persistence", describePrompt)
	default: // "copilot"
		cmd = exec.Command("copilot", "--model", "claude-haiku-4.5", "-sp", describePrompt)
	}
	cmd.Dir = projectPath

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s command failed: %w (stderr: %s)", provider, err, stderr.String())
	}

	// Take the last non-empty line from stdout.
	lines := strings.Split(stdout.String(), "\n")
	description := ""
	for i := len(lines) - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed != "" {
			description = trimmed
			break
		}
	}

	if description == "" {
		return "", fmt.Errorf("%s returned no output", provider)
	}

	// Strip surrounding quotes (single or double).
	if len(description) >= 2 {
		first := description[0]
		last := description[len(description)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			description = description[1 : len(description)-1]
		}
	}

	// Validate length.
	if len(description) < 5 || len(description) > 100 {
		return "", fmt.Errorf("generated description has unexpected length %d: %q", len(description), description)
	}

	return description, nil
}

// BackfillDescriptions spawns background processes to generate descriptions
// for every project in the slice that currently has an empty Description.
// It returns immediately — descriptions are written to disk (.riff.json) by
// the background processes and will be visible the next time the command runs.
// It is safe to call with an empty or fully-described slice (no-op).
func BackfillDescriptions(projects []ProjectInfo) {
	// Collect indices that need descriptions.
	var missing []int
	for i := range projects {
		if projects[i].Description == "" {
			missing = append(missing, i)
		}
	}
	if len(missing) == 0 {
		return
	}

	// No LLM available — bail early with a friendly notice.
	if llmProvider() == "" {
		fmt.Printf("  %s %s\n",
			Dim("~"),
			Dim("No AI helper found — install Claude Code or Copilot, or set \"ai_provider\" in ~/.riff/config.json"),
		)
		return
	}

	// Resolve our own binary path for spawning background processes.
	riffBin, err := os.Executable()
	if err != nil {
		riffBin = "riff" // fallback to PATH lookup
	}

	spawned := 0
	for _, idx := range missing {
		cmd := exec.Command(riffBin, "_update-single", projects[idx].Path)
		cmd.Stdout = nil
		cmd.Stderr = nil
		// Detach from parent — process survives if riff exits.
		detachProcess(cmd)
		if err := cmd.Start(); err == nil {
			spawned++
		}
	}

	if spawned > 0 {
		noun := "description"
		if spawned > 1 {
			noun = "descriptions"
		}
		fmt.Printf("  %s\n",
			Dim(fmt.Sprintf("~ Generating %d %s in background via %s",
				spawned, noun, llmProvider())))
	}
}

// UpdateProjectDescription generates a description for the project at
// projectPath, merges it into the existing .riff.json (creating one if
// absent), and returns the description.
func UpdateProjectDescription(projectPath string) (string, error) {
	description, err := GenerateDescription(projectPath)
	if err != nil {
		return "", err
	}

	// Read existing metadata, ignoring "file not found" so we start fresh.
	meta, err := ReadMetadata(projectPath)
	if err != nil {
		meta = ProjectMetadata{}
	}

	meta.Description = description

	if writeErr := WriteMetadata(projectPath, meta); writeErr != nil {
		return "", fmt.Errorf("failed to write metadata: %w", writeErr)
	}

	return description, nil
}
