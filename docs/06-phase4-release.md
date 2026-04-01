# Phase 4: Release & Distribution

## Goal

Cross-platform binary distribution via GitHub Releases + Homebrew tap.

---

## 1. Version Embedding

Inject version at build time using `ldflags`:

**cmd/version.go**
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var (
    Version   = "dev"
    Commit    = "none"
    BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("jotform %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

Build with version info:
```bash
go build -ldflags="-X 'github.com/zubeyralmaho/jotform-cli/cmd.Version=v1.0.0' \
  -X 'github.com/zubeyralmaho/jotform-cli/cmd.Commit=$(git rev-parse --short HEAD)' \
  -X 'github.com/zubeyralmaho/jotform-cli/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
  -o jotform .
```

---

## 2. GoReleaser Configuration

**.goreleaser.yaml**
```yaml
version: 2

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/zubeyralmaho/jotform-cli/cmd.Version={{.Version}}
      - -X github.com/zubeyralmaho/jotform-cli/cmd.Commit={{.Commit}}
      - -X github.com/zubeyralmaho/jotform-cli/cmd.BuildDate={{.Date}}

archives:
  - format: tar.gz
    name_template: "jotform_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

brews:
  - name: jotform
    repository:
      owner: jotform
      name: homebrew-tap
    description: "AI-native Jotform CLI for developers and AI agents"
    homepage: "https://github.com/zubeyralmaho/jotform-cli"
    install: |
      bin.install "jotform"
    test: |
      system "#{bin}/jotform version"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"
      - Merge pull request
```

---

## 3. GitHub Actions CI/CD

**.github/workflows/release.yml**
```yaml
name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

**.github/workflows/ci.yml**
```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - run: go test ./... -race -coverprofile=coverage.out
      - run: go vet ./...
      - uses: golangci/golangci-lint-action@v4
```

---

## 4. Release Checklist

```bash
# 1. Ensure all tests pass
go test ./... -race

# 2. Update CHANGELOG.md

# 3. Tag the release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GoReleaser will automatically:
# - Build for all platforms
# - Create GitHub Release with binaries
# - Update Homebrew tap
```

---

## 5. Installation Instructions (post-release)

**Homebrew (macOS/Linux)**
```bash
brew install jotform/tap/jotform
```

**Direct download (any platform)**
```bash
curl -sSL https://github.com/zubeyralmaho/jotform-cli/releases/latest/download/install.sh | sh
```

**Go install**
```bash
go install github.com/zubeyralmaho/jotform-cli@latest
```

---

## 6. Shell Completions

Cobra generates completions automatically:

```bash
# Bash
jotform completion bash > /etc/bash_completion.d/jotform

# Zsh
jotform completion zsh > "${fpath[1]}/_jotform"

# Fish
jotform completion fish > ~/.config/fish/completions/jotform.fish
```

Add to GoReleaser to include completion files in the archive.

---

## Acceptance Criteria

- [ ] `goreleaser release --snapshot --clean` builds all 6 platform binaries
- [ ] GitHub Release is created with binaries + checksums
- [ ] `brew install jotform/tap/jotform` works after tag push
- [ ] `jotform version` prints correct version/commit/date
- [ ] CI passes on every PR
