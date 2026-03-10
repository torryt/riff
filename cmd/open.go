package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/torryt/riff/internal"
)

// RunOpen handles the `riff open [id]` command.
// It looks up the project by ID and writes its path to the cd-path file so the
// shell wrapper can cd into it after riff exits.
func RunOpen(args []string) {
	projects, err := internal.GetProjects()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s Could not read projects: %v\n", internal.Red("x"), err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println(internal.Yellow("  No projects found. Run `riff new` to create one."))
		return
	}

	// Fill in any missing descriptions in the background before displaying.
	internal.BackfillDescriptions(projects, 15*time.Second)

	// Helper: print available projects (used in non-interactive fallback).
	printProjects := func() {
		for _, p := range projects {
			desc := p.Description
			if desc == "" {
				desc = internal.Dim("(no description)")
			}
			age := internal.Dim(internal.FormatAge(p.Created))
			fmt.Printf("  %s - %s %s\n", internal.Cyan(p.ID), desc, age)
		}
	}

	var selectedID string

	if len(args) > 0 {
		// A project ID was provided; verify it exists.
		targetID := args[0]
		found := false
		for _, p := range projects {
			if p.ID == targetID {
				selectedID = p.ID
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "  %s Project %q not found.\n\n", internal.Red("x"), targetID)
			fmt.Fprintf(os.Stderr, "  Available projects:\n")
			printProjects()
			os.Exit(1)
		}
	} else if internal.IsInteractive() {
		// No ID supplied + TTY — show interactive picker.
		options := make([]huh.Option[string], 0, len(projects))
		for _, p := range projects {
			desc := p.Description
			if desc == "" {
				desc = "(no description)"
			}
			age := internal.FormatAge(p.Created)
			label := p.ID + "  " + desc + "  " + age
			options = append(options, huh.NewOption[string](label, p.ID))
		}

		err := huh.NewSelect[string]().
			Title("Open a project").
			Options(options...).
			Value(&selectedID).
			Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, internal.Dim("Cancelled."))
			os.Exit(0)
		}
	} else {
		// No ID, not interactive — list and exit.
		fmt.Println("  Available projects:")
		fmt.Println()
		printProjects()
		fmt.Printf("\n  Run %s to open a project.\n", internal.Cyan("riff open <id>"))
		return
	}

	projectPath := filepath.Join(internal.ProjectsDir, selectedID)

	fmt.Printf("  %s %s\n", internal.Green(">"), internal.Bold(internal.Cyan(selectedID)))

	if err := internal.WriteCdPath(projectPath); err != nil {
		fmt.Fprintf(os.Stderr, "  %s Failed to write cd-path: %v\n", internal.Red("x"), err)
		os.Exit(1)
	}

	if !internal.HasShellWrapper() {
		fmt.Printf("\n  %s %s\n",
			internal.Dim("To auto-cd into projects, add to your shell config:"),
			internal.Cyan("eval \"$(riff init)\""),
		)
	}
}
