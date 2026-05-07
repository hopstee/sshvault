package cmd

import (
	"log/slog"

	"github.com/hopstee/sshvault/internal/storage"
	"github.com/spf13/cobra"
)

type UpdateParams struct {
	NewName        string
	Address        string
	User           string
	Port           int
	ChangeAuthType bool
}

func (c *Command) updateCmd() {
	p := UpdateParams{}
	cmd := &cobra.Command{
		Use:                   "update [name]",
		Short:                 "Update an SSH connection",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			slog.Info("check args", slog.Any("list", args))
			oldName := args[0]
			if oldName == "" {
				slog.Error("connection name is required")
				return
			}

			var passwordKey, pathToKey string
			var authType storage.AuthType

			conn, err := c.storage.Find(oldName)
			slog.Info("check name", slog.String("name", oldName))
			if err != nil {
				slog.Error("failed find connection", slog.String("name", oldName))
				return
			}

			if p.ChangeAuthType {
				if err := c.selectAuthType(p.NewName, &passwordKey, &pathToKey, &authType); err != nil {
					return
				}
				if conn.AuthType == storage.PasswordAuth {
					if err := c.keyring.Delete(oldName); err != nil {
						slog.Error("failed to delete old record in keyring")
						return
					}
				}
			} else {
				authType = conn.AuthType
			}

			c.validateUpdateParams(&conn, &p)

			err = c.storage.Update(oldName, storage.UpsertDto{
				Name:        p.NewName,
				Address:     p.Address,
				User:        p.User,
				PathToKey:   pathToKey,
				PasswordKey: passwordKey,
				Port:        p.Port,
				AuthType:    authType,
			})
			if err != nil {
				slog.Error("failed to update connection", "err", err)
				return
			}
			slog.Info("Connection successfully updated")
		},
	}

	c.setUpdateFlags(cmd, &p)
	c.Cmd.AddCommand(cmd)
}

func (c *Command) validateUpdateParams(conn *storage.Record, params *UpdateParams) {
	if params.NewName == "" {
		params.NewName = conn.Name
	}

	if params.Address == "" {
		params.Address = conn.Address
	}

	if params.User == "" {
		params.User = conn.User
	}

	if params.Port == 0 {
		params.Port = conn.Port
	}
}

func (c *Command) setUpdateFlags(cmd *cobra.Command, p *UpdateParams) {
	cmd.Flags().StringVarP(
		&p.NewName,
		"name",
		"n",
		"",
		"New name of connection",
	)

	cmd.Flags().StringVarP(
		&p.Address,
		"addr",
		"a",
		"",
		"New IP address",
	)

	cmd.Flags().StringVarP(
		&p.User,
		"user",
		"u",
		"root",
		"New username",
	)

	cmd.Flags().IntVarP(
		&p.Port,
		"port",
		"p",
		22,
		"Port",
	)

	cmd.Flags().BoolVar(
		&p.ChangeAuthType,
		"change-auth",
		false,
		"Change auth type",
	)
}
