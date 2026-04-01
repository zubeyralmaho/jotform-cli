<p align="center">
	<img src="assets/jotform-logo.png" alt="Jotform CLI logo" width="760" />
</p>

# Jotform CLI

Form-as-Code CLI for creating, managing, and inspecting Jotform forms from the terminal.

## Why It Exists

Jotform CLI is built first as a tool surface for AI agents.

- Give AI agents a reliable, scriptable way to work with Jotform
- Enable universal agentic workflows across terminal, CI/CD, and MCP hosts
- Keep humans and agents on the same command surface with the same outputs
- Turn forms into something that can be created, inspected, diffed, and deployed like code

## What It Does

- Manage authentication with system keychain or `--api-key`
- List, get, create, update, delete, diff, status, export, import, sync, and apply forms
- Initialize project context with `.jotform.yaml`
- Clone a form into a new project directory
- Generate form schemas with AI
- Open forms in a browser or use the built-in MCP server
- Show a branded terminal UI with logo and help screens

## Install

### Go install

```bash
go install github.com/zubeyralmaho/jotform-cli@latest
```

### Build from source

```bash
git clone https://github.com/zubeyralmaho/jotform-cli.git
cd jotform-cli
make build
```

## Quick Start

```bash
jotform auth login
jotform forms list
jotform init
jotform clone 242753193847060
```

If you already have a project context, you can work without repeating the form ID:

```bash
jotform status
jotform diff
jotform push
jotform pull
```

## Command Overview

### Core

| Command | Description |
|---|---|
| `jotform auth login` | Store your API key securely |
| `jotform auth logout` | Remove stored credentials |
| `jotform auth whoami` | Show the current account |
| `jotform forms list` | List all forms |
| `jotform forms get [form-id]` | Fetch a form |
| `jotform forms create --file <file>` | Create a form from JSON or YAML |
| `jotform forms update [form-id] --file <file>` | Update a form from a local file |
| `jotform forms delete [form-id]` | Delete a form |
| `jotform forms sync` | Download all forms to `~/.jotform/` |
| `jotform forms export [form-id]` | Export a form to a local file |
| `jotform forms import` | Import a local form file |
| `jotform forms diff [form-id]` | Compare local and remote form state |
| `jotform forms status [form-id]` | Show local vs remote differences |
| `jotform forms apply [form-id]` | Apply local changes to a remote form |
| `jotform submissions list [form-id]` | List recent submissions |
| `jotform submissions watch [form-id]` | Stream new submissions |

### Workflow

| Command | Description |
|---|---|
| `jotform init` | Create `.jotform.yaml` project context |
| `jotform clone [form-id]` | Clone a form into a new directory |
| `jotform open [form-id]` | Open a form in the browser |
| `jotform status` | Show local and remote differences |
| `jotform diff` | Compare local schema with the remote form |
| `jotform push` | Apply local changes |
| `jotform pull` | Download the latest remote form |

### AI and MCP

| Command | Description |
|---|---|
| `jotform ai generate-schema "..."` | Generate a form schema from a prompt |
| `jotform ai analyze [form-id]` | Get AI suggestions for an existing form |
| `jotform mcp start-server` | Start the MCP server over stdio |

### Other

| Command | Description |
|---|---|
| `jotform dashboard` | Open the interactive dashboard |
| `jotform share` | Display the form URL and QR code |
| `jotform version` | Print version information |

## Configuration

Global flags:

```bash
--config   Path to config file (default: ~/.config/jotform/config.yaml)
--api-key  Jotform API key (overrides keychain)
--base-url Jotform API base URL
--output   Output format: table | json | yaml
```

Environment variables use the `JOTFORM_` prefix:

- `JOTFORM_API_KEY`
- `JOTFORM_BASE_URL`

## Project Context

`jotform init` creates a `.jotform.yaml` file that stores:

- `form_id`
- `name`
- `schema`

This lets commands like `jotform status`, `jotform diff`, `jotform push`, and `jotform pull` work without repeating the form ID or file path.

## Development

```bash
go test ./...
go run . --help
make build
make test
```

## Notes

- The CLI includes a branded terminal logo and help output.
- Shell completions are available via `jotform completion`.
- The MCP server is intended for Claude Desktop, Claude Code, and other MCP-compatible tools.