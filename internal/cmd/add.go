package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/hopstee/sshvault/internal/storage"
	"github.com/spf13/cobra"
)

var (
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrNoSSHKeys         = errors.New("no SSH keys found in ~/.ssh")
	ErrInvalidSelection  = errors.New("invalid auth type selection")
	ErrInvalidAuthMethod = errors.New("invalid auth type method")
	ErrEmptyAuthParams   = errors.New("no one auth params not provided")
)

type params struct {
	Name      string
	Address   string
	User      string
	Password  string
	Port      int
	StorePass bool
	KeyPath   string
	UseAgent  bool
}

func (c *Command) addCmd() {
	p := params{}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an SSH connection",
		Run: func(cmd *cobra.Command, args []string) {
			var passwordKey, pathToKey string
			var authType storage.AuthType

			err := c.detectAuthType(&p, &passwordKey, &pathToKey, &authType)
			if err != nil {
				if err := c.selectAuthType(&p, &passwordKey, &pathToKey, &authType); err != nil {
					return
				}
			}

			err = c.storage.Create(p.Name, p.Address, p.User, pathToKey, passwordKey, p.Port, authType)
			if err != nil {
				slog.Error("failed to create connection", "err", err)
				return
			}
			slog.Info("Connection successfully created")
		},
	}

	c.setFlags(cmd, &p)
	c.Cmd.AddCommand(cmd)
}

func (c *Command) setFlags(cmd *cobra.Command, p *params) {
	cmd.Flags().StringVarP(
		&p.Name,
		"name",
		"n",
		"",
		"Name of connection",
	)
	cmd.MarkFlagRequired("name")

	cmd.Flags().StringVarP(
		&p.Address,
		"address",
		"a",
		"",
		"IP address",
	)
	cmd.MarkFlagRequired("address")

	cmd.Flags().StringVarP(
		&p.User,
		"user",
		"u",
		"root",
		"Username",
	)

	cmd.Flags().IntVarP(
		&p.Port,
		"port",
		"P",
		22,
		"Port",
	)

	cmd.Flags().StringVarP(
		&p.Password,
		"password",
		"p",
		"",
		"Password",
	)

	cmd.Flags().StringVar(
		&p.KeyPath,
		"key-path",
		"",
		"Path to private SSH key",
	)

	cmd.Flags().BoolVar(
		&p.UseAgent,
		"agent",
		false,
		"Use ssh-agent authentication",
	)
}

func (c *Command) detectAuthType(p *params, passwordKey, pathToKey *string, authType *storage.AuthType) error {
	if p.Password != "" {
		*authType = storage.PasswordAuth
		*passwordKey = p.Name

		if err := c.keyring.Set(p.Name, p.Password); err != nil {
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

	return ErrEmptyAuthParams
}

func (c *Command) selectAuthType(p *params, passwordKey, pathToKey *string, authType *storage.AuthType) error {
	fmt.Println("Select authentication method:")
	fmt.Println("1) Password")
	fmt.Println("2) SSH key")
	fmt.Println("3) SSH agent")
	fmt.Print("> ")

	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil {
		slog.Error("failed to read input", "err", err)
		return err
	}

	switch choice {
	case 1:
		return c.processPasswordAuthSelection(p, passwordKey, authType)
	case 2:
		return c.processSSHKeyAuthSElection(pathToKey, authType)
	case 3:
		c.processSSHAgentAuthSelection(authType)
	default:
		slog.Error("invalid authentication method")
		return ErrInvalidAuthMethod
	}

	return nil
}

func (c *Command) processPasswordAuthSelection(p *params, passwordKey *string, authType *storage.AuthType) error {
	*authType = storage.PasswordAuth

	if p.StorePass {
		fmt.Print("Enter password: ")
		_, err := fmt.Scanln(&p.Password)
		if err != nil {
			slog.Error("failed to read password", "err", err)
			return err
		}

		if p.Password == "" {
			slog.Error("password cannot be empty")
			return ErrEmptyPassword
		}

		if err := c.keyring.Set(p.Name, p.Password); err != nil {
			return err
		}
		*passwordKey = p.Name
	}

	return nil
}

func (c *Command) processSSHKeyAuthSElection(pathToKey *string, authType *storage.AuthType) error {
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
