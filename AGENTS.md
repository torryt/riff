# AGENTS.md

## Build & Run

```bash
go build -o riff .        # build
go run . <command>         # run without building
CGO_ENABLED=0 go build -ldflags="-s -w" -o riff .  # release build
```

## Test & Lint

No test suite yet. Validate with:

```bash
go vet ./...
gofmt -l .                # should print nothing
```

## Project Structure

```
main.go          — entry point, command routing, help text
cmd/             — one file per command, exports RunXxx(args []string)
internal/        — shared utilities: config, projects, colors, describe, prompt
.goreleaser.yml  — cross-platform release config
```

Data lives in `~/.riff/`. Per-project metadata in `.riff.json` (not `package.json`).

## Adding a Command

1. Create `cmd/my_command.go` in package `cmd`
2. Export `func RunMyCommand(args []string)`
3. Add a `case` in the `switch` in `main.go`
4. Update `printHelp()`

## Conventions

- **Go 1.26.1** minimum
- **stdlib-first** — no CLI frameworks (cobra, urfave/cli). Flags are parsed manually from `os.Args`
- **Minimal dependencies** — only `charmbracelet/huh` for interactive TUI prompts
- **NO_COLOR** — all terminal color output must go through `internal/colors.go` helpers (`Green()`, `Red()`, `Bold()`, etc.) which respect the `NO_COLOR` env var
- **Interactive fallbacks** — every interactive prompt (`charmbracelet/huh`) must have a non-interactive fallback; use `internal.IsInteractive()` to branch
- **gofmt + go vet** — code must pass both with zero warnings
- **Commit prefixes** — use `docs:`, `test:`, `ci:` etc. These are filtered from release changelogs
