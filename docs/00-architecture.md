# Architecture Overview

## Tech Stack

| Layer | Choice | Rationale |
|---|---|---|
| Language | Go 1.22+ | Single binary, zero deps, fast startup |
| CLI Framework | [Cobra](https://github.com/spf13/cobra) | Industry standard, sub-command tree |
| Config | [Viper](https://github.com/spf13/viper) | Multi-source config (env, file, flags) |
| Keychain | [99designs/keyring](https://github.com/99designs/keyring) | Cross-platform secure credential storage |
| HTTP Client | stdlib `net/http` | Sufficient; no need for extra dep |
| MCP | [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) | Go-native MCP server/client |
| Output | [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Styled terminal output |
| Testing | `testing` + [testify](https://github.com/stretchr/testify) | Standard Go test patterns |

---

## Project Layout

```
jotform-cli/
├── cmd/
│   ├── root.go          # Root command, global flags
│   ├── auth.go          # jotform auth {login,logout,whoami}
│   ├── forms.go         # jotform forms {list,get,create,sync}
│   ├── submissions.go   # jotform submissions {list,watch}
│   ├── ai.go            # jotform ai {generate-schema,analyze}
│   └── mcp.go           # jotform mcp {start-server}
├── internal/
│   ├── api/             # Jotform REST API client
│   │   ├── client.go
│   │   ├── forms.go
│   │   └── submissions.go
│   ├── auth/            # Credential management
│   │   └── keyring.go
│   ├── ai/              # LLM bridge (Claude API)
│   │   └── generator.go
│   ├── mcp/             # MCP server tools
│   │   └── server.go
│   └── output/          # Formatters (table, json, yaml)
│       └── formatter.go
├── docs/                # This directory
├── main.go
├── go.mod
└── go.sum
```

---

## Command Tree

```
jotform
├── auth
│   ├── login            # Store API key in system keychain
│   ├── logout           # Remove stored credentials
│   └── whoami           # Print current user + API usage
├── forms
│   ├── list             # List all forms (table/json/yaml)
│   ├── get [id]         # Fetch form JSON structure
│   ├── create --file    # Create form from local file
│   └── sync             # Pull remote → local .jotform/ dir
├── submissions
│   ├── list [form-id]   # Paginated submission list
│   └── watch [form-id]  # Long-poll → stdout (pipe-friendly)
├── ai
│   ├── generate-schema  # Prompt → Jotform JSON schema
│   └── analyze [id]     # Form → LLM UX improvement suggestions
└── mcp
    └── start-server     # Launch MCP server (stdio transport)
```

---

## Data Flow

```
User / AI Agent
     │
     ▼
  jotform CLI (Cobra)
     │
     ├─── internal/auth  ──▶ System Keychain
     │
     ├─── internal/api   ──▶ api.jotform.com/v1
     │
     ├─── internal/ai    ──▶ Anthropic API (Claude)
     │
     └─── internal/mcp   ──▶ MCP stdio transport
                              └─▶ Claude Desktop / Claude Code
```

---

## Configuration Precedence

1. CLI flags (highest priority)
2. Environment variables (`JOTFORM_API_KEY`, `JOTFORM_BASE_URL`)
3. Config file (`~/.config/jotform/config.yaml`)
4. System keychain (for sensitive credentials only)
