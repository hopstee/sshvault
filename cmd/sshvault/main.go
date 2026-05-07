package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hopstee/sshvault/internal/cmd"
	"github.com/hopstee/sshvault/internal/storage"
)

const APP_NAME = "sshvault"
const VERSION = "0.2.0"

func main() {
	storePath, err := getLocalStorePath(APP_NAME)
	if err != nil {
		slog.Error("failed to get local store path", "err", err)
		os.Exit(1)
	}

	keyring := storage.NewKeyring(APP_NAME)

	storage, err := storage.NewStorage(storePath)
	if err != nil {
		slog.Error("failed to init db", "err", err)
		os.Exit(1)
	}

	c := cmd.NewCommand(storage, keyring, VERSION)
	if err := c.Cmd.Execute(); err != nil {
		slog.Error("failed to execute command", "err", err)
		os.Exit(1)
	}
}

func getLocalStorePath(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, appName), nil
}
