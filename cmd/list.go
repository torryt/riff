package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/torryt/riff/internal"
)

// RunList prints a table of all riff-managed projects.
func RunList(args []string) {
	projects, err := internal.GetProjects()
	if err != nil {
		fmt.Println(internal.Red("  Error reading projects: " + err.Error()))
		return
	}

	if len(projects) == 0 {
		fmt.Println(internal.Yellow("  No projects found. Create one with ") + internal.Bold("riff new"))
		return
	}

	// Fill in any missing descriptions in the background before displaying.
	internal.BackfillDescriptions(projects, 15*time.Second)

	fmt.Printf("  %s\n\n", internal.Bold(fmt.Sprintf("%d project(s)", len(projects))))

	// Pre-compute age strings so we can measure their widths.
	ages := make([]string, len(projects))
	for i, p := range projects {
		ages[i] = internal.FormatAge(p.Created)
	}

	// Calculate column widths.
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

	// Header row.
	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
		maxIDWidth, "ID",
		maxTemplateWidth, "Template",
		maxCreatedWidth, "Created",
		"Description",
	)
	fmt.Println(internal.Dim(internal.Bold(header)))

	// Separator row.
	sep := fmt.Sprintf("  %s  %s  %s  %s",
		strings.Repeat("-", maxIDWidth),
		strings.Repeat("-", maxTemplateWidth),
		strings.Repeat("-", maxCreatedWidth),
		strings.Repeat("-", len("Description")),
	)
	fmt.Println(internal.Dim(sep))

	// Data rows.
	for i, p := range projects {
		idCol := internal.Cyan(fmt.Sprintf("%-*s", maxIDWidth, p.ID))

		templateVal := p.Template
		var templateCol string
		if templateVal == "" {
			templateCol = internal.Dim(fmt.Sprintf("%-*s", maxTemplateWidth, ""))
		} else {
			templateCol = internal.Dim(fmt.Sprintf("%-*s", maxTemplateWidth, templateVal))
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
}
