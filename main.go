package main

import (
	"fmt"
	"os"

	"github.com/torry/riff/cmd"
	"github.com/torry/riff/internal"
)

func main() {
	args := os.Args[1:]

	var command string
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	switch command {
	case "new":
		cmd.RunNew(args)
	case "list", "ls":
		cmd.RunList(args)
	case "open":
		cmd.RunOpen(args)
	case "clean", "rm":
		cmd.RunClean(args)
	case "update-docs":
		cmd.RunUpdateDocs(args)
	case "_update-single":
		// Hidden command: called from git post-commit hook in background.
		// Silently regenerates the project description after a commit.
		if len(args) > 0 {
			_, _ = internal.UpdateProjectDescription(args[0])
		}
	case "help", "--help", "-h", "":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "%s unknown command %q\n\n", internal.Red("Error:"), command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf("\n  %s %s\n\n", internal.Bold(internal.Cyan("riff")), internal.Dim("— manage throwaway projects"))

	fmt.Printf("  %s  %s %s\n\n", internal.Bold("Usage:"), internal.Cyan("riff"), "<command> [options]")

	fmt.Printf("  %s\n", internal.Bold("Commands:"))
	fmt.Printf("    %-20s %s\n", internal.Green("new"), "Create a new project")
	fmt.Printf("    %-20s %s\n", internal.Green("list")+internal.Dim(", ls"), "List all projects")
	fmt.Printf("    %-20s %s\n", internal.Green("open")+" "+internal.Dim("[id]"), "Open a project (picks from list if no ID)")
	fmt.Printf("    %-20s %s\n", internal.Green("clean")+" "+internal.Dim("[id]"), "Delete projects")
	fmt.Printf("    %-20s %s\n", internal.Green("update-docs"), "Regenerate descriptions for all projects")
	fmt.Printf("    %-20s %s\n", internal.Green("help"), "Show this help message")

	fmt.Printf("\n  %s %s\n", internal.Bold("Flags"), internal.Dim("(for new):"))
	fmt.Printf("    %-28s %s\n", internal.Dim("-t, --template")+" <name>", "Use a project template")
	fmt.Printf("    %-28s %s\n", internal.Dim("--run")+" "+internal.Dim(`"<cmd>"`), "Run arbitrary init command")
	fmt.Printf("    %-28s %s\n", internal.Dim("--no-git"), "Skip git initialization")
	fmt.Println()
}
