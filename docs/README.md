# Jotform CLI — Documentation Index

> For CLI usage and installation, see the main [README.md](../README.md).
> For current architecture, see [00-architecture.md](00-architecture.md).

## Reference

| Doc | Content |
|---|---|
| [00-architecture.md](00-architecture.md) | Tech stack, project layout, command tree, data flow |
| [design/REFACTOR-DESIGN.md](design/REFACTOR-DESIGN.md) | TUI redesign specification (design reference, partially implemented) |
| [external/jotform-api-go/README.md](external/jotform-api-go/README.md) | Mirrored upstream API method list (`v2/JotForm.go`) for local reference |

## Implementation Phases (Historical)

The documents below were written as build guides during initial development. They remain useful as implementation reference but may not reflect the current state of the codebase.

| Doc | Phase | Content |
|---|---|---|
| [01-project-setup.md](01-project-setup.md) | Setup | Go module init, Cobra root command, directory scaffold |
| [02-phase1-auth-and-api-client.md](02-phase1-auth-and-api-client.md) | Phase 1 | API client, keychain auth, `jotform auth` commands |
| [03-phase1-forms-crud.md](03-phase1-forms-crud.md) | Phase 1 | Forms CRUD, output formatter, submissions watch |
| [04-phase2-ai-module.md](04-phase2-ai-module.md) | Phase 2 | Claude API integration, `generate-schema`, `analyze` |
| [05-phase3-mcp-server.md](05-phase3-mcp-server.md) | Phase 3 | MCP server tools, Claude Desktop/Code integration |
| [06-phase4-release.md](06-phase4-release.md) | Phase 4 | GoReleaser, GitHub Actions, Homebrew tap |
| [07-testing-strategy.md](07-testing-strategy.md) | All | Unit, integration, E2E test patterns |

### Original Build Order

```
Setup → Phase 1a (Auth) → Phase 1b (Forms) → Phase 2 (AI) → Phase 3 (MCP) → Phase 4 (Release)
```
