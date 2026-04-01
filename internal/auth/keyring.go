package auth

import (
	"errors"
	"fmt"

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
		KeychainTrustApplication: false,
	})
}
