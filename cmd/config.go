package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/torryt/riff/internal"
)

// RunConfig handles the "config" command group.
// Subcommands:
//
//	config init   — create a default config file at ~/.riff/config.json
//	config path   — print the config file path
func RunConfig(args []string) {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "init":
		runConfigInit()
	case "path":
		fmt.Println(internal.ConfigPath())
	case "":
		fmt.Fprintln(os.Stderr, internal.Red("Error:")+" missing subcommand. Usage: riff config <init|path>")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "%s unknown subcommand %q. Usage: riff config <init|path>\n", internal.Red("Error:"), sub)
		os.Exit(1)
	}
}

func runConfigInit() {
	configPath := internal.ConfigPath()

	// Check if config already exists.
	if _, err := os.Stat(configPath); err == nil {
		fmt.Fprintf(os.Stderr, "%s config file already exists at %s\n", internal.Yellow("Warning:"), configPath)
		fmt.Fprintf(os.Stderr, "  Remove it first if you want to re-initialize.\n")
		os.Exit(1)
	}

	// Ensure ~/.riff/ exists.
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "%s could not create directory: %s\n", internal.Red("Error:"), err)
		os.Exit(1)
	}

	// Build default config with remote schema reference.
	cfg := internal.UserConfig{
		Schema: "https://raw.githubusercontent.com/torryt/riff/main/cmd/config.schema.json",
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s could not marshal config: %s\n", internal.Red("Error:"), err)
		os.Exit(1)
	}

	if err := os.WriteFile(configPath, append(data, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "%s could not write config file: %s\n", internal.Red("Error:"), err)
		os.Exit(1)
	}

	fmt.Printf("%s Created %s\n", internal.Green("✓"), configPath)
}
