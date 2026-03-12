package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/torryt/riff/cmd"
	"github.com/torryt/riff/internal"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// padRight pads s to width based on visible (non-ANSI) length.
func padRight(s string, width int) string {
	visible := len(ansiRe.ReplaceAllString(s, ""))
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}

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
	case "archive":
		cmd.RunArchive(args)
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
	fmt.Printf("    %s %s\n", padRight(internal.Green("new")+" "+internal.Dim("[template]"), 20), "Create a new project")
	fmt.Printf("    %s %s\n", padRight(internal.Green("list")+internal.Dim(", ls"), 20), "List all projects")
	fmt.Printf("    %s %s\n", padRight(internal.Green("open")+" "+internal.Dim("[id]"), 20), "Open a project (picks from list if no ID)")
	fmt.Printf("    %s %s\n", padRight(internal.Green("clean")+" "+internal.Dim("[id]"), 20), "Archive projects (move to ~/.riff/archive)")
	fmt.Printf("    %s %s\n", padRight(internal.Green("archive"), 20), "List archived projects")
	fmt.Printf("    %s %s\n", padRight(internal.Green("archive purge")+" "+internal.Dim("[id]"), 20), "Permanently delete archived projects")
	fmt.Printf("    %s %s\n", padRight(internal.Green("export")+" "+internal.Dim("<path> [id]"), 20), "Export a project to a folder (created if needed)")
	fmt.Printf("    %s %s\n", padRight(internal.Green("init")+" "+internal.Dim("[shell]"), 20), "Shell setup for auto-cd (auto-detects shell)")
	fmt.Printf("    %s %s\n", padRight(internal.Green("config")+" "+internal.Dim("<init|path>"), 20), "Manage configuration")
	fmt.Printf("    %s %s\n", padRight(internal.Green("update-docs"), 20), "Regenerate descriptions for all projects")
	fmt.Printf("    %s %s\n", padRight(internal.Green("help"), 20), "Show this help message")

	fmt.Printf("\n  %s\n", internal.Bold("Flags:"))
	fmt.Printf("    %s %s\n", padRight(internal.Dim("--version, -v"), 20), "Print version")

	fmt.Printf("\n  %s %s\n", internal.Bold("Flags"), internal.Dim("(for new):"))
	fmt.Printf("    %s %s\n", padRight(internal.Dim("--run")+" "+internal.Dim(`"<cmd>"`), 20), "Run arbitrary init command")
	fmt.Printf("    %s %s\n", padRight(internal.Dim("--no-git"), 20), "Skip git initialization")
	fmt.Println()
}
