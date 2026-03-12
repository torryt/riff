package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/torryt/riff/internal"
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

	// Skip description generation — clean only needs project names/paths.

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
			age := internal.FormatAge(p.Created)
			label := p.ID + "  " + desc + "  " + age
			options = append(options, huh.NewOption[string](label, p.ID))
		}

		var selectedIDs []string
		err := huh.NewMultiSelect[string]().
			Title("Select projects to archive").
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

	// Confirm archival.
	var confirmed bool
	confirmMsg := fmt.Sprintf("Archive %d project(s)?", len(toDelete))

	if internal.IsInteractive() {
		err := huh.NewConfirm().
			Title(confirmMsg).
			Affirmative("Yes, archive").
			Negative("Cancel").
			Value(&confirmed).
			Run()
		if err != nil || !confirmed {
			fmt.Println("  Aborted.")
			return
		}
	} else {
		fmt.Printf("  %s [y/N] ", confirmMsg)
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("  Aborted.")
			return
		}
	}

	// Archive projects.
	archived := 0
	for _, p := range toDelete {
		if err := internal.ArchiveProject(p); err != nil {
			fmt.Printf("  %s Failed to archive %s: %v\n", internal.Red("x"), internal.Cyan(p.ID), err)
		} else {
			fmt.Printf("  %s Archived %s\n", internal.Yellow("~"), internal.Cyan(p.ID))
			archived++
		}
	}

	fmt.Printf("  %s %s archived\n", internal.Bold("Done."), internal.Yellow(fmt.Sprintf("%d", archived)))
	fmt.Printf("  %s\n", internal.Dim("Restore with: riff archive  •  Purge with: riff archive purge"))
}
