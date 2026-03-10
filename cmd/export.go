package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/torryt/riff/internal"
)

// RunExport handles the `riff export <folder> [id]` command.
// It copies the riff project directory into the specified folder (relative to
// the current working directory). If no ID is provided the user selects from
// an interactive list.
func RunExport(args []string) {
	// Parse positional args: first is destination folder, second (optional) is ID.
	var destFolder string
	var projectID string

	for _, a := range args {
		if destFolder == "" {
			destFolder = a
		} else if projectID == "" {
			projectID = a
		}
	}

	if destFolder == "" {
		fmt.Fprintf(os.Stderr, "  %s Usage: riff export <folder> [id]\n", internal.Red("x"))
		os.Exit(1)
	}

	projects, err := internal.GetProjects()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s Could not read projects: %v\n", internal.Red("x"), err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println(internal.Yellow("  No projects found. Run `riff new` to create one."))
		return
	}

	// Fill in any missing descriptions before displaying the picker.
	internal.BackfillDescriptions(projects, 15*time.Second)

	// Helper: print available projects for non-interactive fallback.
	printProjects := func() {
		for _, p := range projects {
			desc := p.Description
			if desc == "" {
				desc = internal.Dim("(no description)")
			}
			fmt.Printf("  %s - %s\n", internal.Cyan(p.ID), desc)
		}
	}

	// Resolve the project ID.
	if projectID != "" {
		// ID supplied on command line — verify it exists.
		found := false
		for _, p := range projects {
			if p.ID == projectID {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "  %s Project %q not found.\n\n", internal.Red("x"), projectID)
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
			label := p.ID + "  " + desc
			options = append(options, huh.NewOption[string](label, p.ID))
		}

		err := huh.NewSelect[string]().
			Title("Select a project to export").
			Options(options...).
			Value(&projectID).
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
		fmt.Printf("\n  Run %s to export a project.\n", internal.Cyan("riff export <folder> <id>"))
		return
	}

	// Resolve the source path.
	srcPath := filepath.Join(internal.RiffDir, projectID)

	// Resolve destination: relative to current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s Could not determine current directory: %v\n", internal.Red("x"), err)
		os.Exit(1)
	}
	destPath := filepath.Join(cwd, destFolder)

	// Refuse to overwrite an existing destination.
	if _, statErr := os.Stat(destPath); statErr == nil {
		fmt.Fprintf(os.Stderr, "  %s Destination %q already exists. Choose a different folder name.\n",
			internal.Red("x"), destFolder)
		os.Exit(1)
	}

	// Copy the project tree.
	if err := copyDir(srcPath, destPath); err != nil {
		fmt.Fprintf(os.Stderr, "  %s Export failed: %v\n", internal.Red("x"), err)
		os.Exit(1)
	}

	fmt.Printf("  %s Exported %s → %s\n",
		internal.Green("+"),
		internal.Cyan(projectID),
		internal.Bold(destFolder),
	)
}

// copyDir recursively copies src into dst, preserving file modes.
// dst must not already exist.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Compute the corresponding destination path.
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target, info.Mode())
	})
}

// copyFile copies a single file from src to dst, creating dst with the given mode.
func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
