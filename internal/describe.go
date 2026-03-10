package internal

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const describePrompt = "Describe this project's contents and theme in exactly 7-8 words. Reply with ONLY the description, nothing else."

// GenerateDescription runs the copilot CLI inside projectPath and returns a
// short description of the project. It returns an empty string and a non-nil
// error when generation fails or the result does not pass basic validation.
func GenerateDescription(projectPath string) (string, error) {
	cmd := exec.Command("copilot", "--model", "claude-haiku-4.5", "-sp", describePrompt)
	cmd.Dir = projectPath

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("copilot command failed: %w (stderr: %s)", err, stderr.String())
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
		return "", errors.New("copilot returned no output")
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

// BackfillDescriptions generates descriptions for every project in the
// slice that currently has an empty Description. Generations run in
// parallel and are capped by timeout. Projects whose descriptions arrive
// in time are updated both in the slice and on disk (.riff.json).
// The function prints a brief status line while working and clears it when
// done. It is safe to call with an empty or fully-described slice (no-op).
func BackfillDescriptions(projects []ProjectInfo, timeout time.Duration) {
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

	// Show a brief status line.
	noun := "description"
	if len(missing) > 1 {
		noun = "descriptions"
	}
	statusMsg := fmt.Sprintf("  %s Generating %d %s…",
		Dim("~"), len(missing), noun)
	fmt.Print(statusMsg)

	type result struct {
		index int
		desc  string
	}

	results := make(chan result, len(missing))
	var wg sync.WaitGroup

	for _, idx := range missing {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			desc, err := UpdateProjectDescription(projects[i].Path)
			if err == nil && desc != "" {
				results <- result{index: i, desc: desc}
			}
		}(idx)
	}

	// Close results channel once all goroutines finish.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results until timeout or all done.
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	filled := 0
loop:
	for {
		select {
		case r, ok := <-results:
			if !ok {
				break loop // channel closed, all done
			}
			projects[r.index].Description = r.desc
			filled++
		case <-timer.C:
			break loop
		}
	}

	// Clear the status line.
	fmt.Print("\r" + strings.Repeat(" ", len(statusMsg)) + "\r")

	_ = filled
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
