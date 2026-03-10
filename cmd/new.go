package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/torryt/riff/internal"
)

// RunNew handles the `riff new` command.
func RunNew(args []string) {
	// --- Parse flags manually ---
	var templateName string
	var runCmd string
	noGit := false

	var positional []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--run":
			if i+1 < len(args) {
				i++
				runCmd = args[i]
			} else {
				fmt.Fprintln(os.Stderr, internal.Red("Error: --run requires a value"))
				os.Exit(1)
			}
		case "--no-git":
			noGit = true
		default:
			positional = append(positional, args[i])
		}
	}

	// Template name as a positional argument: `riff new bun`
	if len(positional) > 0 {
		templateName = positional[0]
	}

	// --- Interactive template picker ---
	// When no template was given, no --run was given, and stdin is a TTY,
	// show an interactive picker so the user can choose a template.
	if templateName == "" && runCmd == "" && internal.IsInteractive() {
		templates := internal.GetTemplates()
		names := make([]string, 0, len(templates))
		for k := range templates {
			names = append(names, k)
		}
		sort.Strings(names)

		// Build options: "empty" first, then sorted template names.
		options := []huh.Option[string]{
			huh.NewOption[string]("empty (just a folder + git)", ""),
		}
		for _, name := range names {
			label := name + "  " + internal.Dim(templates[name].Command)
			options = append(options, huh.NewOption[string](label, name))
		}

		var picked string
		err := huh.NewSelect[string]().
			Title("Choose a template").
			Options(options...).
			Value(&picked).
			Run()
		if err != nil {
			// User cancelled (ctrl-c) — exit cleanly.
			fmt.Fprintln(os.Stderr, internal.Dim("Cancelled."))
			os.Exit(0)
		}
		templateName = picked
	}

	// --- Ensure riff dir exists ---
	if err := internal.EnsureRiffDir(); err != nil {
		fmt.Fprintln(os.Stderr, internal.Red("Error: could not create riff directory: "+err.Error()))
		os.Exit(1)
	}

	// --- Generate project ID and create directory ---
	id := internal.GenerateID(7)
	projectPath := filepath.Join(internal.RiffDir, id)

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		fmt.Fprintln(os.Stderr, internal.Red("Error: could not create project directory: "+err.Error()))
		os.Exit(1)
	}

	// --- Run template command ---
	if templateName != "" {
		templates := internal.GetTemplates()
		tmpl, ok := templates[templateName]
		if !ok {
			fmt.Fprintln(os.Stderr, internal.Red("Error: unknown template \""+templateName+"\""))
			fmt.Fprintln(os.Stderr, internal.Dim("Available templates:"))

			names := make([]string, 0, len(templates))
			for k := range templates {
				names = append(names, k)
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Fprintln(os.Stderr, "  "+internal.Cyan(name))
			}
			// Clean up the empty directory we just created.
			os.RemoveAll(projectPath)
			os.Exit(1)
		}

		fmt.Println(internal.Dim("Running template: " + templateName))
		cmd := exec.Command("sh", "-c", tmpl.Command)
		cmd.Dir = projectPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, internal.Red("Warning: template command failed: "+err.Error()))
			// Keep the directory — don't exit
		}
	}

	// --- Run arbitrary command ---
	if runCmd != "" {
		fmt.Println(internal.Dim("Running: " + runCmd))
		cmd := exec.Command("sh", "-c", runCmd)
		cmd.Dir = projectPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, internal.Red("Warning: command failed: "+err.Error()))
		}
	}

	// --- Git init ---
	if !noGit {
		gitCmd := exec.Command("git", "init")
		gitCmd.Dir = projectPath
		// Suppress output
		_ = gitCmd.Run()

		hookPath := filepath.Join(projectPath, ".git", "hooks", "post-commit")
		hookContent := "#!/bin/sh\nriff _update-single \"$(git rev-parse --show-toplevel)\" &\n"
		if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
			// Non-fatal: git hook setup failure shouldn't abort project creation
			fmt.Fprintln(os.Stderr, internal.Dim("Note: could not write git post-commit hook: "+err.Error()))
		}
	}

	// --- Write metadata ---
	meta := internal.ProjectMetadata{
		Description: "",
		Created:     time.Now().Format(time.RFC3339),
		Template:    templateName,
		Tags:        []string{},
	}
	if err := internal.WriteMetadata(projectPath, meta); err != nil {
		fmt.Fprintln(os.Stderr, internal.Red("Warning: could not write metadata: "+err.Error()))
	}

	// --- Write cd-path ---
	if err := internal.WriteCdPath(projectPath); err != nil {
		fmt.Fprintln(os.Stderr, internal.Red("Warning: could not write cd-path: "+err.Error()))
	}

	// --- Print success ---
	fmt.Printf("  %s Created new project %s\n",
		internal.Green("+"),
		internal.Bold(internal.Cyan(id)),
	)
	fmt.Printf("  %s %s\n",
		internal.Dim("Path:"),
		projectPath,
	)
	if !internal.HasShellWrapper() {
		fmt.Printf("\n  %s %s\n",
			internal.Dim("To auto-cd into new projects, add to your shell config:"),
			internal.Cyan("eval \"$(riff init)\""),
		)
	}
}
