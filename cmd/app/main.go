package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"sshvault/internal/cmd"
	"sshvault/internal/storage"
)

const APP_NAME = "sshvault"

func main() {
	storePath, err := getLocalStorePath(APP_NAME)
	if err != nil {
		slog.Error("failed to get local store path", "err", err)
		os.Exit(1)
	}

	storage, err := storage.NewStorage(storePath)
	if err != nil {
		slog.Error("failed to init db", "err", err)
		os.Exit(1)
	}

	c := cmd.NewCommand(storage)
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
