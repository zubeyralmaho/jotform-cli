# Phase 0: Project Setup

## Prerequisites

- Go 1.22+
- A Jotform account with an API key

---

## 1. Initialize the Go Module

```bash
mkdir jotform-cli && cd jotform-cli
git init
go mod init github.com/zubeyralmaho/jotform-cli
```

---

## 2. Install Core Dependencies

```bash
# CLI framework
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest

# Secure credential storage
go get github.com/99designs/keyring@latest

# Terminal output styling
go get github.com/charmbracelet/lipgloss@latest

# Table output
go get github.com/olekukonko/tablewriter@latest

# Testing
go get github.com/stretchr/testify@latest
```

---

## 3. Create the Entrypoint

**main.go**
```go
package main

import "github.com/zubeyralmaho/jotform-cli/cmd"

func main() {
    cmd.Execute()
}
```

---

## 4. Root Command

**cmd/root.go**
```go
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
    Use:   "jotform",
    Short: "Jotform CLI — AI-native data collection at the terminal",
    Long: `Jotform CLI lets developers and AI agents create, manage,
and stream Jotform data directly from the terminal or CI/CD pipelines.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/jotform/config.yaml)")
    rootCmd.PersistentFlags().String("api-key", "", "Jotform API key (overrides keychain)")
    rootCmd.PersistentFlags().String("output", "table", "Output format: table | json | yaml")
    viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
    viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, _ := os.UserHomeDir()
        viper.AddConfigPath(home + "/.config/jotform")
        viper.SetConfigName("config")
        viper.SetConfigType("yaml")
    }
    viper.SetEnvPrefix("JOTFORM")
    viper.AutomaticEnv()
    viper.ReadInConfig()
}
```

---

## 5. Scaffold the Directory Structure

```bash
mkdir -p cmd internal/{api,auth,ai,mcp,output} docs
touch cmd/{auth,forms,submissions,ai,mcp}.go
touch internal/api/{client,forms,submissions}.go
touch internal/auth/keyring.go
touch internal/ai/generator.go
touch internal/mcp/server.go
touch internal/output/formatter.go
```

---

## 6. Verify the Build

```bash
go build -o jotform .
./jotform --help
```

Expected output:
```
Jotform CLI — AI-native data collection at the terminal

Usage:
  jotform [command]

Available Commands:
  auth        Manage authentication credentials
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
      --api-key string   Jotform API key (overrides keychain)
      --config string    config file (default: ~/.config/jotform/config.yaml)
  -h, --help             help for jotform
      --output string    Output format: table | json | yaml (default "table")
```

---

## Next Step

→ [02-phase1-auth-and-api-client.md](02-phase1-auth-and-api-client.md)
