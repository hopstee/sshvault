package cmd

import (
	"errors"
	"log/slog"

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

type CreateParams struct {
	Name         string
	Address      string
	User         string
	Port         int
	WithPassword bool
	KeyPath      string
	UseAgent     bool
}

func (c *Command) addCmd() {
	p := CreateParams{}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an SSH connection",
		Run: func(cmd *cobra.Command, args []string) {
			var passwordKey, pathToKey string
			var authType storage.AuthType

			err := c.detectAuthType(&p, &passwordKey, &pathToKey, &authType)
			if err != nil {
				if err := c.selectAuthType(p.Name, &passwordKey, &pathToKey, &authType); err != nil {
					return
				}
			}

			err = c.storage.Create(storage.UpsertDto{
				Name:        p.Name,
				Address:     p.Address,
				User:        p.User,
				PathToKey:   pathToKey,
				PasswordKey: passwordKey,
				Port:        p.Port,
				AuthType:    authType,
			})
			if err != nil {
				slog.Error("failed to create connection", "err", err)
				return
			}
			slog.Info("Connection successfully created")
		},
	}

	c.setAddFlags(cmd, &p)
	c.Cmd.AddCommand(cmd)
}

func (c *Command) setAddFlags(cmd *cobra.Command, p *CreateParams) {
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
		"addr",
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
		"p",
		22,
		"Port",
	)

	cmd.Flags().BoolVar(
		&p.WithPassword,
		"with-pass",
		false,
		"Set password for SSH connection",
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
