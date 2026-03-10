package cmd

import (
	"fmt"

	"github.com/torryt/riff/internal"
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

	if !internal.HasLLM() {
		fmt.Printf("  %s\n\n", internal.Dim("No AI helper found — install one to generate descriptions:"))
		fmt.Println(internal.Dim("  Supported: claude (Claude Code), copilot (GitHub Copilot CLI)"))
		fmt.Println(internal.Dim("  Set a default with \"ai_provider\" in ~/.riff/config.json"))
		return
	}

	fmt.Printf("  %s for %s %s\n\n",
		internal.Bold("Updating descriptions"),
		internal.Cyan(fmt.Sprintf("%d project(s)", len(projects))),
		internal.Dim("via "+internal.LLMProvider()),
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
