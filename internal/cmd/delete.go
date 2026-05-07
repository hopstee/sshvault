package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hopstee/sshvault/internal/storage"
	"github.com/spf13/cobra"
)

func (c *Command) deleteCmd() {
	c.Cmd.AddCommand(&cobra.Command{
		Use:                   "del [name]",
		Short:                 "Delete an SSH connection",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			// TODO: remove from known host if user provide special flag
			conn, err := c.storage.Find(name)
			if err != nil {
				slog.Error("connection not found")
				return
			}

			fmt.Printf("Delete %s connection? (Y/n): ", conn.Name)
			var deleteRes string
			_, err = fmt.Scanln(&deleteRes)
			if err != nil {
				slog.Error("failed to read your decision", slog.Any("error", err))
				return
			}
			if strings.ToLower(deleteRes) == "n" {
				slog.Info("deletion rejected")
				return
			}

			if err := c.storage.Delete(name); err != nil {
				slog.Error("failed to delete connection from store", slog.Any("error", err))
				return
			}

			if conn.AuthType == storage.PasswordAuth && conn.PasswordKey != "" {
				if err := c.keyring.Delete(conn.PasswordKey); err != nil {
					slog.Error("failed to delete connection from keyring", slog.Any("error", err))
					return
				}
			}
			slog.Info("Connection successfully deleted")
		},
	})
}
