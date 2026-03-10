# 🎸 riff

> Throwaway projects for every language, without the existential dread of naming them.

**riff** is a CLI tool for creating and managing disposable project workspaces. Need to spike something out? Test a weird idea at 2am? Prototype that thing your coworker said was "impossible"? Spin up an isolated project in any language, hack away, and clean up when you're done — or don't. We won't judge.

Each project lives in `~/.riff/` and gets an auto-generated AI description so you can figure out what past-you was thinking. 🔮

## Index

- [Features](#-features)
- [Installation](#-installation)
- [Usage](#-usage)
  - [Commands](#commands)
  - [Quick start](#quick-start-)
  - [Flags for `new`](#flags-for-new)
- [How it works](#-how-it-works)
- [Configuration](#%EF%B8%8F-configuration)
  - [Initializing the config](#initializing-the-config)
  - [Config options](#config-options)

- [AI descriptions](#-ai-descriptions)
- [Templates](#-templates)
  - [Custom templates](#custom-templates)
- [Shell integration](#-shell-integration)
- [Contributing](#-contributing)
- [License](#-license)

## ✨ Features

- 🆕 **Create** isolated projects in one command — any language, any framework
- 📋 **List** all your projects with AI-generated descriptions
- 📂 **Open** projects interactively or by ID
- 📤 **Export** a project to any local folder when you're ready to ship it
- 🧹 **Clean** up projects individually or in bulk (Marie Kondo mode included)
- 🤖 **Auto-describe** projects via Claude Code or GitHub Copilot CLI — because you *will* forget what this one does
- 🌍 **Framework-agnostic** — Bun, Python, Rust, Go, Node, React, Next.js, or just an empty folder

## 📦 Installation

**Prerequisites:**
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) or [GitHub Copilot CLI](https://docs.github.com/en/copilot/github-copilot-in-the-cli) for auto-generated descriptions (optional, but your future self will thank you)

**Homebrew (macOS/Linux):**

```bash
brew install torryt/tap/riff
```

**Go:**

```bash
go install github.com/torryt/riff@latest
```

Or grab a prebuilt binary from [Releases](https://github.com/torryt/riff/releases) if commitment isn't your thing.

## 🚀 Usage

```
riff <command> [options]
```

### Commands

| Command | Description |
|---|---|
| `riff new [template]` | Create a fresh project (pass template name or use interactive picker) |
| `riff list` (or `ls`) | List all projects with descriptions |
| `riff open [id]` | Open a project (interactive picker if no ID) |
| `riff clean [id]` (or `rm`) | Delete projects (multi-select if no ID) |
| `riff export <folder> [id]` | Export a project to a local folder (interactive picker if no ID) |
| `riff config <init\|path>` | Manage configuration (initialize or print path) |
| `riff init [shell]` | Shell setup for auto-cd (auto-detects shell) |
| `riff update-docs` | Regenerate descriptions for all projects |
| `riff help` | Show help |

### Quick start 🏃

```bash
# Create a new project — pick a template or go empty
riff new

# See what you've been up to
riff list

# Jump back into one
riff open

# Marie Kondo the ones that no longer spark joy
riff clean

# Graduate a prototype to a real project
riff export ~/code/my-new-thing
```

### Flags for `new`

```bash
riff new                        # interactive picker (or empty folder + git)
riff new bun                 # bun init
riff new dotnet              # dotnet new console
riff new react               # create-vite react-ts
riff new python              # uv init
riff new rust                # cargo init
riff new go                  # go mod init temp
riff new node                # npm init
riff new next                # create-next-app
riff new --run "uv init"        # arbitrary init command
riff new --no-git               # skip git init
```

## 🧠 How it works

1. `riff new` creates a directory under `~/.riff/` with a random 7-char ID, optionally runs a template command, and sets up a git repo with a post-commit hook
2. Every time you commit, the hook asks your AI provider to summarize your project in ~7 words (it's surprisingly good at this)
3. `riff list` shows all your projects with their descriptions — no more opening 14 folders to find the one with the WebSocket experiment
4. `riff clean` lets you select and delete projects when the guilt of digital hoarding sets in

## ⚙️ Configuration

riff is configured via a JSON file at `~/.riff/config.json`. All settings are optional — riff works out of the box with sensible defaults.

### Initializing the config

Generate a starter config file:

```bash
riff config init
```

This creates `~/.riff/config.json`.

To print the config file path (useful in scripts):

```bash
riff config path
```

### Config options

| Key | Type | Default | Description |
|---|---|---|---|
| `ai_provider` | `string` | `""` (auto-detect) | Preferred AI provider for project descriptions. Valid values: `"claude"`, `"copilot"`, or `""`. When empty, riff auto-detects (Claude Code preferred). Falls back to auto-detection if the chosen provider is not in `$PATH`. |
| `templates` | `object` | `{}` | Custom project templates. Each key is a template name used with `riff new <name>`. Templates with the same name as a built-in override it. See [Custom templates](#custom-templates). |
| `templates.<name>.command` | `string` | — | Shell command to initialize the project. Runs via `sh -c` in the new project directory. |

**Full example:**

```json
{
  "ai_provider": "copilot",
  "templates": {
    "python": { "command": "python -m venv .venv && pip install pytest" },
    "django": { "command": "uv init && uv pip install django && django-admin startproject app ." },
    "svelte": { "command": "bunx create-vite . --template svelte-ts" }
  }
}
```

## 🤖 AI descriptions

riff uses an LLM CLI tool to auto-generate short project descriptions. It works out of the box — no config needed — and degrades gracefully if nothing is installed (you just won't get descriptions).

### Supported providers

| Provider | Binary | Detection priority |
|---|---|---|
| [Claude Code](https://docs.anthropic.com/en/docs/claude-code) | `claude` | 1st (preferred) |
| [GitHub Copilot CLI](https://docs.github.com/en/copilot/github-copilot-in-the-cli) | `copilot` | 2nd |

riff auto-detects whichever is available in your `$PATH`. If both are installed, Claude Code wins.

To pin a specific provider, set [`ai_provider`](#config-options) in your config.

### No provider installed?

No problem. riff works fine without one — descriptions are simply skipped.

## 🎨 Templates

riff ships with built-in templates that Just Work™:

| Name     | Command                                                  |
| -------- | -------------------------------------------------------- |
| `bun`    | `bun init -y`                                            |
| `dotnet` | `dotnet new console`                                     |
| `react`  | `bunx create-vite . --template react-ts`                 |
| `python` | `uv init`                                                |
| `rust`   | `cargo init .`                                           |
| `node`   | `npm init -y`                                            |
| `go`     | `go mod init temp`                                       |
| `next`   | `bunx create-next-app . --ts --eslint --app --use-bun`   |

### Custom templates

Override built-ins or add your own via the `templates` key in `~/.riff/config.json` (see [Configuration](#%EF%B8%8F-configuration)):

```json
{
  "templates": {
    "python": { "command": "python -m venv .venv && pip install pytest" },
    "django": { "command": "uv init && uv pip install django && django-admin startproject app ." },
    "svelte": { "command": "bunx create-vite . --template svelte-ts" }
  }
}
```

User entries with the same name as a built-in override it. Your config, your rules. 🤷

## 🐚 Shell integration

`riff new` and `riff open` automatically `cd` into the project directory — but only if you set up the shell hook. Add one line to your shell config:

### Bash / Zsh

Add to `~/.bashrc` or `~/.zshrc`:

```sh
eval "$(riff init)"
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
riff init fish | source
```

## 🤝 Contributing

Want to help make riff better? Excellent taste. See [CONTRIBUTING.md](CONTRIBUTING.md) for setup instructions and guidelines.

## 📄 License

[MIT](LICENSE) — go wild. 🎉
