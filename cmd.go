package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

type Command struct {
	cmd *cobra.Command
}

var rootCmd = &cobra.Command{
	Use:   "sshvault",
	Short: "SSH Vault is a CLI for managing your SSH connections",
	Long:  `SSH Vault is a usefull CLI for managing your SSH connections. You can store and run your SSH keys from a central location.`,
}

func NewCommand(storage *Storage) *Command {
	cmd := &Command{
		cmd: rootCmd,
	}
	cmd.newAddCmd(storage)
	cmd.newListCmd(storage)
	cmd.newDeleteCmd(storage)
	cmd.newConnectCmd(storage)
	return cmd
}

func (c *Command) newAddCmd(storage *Storage) {
	var name, connectionCmd string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an SSH connection",
		Run: func(cmd *cobra.Command, args []string) {
			err := storage.create(name, connectionCmd)
			if err != nil {
				slog.Error("failed to create connection", "err", err)
				return
			}
			slog.Info("Connection successfully created", slog.String("name", name))
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of connection")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&connectionCmd, "cmd", "c", "", "Connection command")
	cmd.MarkFlagRequired("cmd")
	c.cmd.AddCommand(cmd)
}

func (c *Command) newUpdateCmd(storage *Storage) {
	var name, connectionCmd string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an SSH connection",
		Run: func(cmd *cobra.Command, args []string) {
			err := storage.create(name, connectionCmd)
			if err != nil {
				slog.Error("failed to create connection", "err", err)
				return
			}
			slog.Info("Connection successfully created", slog.String("name", name))
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of connection")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&connectionCmd, "cmd", "c", "", "Connection command")
	cmd.MarkFlagRequired("cmd")
	c.cmd.AddCommand(cmd)
}

func (c *Command) newListCmd(storage *Storage) {
	c.cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all SSH connections",
		Run: func(cmd *cobra.Command, args []string) {
			conns, err := storage.records()
			if err != nil {
				slog.Error("failed to list connections", "err", err)
				return
			}

			names := make([]string, 0, len(conns))
			for _, conn := range conns {
				names = append(names, conn.name)
			}
			maxNameLength := CalculateMaxNameLength(names)

			for _, conn := range conns {
				fmt.Println(RenderConnectionRow(conn.name, conn.connectionCmd, maxNameLength))
			}
		},
	})
}

func (c *Command) newDeleteCmd(storage *Storage) {
	var name string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an SSH connection",
		Run: func(cmd *cobra.Command, args []string) {
			if err := storage.delete(name); err != nil {
				slog.Error("failed to delete connection", "err", err)
				return
			}
			slog.Info("Connection successfully deleted", slog.String("name", name))
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of connaction")
	cmd.MarkFlagRequired("name")
	c.cmd.AddCommand(cmd)
}

func (c *Command) newConnectCmd(storage *Storage) {
	var name string

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to an SSH connection",
		Run: func(cmd *cobra.Command, args []string) {
			conn, err := storage.find(name)
			if err != nil {
				slog.Error("failed to find connection", "err", err)
				return
			}

			sshCmd := exec.Command("ssh", conn.connectionCmd)

			sshCmd.Stdin = os.Stdin
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr

			if err := sshCmd.Run(); err != nil {
				slog.Error("ssh failed", "err", err)
			}
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of connection")
	cmd.MarkFlagRequired("name")
	c.cmd.AddCommand(cmd)
}
