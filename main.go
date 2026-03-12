package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/torryt/riff/cmd"
	"github.com/torryt/riff/internal"
)

// version is set at build time via -ldflags "-X main.version=..."
// When installed via `go install`, it falls back to the module version.
var version = "dev"

func getVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

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
	case "init":
		cmd.RunInit(args)
	case "config":
		cmd.RunConfig(args)
	case "export":
		cmd.RunExport(args)
	case "update-docs":
		cmd.RunUpdateDocs(args)
	case "_update-single":
		// Hidden command: called from git post-commit hook in background.
		// Silently regenerates the project description after a commit.
		if len(args) > 0 {
			_, _ = internal.UpdateProjectDescription(args[0])
		}
	case "":
		cmd.RunNew(args)
	case "version", "--version", "-v":
		fmt.Println("riff " + getVersion())
	case "help", "--help", "-h":
		printHelp()
	default:
		// Treat unknown commands as template names: `riff bun` → `riff new bun`
		templates := internal.GetTemplates()
		if _, ok := templates[command]; ok {
			cmd.RunNew(append([]string{command}, args...))
		} else {
			fmt.Fprintf(os.Stderr, "%s unknown command %q\n\n", internal.Red("Error:"), command)
			printHelp()
			os.Exit(1)
		}
	}
}

func printHelp() {
	fmt.Printf("\n  %s %s\n\n", internal.Bold(internal.Cyan("riff")), internal.Dim("— manage throwaway projects"))

	fmt.Printf("  %s  %s %s\n\n", internal.Bold("Usage:"), internal.Cyan("riff"), "[command|template] [options]")

	fmt.Printf("  %s\n", internal.Bold("Commands:"))
	fmt.Printf("    %-20s %s\n", internal.Green("new")+" "+internal.Dim("[template]"), "Create a new project")
	fmt.Printf("    %-20s %s\n", internal.Green("list")+internal.Dim(", ls"), "List all projects")
	fmt.Printf("    %-20s %s\n", internal.Green("open")+" "+internal.Dim("[id]"), "Open a project (picks from list if no ID)")
	fmt.Printf("    %-20s %s\n", internal.Green("clean")+" "+internal.Dim("[id]"), "Delete projects")
	fmt.Printf("    %-20s %s\n", internal.Green("export")+" "+internal.Dim("<path> [id]"), "Export a project to a folder (created if needed)")
	fmt.Printf("    %-20s %s\n", internal.Green("init")+" "+internal.Dim("[shell]"), "Shell setup for auto-cd (auto-detects shell)")
	fmt.Printf("    %-20s %s\n", internal.Green("config")+" "+internal.Dim("<init|path>"), "Manage configuration")
	fmt.Printf("    %-20s %s\n", internal.Green("update-docs"), "Regenerate descriptions for all projects")
	fmt.Printf("    %-20s %s\n", internal.Green("help"), "Show this help message")

	fmt.Printf("\n  %s\n", internal.Bold("Flags:"))
	fmt.Printf("    %-28s %s\n", internal.Dim("--version, -v"), "Print version")

	fmt.Printf("\n  %s %s\n", internal.Bold("Flags"), internal.Dim("(for new):"))
	fmt.Printf("    %-28s %s\n", internal.Dim("--run")+" "+internal.Dim(`"<cmd>"`), "Run arbitrary init command")
	fmt.Printf("    %-28s %s\n", internal.Dim("--no-git"), "Skip git initialization")
	fmt.Println()
}
