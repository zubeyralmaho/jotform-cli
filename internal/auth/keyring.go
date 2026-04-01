package auth

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/99designs/keyring"
)

const (
	serviceName = "jotform-cli"
	keyName     = "api-key"
)

func SaveAPIKey(key string) error {
	if runtime.GOOS == "darwin" {
		return saveAPIKeyDarwin(key)
	}

	ring, err := keyring.Open(keyring.Config{ServiceName: serviceName})
	if err != nil {
		return err
	}
	return ring.Set(keyring.Item{Key: keyName, Data: []byte(key)})
}

func LoadAPIKey() (string, error) {
	if runtime.GOOS == "darwin" {
		return loadAPIKeyDarwin()
	}

	ring, err := keyring.Open(keyring.Config{ServiceName: serviceName})
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
	if runtime.GOOS == "darwin" {
		return deleteAPIKeyDarwin()
	}

	ring, err := keyring.Open(keyring.Config{ServiceName: serviceName})
	if err != nil {
		return err
	}
	return ring.Remove(keyName)
}

func saveAPIKeyDarwin(key string) error {
	args := []string{"add-generic-password", "-U", "-s", serviceName, "-a", keyName, "-w", key}
	if exe, err := os.Executable(); err == nil && exe != "" {
		args = append(args, "-T", exe)
	}

	cmd := exec.Command("security", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("failed to save API key to macOS keychain: %s", msg)
	}
	return nil
}

func loadAPIKeyDarwin() (string, error) {
	cmd := exec.Command("security", "find-generic-password", "-w", "-s", serviceName, "-a", keyName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if strings.Contains(strings.ToLower(msg), "could not be found") {
			return "", fmt.Errorf("not logged in — run `jotform auth login`")
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("failed to load API key from macOS keychain: %s", msg)
	}
	key := strings.TrimSpace(stdout.String())

	// Re-save with the current executable trusted to migrate old ACL entries.
	// This is best-effort and should not block reads if the update fails.
	_ = saveAPIKeyDarwin(key)

	return key, nil
}

func deleteAPIKeyDarwin() error {
	cmd := exec.Command("security", "delete-generic-password", "-s", serviceName, "-a", keyName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if strings.Contains(strings.ToLower(msg), "could not be found") {
			return nil
		}
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("failed to delete API key from macOS keychain: %s", msg)
	}
	return nil
}
