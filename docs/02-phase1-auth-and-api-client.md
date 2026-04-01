# Phase 1a: Auth Module + API Client

## Goal

Implement `jotform auth login/logout/whoami` and the underlying Jotform REST API client.

---

## 1. API Client (`internal/api/client.go`)

```go
package api

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

const defaultBaseURL = "https://api.jotform.com"

type Client struct {
    APIKey  string
    BaseURL string
    http    *http.Client
}

func New(apiKey string) *Client {
    return &Client{
        APIKey:  apiKey,
        BaseURL: defaultBaseURL,
        http:    &http.Client{Timeout: 15 * time.Second},
    }
}

// get performs an authenticated GET request and decodes the JSON body.
func (c *Client) get(path string, out any) error {
    url := fmt.Sprintf("%s%s?apiKey=%s", c.BaseURL, path, c.APIKey)
    resp, err := c.http.Get(url)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusUnauthorized {
        return fmt.Errorf("invalid API key — run `jotform auth login`")
    }
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    return json.NewDecoder(resp.Body).Decode(out)
}

// Jotform API response envelope
type apiResponse[T any] struct {
    ResponseCode int    `json:"responseCode"`
    Message      string `json:"message"`
    Content      T      `json:"content"`
}
```

---

## 2. Credential Storage (`internal/auth/keyring.go`)

```go
package auth

import (
    "errors"
    "github.com/99designs/keyring"
)

const (
    serviceName = "jotform-cli"
    keyName     = "api-key"
)

func SaveAPIKey(key string) error {
    ring, err := openKeyring()
    if err != nil {
        return err
    }
    return ring.Set(keyring.Item{Key: keyName, Data: []byte(key)})
}

func LoadAPIKey() (string, error) {
    ring, err := openKeyring()
    if err != nil {
        return "", err
    }
    item, err := ring.Get(keyName)
    if err != nil {
        if errors.Is(err, keyring.ErrKeyNotFound) {
            return "", fmt.Errorf("not logged in — run `jotform auth login`")
        }
        return "", err
    }
    return string(item.Data), nil
}

func DeleteAPIKey() error {
    ring, err := openKeyring()
    if err != nil {
        return err
    }
    return ring.Remove(keyName)
}

func openKeyring() (keyring.Keyring, error) {
    return keyring.Open(keyring.Config{
        ServiceName:              serviceName,
        KeychainTrustApplication: true,
    })
}
```

---

## 3. User Info API (`internal/api/user.go`)

```go
package api

type User struct {
    Username   string `json:"username"`
    Email      string `json:"email"`
    Name       string `json:"name"`
    AccountType string `json:"account_type"`
}

type APILimits struct {
    APIRequests    int `json:"api-requests"`
    APIRequestsMax int `json:"api-requests-max"`
}

func (c *Client) GetUser() (*User, error) {
    var resp apiResponse[User]
    if err := c.get("/user", &resp); err != nil {
        return nil, err
    }
    return &resp.Content, nil
}
```

---

## 4. Auth Command (`cmd/auth.go`)

```go
package cmd

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/jotform/jotform-cli/internal/api"
    "github.com/jotform/jotform-cli/internal/auth"
    "github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
    Use:   "auth",
    Short: "Manage authentication credentials",
}

var loginCmd = &cobra.Command{
    Use:   "login",
    Short: "Store your Jotform API key securely",
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Print("Enter your Jotform API key: ")
        reader := bufio.NewReader(os.Stdin)
        key, _ := reader.ReadString('\n')
        key = strings.TrimSpace(key)

        // Validate by calling the API
        client := api.New(key)
        user, err := client.GetUser()
        if err != nil {
            return fmt.Errorf("invalid API key: %w", err)
        }

        if err := auth.SaveAPIKey(key); err != nil {
            return err
        }
        fmt.Printf("Logged in as %s (%s)\n", user.Name, user.Email)
        return nil
    },
}

var logoutCmd = &cobra.Command{
    Use:   "logout",
    Short: "Remove stored credentials",
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := auth.DeleteAPIKey(); err != nil {
            return err
        }
        fmt.Println("Logged out.")
        return nil
    },
}

var whoamiCmd = &cobra.Command{
    Use:   "whoami",
    Short: "Show current authenticated user",
    RunE: func(cmd *cobra.Command, args []string) error {
        key, err := auth.LoadAPIKey()
        if err != nil {
            return err
        }
        client := api.New(key)
        user, err := client.GetUser()
        if err != nil {
            return err
        }
        fmt.Printf("User:    %s\nEmail:   %s\nPlan:    %s\n", user.Username, user.Email, user.AccountType)
        return nil
    },
}

func init() {
    authCmd.AddCommand(loginCmd, logoutCmd, whoamiCmd)
    rootCmd.AddCommand(authCmd)
}
```

---

## 5. Testing

```bash
# Manual smoke test
./jotform auth login
./jotform auth whoami
./jotform auth logout
./jotform auth whoami  # should error: "not logged in"
```

Unit test for the API client (`internal/api/client_test.go`):

```go
func TestGetUser_InvalidKey(t *testing.T) {
    client := api.New("invalid-key-000")
    _, err := client.GetUser()
    require.Error(t, err)
    assert.Contains(t, err.Error(), "invalid API key")
}
```

---

## Acceptance Criteria

- [ ] `jotform auth login` validates key and stores in keychain
- [ ] `jotform auth whoami` prints user info
- [ ] `jotform auth logout` removes the key
- [ ] Invalid key gives a clear error message
- [ ] `JOTFORM_API_KEY` env var bypasses keychain lookup

---

## Next Step

→ [03-phase1-forms-crud.md](03-phase1-forms-crud.md)
