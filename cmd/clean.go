package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/torry/riff/internal"
)

// RunClean handles the `riff clean [id]` command.
// It deletes the on-disk project directory (and its contents) after prompting
// the user for confirmation.
func RunClean(args []string) {
	projects, err := internal.GetProjects()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s Could not read projects: %v\n", internal.Red("x"), err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println(internal.Yellow("  No projects to clean."))
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
			fmt.Printf("  %s - %s\n", internal.Cyan(p.ID), desc)
		}
	}

	// toDelete is the set of projects that will be removed.
	var toDelete []internal.ProjectInfo

	if len(args) > 0 {
		// A specific project ID was given.
		targetID := args[0]
		found := false
		for _, p := range projects {
			if p.ID == targetID {
				toDelete = append(toDelete, p)
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
		// No ID supplied + TTY — show interactive multi-select.

		// Build options: one per project, plus "All projects" at the top.
		const allValue = "__all__"
		options := []huh.Option[string]{
			huh.NewOption[string](
				internal.Bold("All projects")+"  "+internal.Dim(fmt.Sprintf("(%d)", len(projects))),
				allValue,
			),
		}
		for _, p := range projects {
			desc := p.Description
			if desc == "" {
				desc = "(no description)"
			}
			label := p.ID + "  " + desc
			options = append(options, huh.NewOption[string](label, p.ID))
		}

		var selectedIDs []string
		err := huh.NewMultiSelect[string]().
			Title("Select projects to delete").
			Options(options...).
			Value(&selectedIDs).
			Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, internal.Dim("Cancelled."))
			os.Exit(0)
		}

		if len(selectedIDs) == 0 {
			fmt.Println(internal.Dim("  No projects selected."))
			return
		}

		// Check if "All projects" was selected.
		allSelected := false
		for _, id := range selectedIDs {
			if id == allValue {
				allSelected = true
				break
			}
		}

		if allSelected {
			toDelete = projects
		} else {
			// Map selected IDs to project infos.
			idSet := make(map[string]bool, len(selectedIDs))
			for _, id := range selectedIDs {
				idSet[id] = true
			}
			for _, p := range projects {
				if idSet[p.ID] {
					toDelete = append(toDelete, p)
				}
			}
		}
	} else {
		// No ID, not interactive — list projects and guide the user.
		fmt.Println("  Available projects:")
		fmt.Println()
		printProjects()
		fmt.Printf("\n  %s\n", internal.Cyan("Use: riff clean <id>"))
		return
	}

	// Confirm deletion.
	var confirmed bool
	confirmMsg := fmt.Sprintf("Delete %d project(s)? This cannot be undone.", len(toDelete))

	if internal.IsInteractive() {
		err := huh.NewConfirm().
			Title(confirmMsg).
			Affirmative("Yes, delete").
			Negative("Cancel").
			Value(&confirmed).
			Run()
		if err != nil || !confirmed {
			fmt.Println("  Aborted.")
			return
		}
	} else {
		// Non-interactive: when called with an explicit ID, use a simple y/N prompt.
		fmt.Printf("  %s [y/N] ", confirmMsg)
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("  Aborted.")
			return
		}
	}

	// Delete projects.
	deleted := 0
	for _, p := range toDelete {
		if err := os.RemoveAll(p.Path); err != nil {
			fmt.Printf("  %s Failed to delete %s\n", internal.Red("x"), internal.Cyan(p.ID))
		} else {
			fmt.Printf("  %s Deleted %s\n", internal.Red("-"), internal.Cyan(p.ID))
			deleted++
		}
	}

	fmt.Printf("  %s %s deleted\n", internal.Bold("Done."), internal.Red(fmt.Sprintf("%d", deleted)))
}
