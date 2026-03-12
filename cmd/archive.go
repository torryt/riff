package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/torryt/riff/internal"
)

// RunArchive handles `riff archive` (list) and `riff archive purge [id]`.
func RunArchive(args []string) {
	if len(args) > 0 && args[0] == "purge" {
		runArchivePurge(args[1:])
		return
	}

	// Default: list archived projects.
	projects, err := internal.GetArchivedProjects()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s Could not read archive: %v\n", internal.Red("x"), err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println(internal.Dim("  No archived projects."))
		return
	}

	fmt.Printf("  %s\n\n", internal.Bold(fmt.Sprintf("%d archived project(s)", len(projects))))

	// Pre-compute age strings.
	ages := make([]string, len(projects))
	for i, p := range projects {
		ages[i] = internal.FormatAge(p.Created)
	}

	// Column widths.
	maxIDWidth := len("ID")
	maxTemplateWidth := len("Template")
	maxCreatedWidth := len("Created")
	for i, p := range projects {
		if len(p.ID) > maxIDWidth {
			maxIDWidth = len(p.ID)
		}
		if len(p.Template) > maxTemplateWidth {
			maxTemplateWidth = len(p.Template)
		}
		if len(ages[i]) > maxCreatedWidth {
			maxCreatedWidth = len(ages[i])
		}
	}

	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
		maxIDWidth, "ID",
		maxTemplateWidth, "Template",
		maxCreatedWidth, "Created",
		"Description",
	)
	fmt.Println(internal.Dim(internal.Bold(header)))

	sep := fmt.Sprintf("  %s  %s  %s  %s",
		strings.Repeat("-", maxIDWidth),
		strings.Repeat("-", maxTemplateWidth),
		strings.Repeat("-", maxCreatedWidth),
		strings.Repeat("-", len("Description")),
	)
	fmt.Println(internal.Dim(sep))

	for i, p := range projects {
		idCol := internal.Cyan(fmt.Sprintf("%-*s", maxIDWidth, p.ID))

		var templateCol string
		if p.Template == "" {
			templateCol = internal.Dim(fmt.Sprintf("%-*s", maxTemplateWidth, ""))
		} else {
			templateCol = internal.Dim(fmt.Sprintf("%-*s", maxTemplateWidth, p.Template))
		}

		createdCol := internal.Dim(fmt.Sprintf("%-*s", maxCreatedWidth, ages[i]))

		var descCol string
		if p.Description == "" {
			descCol = internal.Dim("(no description)")
		} else {
			descCol = p.Description
		}

		fmt.Printf("  %s  %s  %s  %s\n", idCol, templateCol, createdCol, descCol)
	}

	fmt.Printf("\n  %s\n", internal.Dim("Permanently delete with: riff archive purge [id]"))
}

func runArchivePurge(args []string) {
	projects, err := internal.GetArchivedProjects()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s Could not read archive: %v\n", internal.Red("x"), err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println(internal.Dim("  No archived projects to purge."))
		return
	}

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

	var toPurge []internal.ProjectInfo

	if len(args) > 0 {
		targetID := args[0]
		found := false
		for _, p := range projects {
			if p.ID == targetID {
				toPurge = append(toPurge, p)
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "  %s Archived project %q not found.\n\n", internal.Red("x"), targetID)
			fmt.Fprintf(os.Stderr, "  Archived projects:\n")
			printProjects()
			os.Exit(1)
		}
	} else if internal.IsInteractive() {
		const allValue = "__all__"
		options := []huh.Option[string]{
			huh.NewOption[string](
				internal.Bold("All archived projects")+"  "+internal.Dim(fmt.Sprintf("(%d)", len(projects))),
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
			Title("Select archived projects to permanently delete").
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

		allSelected := false
		for _, id := range selectedIDs {
			if id == allValue {
				allSelected = true
				break
			}
		}

		if allSelected {
			toPurge = projects
		} else {
			idSet := make(map[string]bool, len(selectedIDs))
			for _, id := range selectedIDs {
				idSet[id] = true
			}
			for _, p := range projects {
				if idSet[p.ID] {
					toPurge = append(toPurge, p)
				}
			}
		}
	} else {
		fmt.Println("  Archived projects:")
		fmt.Println()
		printProjects()
		fmt.Printf("\n  %s\n", internal.Cyan("Use: riff archive purge <id>"))
		return
	}

	// Confirm permanent deletion.
	var confirmed bool
	confirmMsg := fmt.Sprintf("Permanently delete %d project(s)? This cannot be undone.", len(toPurge))

	if internal.IsInteractive() {
		err := huh.NewConfirm().
			Title(confirmMsg).
			Affirmative("Yes, delete permanently").
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

	deleted := 0
	for _, p := range toPurge {
		if err := os.RemoveAll(p.Path); err != nil {
			fmt.Printf("  %s Failed to delete %s\n", internal.Red("x"), internal.Cyan(p.ID))
		} else {
			fmt.Printf("  %s Deleted %s\n", internal.Red("-"), internal.Cyan(p.ID))
			deleted++
		}
	}

	fmt.Printf("  %s %s permanently deleted\n", internal.Bold("Done."), internal.Red(fmt.Sprintf("%d", deleted)))
}
