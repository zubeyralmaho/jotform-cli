# Jotform CLI — Documentation Index

## Build Phases

| Doc | Phase | Content |
|---|---|---|
| [00-architecture.md](00-architecture.md) | Setup | Tech stack, project layout, command tree, data flow |
| [01-project-setup.md](01-project-setup.md) | Setup | Go module init, Cobra root command, directory scaffold |
| [02-phase1-auth-and-api-client.md](02-phase1-auth-and-api-client.md) | Phase 1 | API client, keychain auth, `jotform auth` commands |
| [03-phase1-forms-crud.md](03-phase1-forms-crud.md) | Phase 1 | Forms CRUD, output formatter, submissions watch |
| [04-phase2-ai-module.md](04-phase2-ai-module.md) | Phase 2 | Claude API integration, `generate-schema`, `analyze` |
| [05-phase3-mcp-server.md](05-phase3-mcp-server.md) | Phase 3 | MCP server tools, Claude Desktop/Code integration |
| [06-phase4-release.md](06-phase4-release.md) | Phase 4 | GoReleaser, GitHub Actions, Homebrew tap |
| [07-testing-strategy.md](07-testing-strategy.md) | All | Unit, integration, E2E test patterns |

## Recommended Build Order

```
Setup → Phase 1a (Auth) → Phase 1b (Forms) → Phase 2 (AI) → Phase 3 (MCP) → Phase 4 (Release)
```

Each phase produces a working, shippable increment:
- **After Phase 1:** Usable as a developer CLI tool
- **After Phase 2:** AI-powered form generation
- **After Phase 3:** Claude Desktop / Claude Code native integration
- **After Phase 4:** Public open-source release
