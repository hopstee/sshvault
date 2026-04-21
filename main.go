package main

import (
	"log/slog"
	"os"
	"path/filepath"
)

const APP_NAME = "sshvault"

func main() {
	storePath, err := getLocalStorePath(APP_NAME)
	if err != nil {
		slog.Error("failed to get local store path", "err", err)
		os.Exit(1)
	}

	storage, err := NewStorage(storePath)
	if err != nil {
		slog.Error("failed to init db", "err", err)
		os.Exit(1)
	}

	cmd := NewCommand(storage)
	if err := cmd.cmd.Execute(); err != nil {
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
