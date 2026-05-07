package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/hopstee/sshvault/internal/storage"
	"golang.org/x/term"
)

var (
	ErrPasswordNotIdentical = errors.New("Passwords should be identical")
)

func (c *Command) detectAuthType(p *Params, passwordKey, pathToKey *string, authType *storage.AuthType) error {
	if p.WithPassword {
		*authType = storage.PasswordAuth
		*passwordKey = p.Name

		password, err := c.readPassword()
		if err != nil {
			slog.Error("failed read password", slog.Any("error", err))
		}

		if err := c.keyring.Set(p.Name, password); err != nil {
			return err
		}

		return nil
	}

	if p.KeyPath != "" {
		*authType = storage.KeyAuth
		*pathToKey = p.KeyPath
		return nil
	}

	if p.UseAgent {
		*authType = storage.AgentAuth
		return nil
	}

	slog.Warn("unknown auth type")
	return ErrEmptyAuthParams
}

func (c *Command) selectAuthType(p *Params, passwordKey, pathToKey *string, authType *storage.AuthType) error {
	fmt.Println("Select authentication method:")
	fmt.Println("1) Password")
	fmt.Println("2) SSH key")
	fmt.Println("3) SSH agent")
	fmt.Print("> ")

	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil {
		slog.Error("failed to read input", slog.Any("error", err))
		return err
	}

	switch choice {
	case 1:
		return c.processPasswordAuthSelection(p, passwordKey, authType)
	case 2:
		return c.processSSHKeyAuthSelection(pathToKey, authType)
	case 3:
		c.processSSHAgentAuthSelection(authType)
	default:
		slog.Error("invalid authentication method")
		return ErrInvalidAuthMethod
	}

	return nil
}

func (c *Command) processPasswordAuthSelection(p *Params, passwordKey *string, authType *storage.AuthType) error {
	*authType = storage.PasswordAuth

	password, err := c.readPassword()
	if err != nil {
		return err
	}

	if err := c.keyring.Set(p.Name, password); err != nil {
		return err
	}
	*passwordKey = p.Name

	return nil
}

func (c *Command) processSSHKeyAuthSelection(pathToKey *string, authType *storage.AuthType) error {
	*authType = storage.KeyAuth

	keys, err := c.listSSHKeys()
	if err != nil {
		return err
	}

	if err := c.readChoise(pathToKey, keys); err != nil {
		return err
	}

	return nil
}

func (c *Command) processSSHAgentAuthSelection(authType *storage.AuthType) {
	*authType = storage.AgentAuth
}

func (c *Command) listSSHKeys() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failed to get home dir", "err", err)
		return []string{}, err
	}

	sshDir := filepath.Join(home, ".ssh")

	entries, err := os.ReadDir(sshDir)
	if err != nil {
		slog.Error("failed to read ~/.ssh directory", "err", err)
		return []string{}, err
	}

	var keys []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if strings.HasSuffix(name, ".pub") {
			continue
		}

		switch name {
		case "config", "known_hosts", "known_hosts.old", "authorized_keys":
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		mode := info.Mode().Perm()
		if mode&0o077 != 0 {
			continue
		}

		fullPath := filepath.Join(sshDir, name)
		keys = append(keys, fullPath)
	}

	if len(keys) == 0 {
		slog.Error("no SSH keys found in ~/.ssh")
		return []string{}, ErrNoSSHKeys
	}

	fmt.Println("Detected SSH keys:")

	for i, key := range keys {
		fmt.Printf("%d) %s\n", i+1, key)
	}

	fmt.Printf("%d) Custom path\n", len(keys)+1)
	fmt.Print("> ")

	return keys, nil
}

func (c *Command) readChoise(pathToKey *string, keys []string) error {
	var keyChoice int
	_, err := fmt.Scanln(&keyChoice)
	if err != nil {
		slog.Error("failed to read input", "err", err)
		return err
	}

	if keyChoice >= 1 && keyChoice <= len(keys) {
		*pathToKey = keys[keyChoice-1]
	} else if keyChoice == len(keys)+1 {
		fmt.Print("Enter custom key path: ")
		_, err = fmt.Scanln(&pathToKey)
		if err != nil {
			slog.Error("failed to read key path", "err", err)
			return err
		}
	} else {
		slog.Error("invalid selection")
		return ErrInvalidSelection
	}

	return nil
}

func (c *Command) readPassword() (string, error) {
	fmt.Print("Enter password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		slog.Error("failed to read password", slog.Any("error", err))
		return "", err
	}
	if string(passwordBytes) == "" {
		slog.Error("password cannot be empty")
		return "", ErrEmptyPassword
	}

	fmt.Print("\nConfirm password: ")
	confirmPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		slog.Error("failed to read password", slog.Any("error", err))
		return "", err
	}
	fmt.Println()

	if string(passwordBytes) != string(confirmPasswordBytes) {
		slog.Error("passwords does not match", slog.Any("error", err))
		return "", ErrPasswordNotIdentical
	}

	return string(confirmPasswordBytes), nil
}
