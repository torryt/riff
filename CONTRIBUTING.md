# 🤝 Contributing to riff

First off — thanks for wanting to contribute to a tool for managing *temporary* projects. The irony is not lost on us. 😄

## 🛠️ Getting started

```bash
# Clone the repo
git clone https://github.com/torryt/riff.git
cd riff

# Build it
go build -o riff .

# Run it
./riff new
./riff list
```

That's it. No 47-step setup guide. One binary, zero runtime dependencies. You're welcome.

## 🏃 Running locally

```bash
# Build and run in one go
go run . new
go run . list
go run . open
go run . clean

# Or build the binary and use it like a real person
go build -o riff .
./riff new
```

## 📁 Project structure

```
riff/
├── main.go              # 🚪 Entry point & command routing
├── go.mod
├── internal/
│   ├── config.go        # 📌 Constants, config loading, template registry
│   ├── projects.go      # 📦 Project listing, ID generation, metadata read/write
│   ├── describe.go      # 🤖 LLM subprocess calls (Copilot CLI)
│   ├── colors.go        # 🎨 ANSI color helpers (respects NO_COLOR)
│   └── prompt.go        # 🖥️ TTY detection
├── cmd/
│   ├── new.go           # 🆕 riff new
│   ├── list.go          # 📋 riff list
│   ├── open.go          # 📂 riff open
│   ├── clean.go         # 🧹 riff clean
│   └── update_docs.go   # 🤖 riff update-docs
├── .goreleaser.yml      # 📦 Cross-platform release config
├── .gitignore
└── LICENSE
```

## ➕ Adding a new command

1. Create a new file in `cmd/` — follow the existing pattern (export a `RunXxx(args []string)` function)
2. Wire it up in `main.go` — add a case to the `switch` statement
3. Add it to `printHelp()` so people actually know it exists
4. Update the README while you're at it. Future you will appreciate this.

Example skeleton:

```go
// cmd/my_cool_command.go
package cmd

import (
    "fmt"
    "github.com/torryt/riff/internal"
)

func RunMyCoolCommand(args []string) {
    fmt.Printf("  %s Did the cool thing\n", internal.Green("✓"))
}
```

## 📐 Code style & conventions

- **Go** — standard `go vet` and `gofmt`. The compiler is your friend, even when it doesn't feel like it.
- **Stdlib-first** — reach for the standard library before adding dependencies. The `node_modules` diet is real and we're committed. 🥗
- **`NO_COLOR` support** — the color system in `internal/colors.go` already respects this. Don't bypass it with raw ANSI codes.
- **Interactive prompts** — use [`charmbracelet/huh`](https://github.com/charmbracelet/huh) for interactive TUI elements. Always check `internal.IsInteractive()` and provide a non-interactive fallback.
- **No frameworks** — no cobra, no urfave/cli, no viper. Flag parsing is manual and that's fine — the CLI has five commands.

## 🐛 Reporting bugs

Open an issue. Include:
- What you ran
- What you expected
- What actually happened
- Your OS and Go version (`go version`)

Screenshots welcome. Interpretive dances less so, but we respect the effort. 💃

## 💡 Feature ideas

Got an idea? Open an issue and describe:
- **What** you want
- **Why** it's useful
- **How** you'd use it

We're pretty open-minded, but "rewrite it in Rust" will be politely declined. We just *left* a rewrite. 🦀❌

## 🔑 Key things to know

- **All project data lives in `~/.riff/`** — each project gets its own subdirectory with a random 7-character ID
- **Metadata lives in `.riff.json`** inside each project directory (not `package.json` — we're language-agnostic now)
- **`GenerateDescription()` in `internal/describe.go`** calls GitHub Copilot CLI under the hood — keep this in mind if you're working offline
- **The post-commit hook** (`cmd/new.go`) auto-updates descriptions after each commit in a riff project
- **Templates are just shell commands** — adding one is literally adding a line to a Go map in `internal/config.go`

## 🚢 Releasing

Releases are handled by [GoReleaser](https://goreleaser.com/). Tag a version and push:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GoReleaser builds binaries for linux/darwin/windows on amd64/arm64 and uploads them to GitHub Releases. See `.goreleaser.yml` for the config.

---

Happy hacking! 🎸
