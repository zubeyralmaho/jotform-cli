package auth

import (
	"errors"
	"fmt"
	"sync"

	"github.com/99designs/keyring"
)

const (
	serviceName = "jotform-cli"
	keyName     = "api-key"
)

var (
	cachedKeyring keyring.Keyring
	keyringMu     sync.Mutex
)

// getKeyring returns a cached keyring instance to avoid multiple macOS permission prompts
func getKeyring() (keyring.Keyring, error) {
	keyringMu.Lock()
	defer keyringMu.Unlock()

	if cachedKeyring != nil {
		return cachedKeyring, nil
	}

	ring, err := keyring.Open(keyring.Config{
		ServiceName:              serviceName,
		KeychainTrustApplication: false,
	})
	if err != nil {
		return nil, err
	}

	cachedKeyring = ring
	return ring, nil
}

func SaveAPIKey(key string) error {
	ring, err := getKeyring()
	if err != nil {
		return err
	}
	return ring.Set(keyring.Item{Key: keyName, Data: []byte(key)})
}

func LoadAPIKey() (string, error) {
	ring, err := getKeyring()
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
	ring, err := getKeyring()
	if err != nil {
		return err
	}
	return ring.Remove(keyName)
}
