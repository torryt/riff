package cmd

import (
	"fmt"

	"github.com/torry/riff/internal"
)

// RunUpdateDocs regenerates descriptions for all riff-managed projects.
func RunUpdateDocs(args []string) {
	projects, err := internal.GetProjects()
	if err != nil {
		fmt.Println(internal.Red("  Error reading projects: " + err.Error()))
		return
	}

	if len(projects) == 0 {
		fmt.Println(internal.Yellow("  No projects found. Create one with ") + internal.Bold("riff new"))
		return
	}

	fmt.Printf("  %s for %s...\n\n",
		internal.Bold("Updating descriptions"),
		internal.Cyan(fmt.Sprintf("%d project(s)", len(projects))),
	)

	updated := 0
	failed := 0

	for _, p := range projects {
		fmt.Printf("  %s  ", internal.Cyan(p.ID))

		desc, err := internal.UpdateProjectDescription(p.Path)
		if err != nil {
			fmt.Printf("%s  %s\n", internal.Red("x"), internal.Dim("failed to generate description"))
			failed++
		} else {
			fmt.Printf("%s  %s\n", internal.Green("+"), desc)
			updated++
		}
	}

	fmt.Println()

	summary := fmt.Sprintf("  Done. %s updated", internal.Green(fmt.Sprintf("%d", updated)))
	if failed > 0 {
		summary += fmt.Sprintf(", %s failed", internal.Red(fmt.Sprintf("%d", failed)))
	}
	fmt.Println(summary)
}
