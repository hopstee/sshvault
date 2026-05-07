package storage

import (
	"log/slog"

	"github.com/zalando/go-keyring"
)

type Keyring struct {
	appName string
}

func NewKeyring(appName string) *Keyring {
	return &Keyring{
		appName: appName,
	}
}

func (k *Keyring) Set(connectionName, password string) error {
	if err := keyring.Set(k.appName, connectionName, password); err != nil {
		slog.Error("failed store password in system keyring")
		return err
	}
	return nil
}

func (k *Keyring) Get(connectionName string) (string, error) {
	password, err := keyring.Get(k.appName, connectionName)
	if err != nil {
		slog.Error("password not found in system keyring")
		return "", err
	}
	return password, nil
}

func (k *Keyring) Delete(connectionName string) error {
	if err := keyring.Delete(k.appName, connectionName); err != nil {
		slog.Error("failed delete password from system keyring")
		return err
	}
	return nil
}
