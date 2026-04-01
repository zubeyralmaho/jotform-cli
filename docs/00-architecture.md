# Architecture Overview

## Tech Stack

| Layer | Choice | Rationale |
|---|---|---|
| Language | Go 1.26.1 | Single binary, zero deps, fast startup |
| CLI Framework | [spf13/cobra](https://github.com/spf13/cobra) | Sub-command tree, auto-completion |
| Config | [spf13/viper](https://github.com/spf13/viper) | Multi-source config (env, file, flags) |
| Keychain | [99designs/keyring](https://github.com/99designs/keyring) | Cross-platform secure credential storage |
| HTTP Client | stdlib `net/http` | Sufficient; no extra dep |
| TUI Architecture | [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | Elm-style interactive TUI (login, dashboard, spinners) |
| TUI Components | [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) | Spinner, text input, viewport |
| TUI Styling | [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Brand color palette, styled terminal output |
| AI | [anthropics/anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) | Claude API for schema generation and analysis |
| MCP | [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) | MCP server for Claude Desktop / Claude Code |
| QR Code | [skip2/go-qrcode](https://github.com/skip2/go-qrcode) | Terminal QR code rendering |
| Table Output | [olekukonko/tablewriter](https://github.com/olekukonko/tablewriter) | Structured table formatting |
| Testing | `testing` + [stretchr/testify](https://github.com/stretchr/testify) | Standard Go test patterns |

---

## Project Layout

```
jotform-cli/
├── main.go              # Entry point — calls cmd.Execute()
├── go.mod
├── Makefile             # build, test, lint, install, dev
├── .goreleaser.yaml     # Cross-platform release (linux/darwin/windows)
├── cmd/
│   ├── root.go          # Root command, global flags, branded help
│   ├── auth.go          # auth {login,logout,whoami} — login is bubbletea TUI
│   ├── forms.go         # forms {list,get,create,update,delete,sync,export,import,diff,apply,status}
│   ├── submissions.go   # submissions {list,watch}
│   ├── ai.go            # ai {generate-schema,analyze}
│   ├── mcp.go           # mcp {start-server}
│   ├── init.go          # jotform init (interactive + non-interactive)
│   ├── clone.go         # jotform clone
│   ├── open.go          # jotform open (browser launch)
│   ├── dashboard.go     # jotform dashboard (full-screen bubbletea TUI)
│   ├── share.go         # jotform share (QR code + URL display)
│   ├── shortcuts.go     # Root-level shortcut aliases (ls, get, new, rm, etc.)
│   ├── helpers.go       # Shared helpers (newClient, resolveAPIKey, confirmPrompt)
│   └── version.go       # jotform version (logo + build info)
├── internal/
│   ├── api/             # Jotform REST client
│   │   ├── client.go    # HTTP client, apiResponse[T] envelope parsing
│   │   ├── forms.go     # ListForms, GetForm, CreateForm, UpdateForm, DeleteForm
│   │   ├── submissions.go # GetSubmissions
│   │   └── user.go      # GetUser
│   ├── auth/            # Keyring credential storage
│   │   └── keyring.go   # Save/Load/Delete API key from system keychain
│   ├── ai/              # Claude API bridge
│   │   ├── generator.go # GenerateSchema, AnalyzeForm
│   │   └── schema_contract.go # Jotform field type definitions for LLM
│   ├── config/          # Project context (.jotform.yaml)
│   │   └── context.go   # Load/Save/Resolve project config, walk-up search
│   ├── formcode/        # Form-as-Code operations
│   │   ├── codec.go     # Read/Write form files (JSON/YAML)
│   │   ├── diff.go      # Compute unified diffs between local and remote
│   │   ├── validator.go # Schema validation against Jotform constraints
│   │   └── status.go    # git-status-style change reports
│   ├── mcp/             # MCP server tools
│   │   └── server.go    # 6 tools: list/get/create/update/delete forms + generate_schema
│   ├── output/          # Formatters (table, json, yaml)
│   │   └── formatter.go # Print() dispatches to correct format
│   ├── ui/              # TUI layer (Jotform 2026 brand identity)
│   │   ├── theme.go     # Brand color palette as lipgloss tokens
│   │   ├── logo.go      # Geometric logo rendering with ANSI block chars
│   │   ├── components.go # Spinner, staggered list, banners, key-value formatting
│   │   └── help.go      # Custom branded help template (logo + command groups)
│   └── watch/           # Submission streaming
│       └── checkpoint.go # Cursor persistence (~/.jotform/watch-*.cursor)
├── docs/
└── assets/
```

---

## Command Tree

```
jotform
├── auth
│   ├── login              # Bubbletea TUI: masked input → spinner → success/error banner
│   ├── logout             # Remove credentials from keychain
│   └── whoami             # Spinner → branded key-value display
├── forms (f, form)
│   ├── list               # Spinner + staggered animated list
│   ├── get [id]           # Spinner → form structure
│   ├── create --file      # --skip-validation
│   ├── update [id] --file # --skip-validation, --dry-run
│   ├── delete [id]        # --force, --dry-run, confirmation prompt
│   ├── sync               # Bulk download to ~/.jotform/
│   ├── export [id]        # -o/--out
│   ├── import             # Alias for create
│   ├── diff [id]          # --file
│   ├── apply [id]         # --file, --skip-validation, --dry-run
│   └── status [id]        # --file, --summary
├── submissions (subs, sub)
│   ├── list [form-id]     # --limit
│   └── watch [form-id]    # --interval, --no-checkpoint
├── ai
│   ├── generate-schema    # --out, --model, --max-tokens, --timeout, --max-retries, --show-usage
│   └── analyze [id]       # Same flags as generate-schema
├── mcp
│   └── start-server       # MCP over stdio
├── init                   # --form-id, --new, --title, --schema (interactive default)
├── clone [id]             # --name, --force
├── open [id]              # Launch form in default browser
├── dashboard (dash, d)    # Full-screen split-pane TUI
├── share [id]             # QR code + form URL
├── version                # Logo + version/commit/date
├── completion             # bash, zsh, fish, powershell (auto-generated)
└── [root shortcuts]
    ├── login, logout, whoami
    ├── ls (list), get, new (create), rm (remove, delete)
    ├── pull (export), push, diff, status
    ├── watch
    └── generate (gen)
```

---

## Data Flow

```
User / AI Agent
     │
     ▼
  jotform CLI (Cobra + branded help)
     │
     ├─── internal/ui       ──▶ TUI layer (bubbletea, lipgloss, logo, spinners)
     │
     ├─── internal/config   ──▶ .jotform.yaml (project context, walk-up resolution)
     │
     ├─── internal/formcode ──▶ Schema I/O, validation, diff, status reports
     │
     ├─── internal/auth     ──▶ System Keychain (99designs/keyring)
     │
     ├─── internal/api      ──▶ api.jotform.com
     │
     ├─── internal/ai       ──▶ Anthropic API (Claude) via anthropic-sdk-go
     │
     ├─── internal/watch    ──▶ Checkpoint persistence (~/.jotform/watch-*.cursor)
     │
     └─── internal/mcp      ──▶ MCP stdio transport
                                  └─▶ Claude Desktop / Claude Code
```

---

## Configuration Precedence

1. CLI flags (highest priority)
2. Environment variables (`JOTFORM_API_KEY`, `JOTFORM_BASE_URL`, `ANTHROPIC_API_KEY`)
3. Config file (`~/.config/jotform/config.yaml`)
4. System keychain (for API key only)
