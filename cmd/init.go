package cmd

import (
	"fmt"
	"os"

	"github.com/torryt/riff/internal"
)

const bashWrapper = `riff() {
  RIFF_WRAPPER=1 command riff "$@"
  local cd_path="$HOME/.riff/.cd-path"
  if [ -f "$cd_path" ]; then
    local target
    target=$(cat "$cd_path")
    rm -f "$cd_path"
    if [ -n "$target" ] && [ -d "$target" ]; then
      cd "$target" || return
    fi
  fi
}
`

const fishWrapper = `function riff
    set -lx RIFF_WRAPPER 1
    command riff $argv
    set -l cd_path "$HOME/.riff/.cd-path"
    if test -f "$cd_path"
        set -l target (cat "$cd_path")
        rm -f "$cd_path"
        if test -n "$target" -a -d "$target"
            cd "$target"
        end
    end
end
`

// RunInit outputs a shell wrapper function for the given or detected shell.
// Usage: eval "$(riff init)" or eval "$(riff init bash)" or riff init fish | source
func RunInit(args []string) {
	shell := ""
	if len(args) > 0 {
		shell = args[0]
	} else {
		shell = internal.DetectShell()
		if shell == "" {
			fmt.Fprintln(os.Stderr, internal.Red("Error: could not detect shell. Specify one: riff init bash|zsh|fish"))
			os.Exit(1)
		}
	}

	switch shell {
	case "bash", "zsh":
		fmt.Print(bashWrapper)
	case "fish":
		fmt.Print(fishWrapper)
	default:
		fmt.Fprintf(os.Stderr, internal.Red("Error: unsupported shell %q. Supported: bash, zsh, fish\n"), shell)
		os.Exit(1)
	}
}
