package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sshvault/internal/components"
	"sshvault/internal/storage"
	"sshvault/internal/utils"
	"strconv"
	"sync"

	"github.com/spf13/cobra"
)

type Command struct {
	Cmd *cobra.Command
	wg  *sync.WaitGroup
}

var rootCmd = &cobra.Command{
	Use:   "sshvault",
	Short: "SSH Vault is a CLI for managing your SSH connections",
	Long:  `SSH Vault is a usefull CLI for managing your SSH connections. You can store and run your SSH keys from a central location.`,
}

func NewCommand(s *storage.Storage) *Command {
	cmd := &Command{
		Cmd: rootCmd,
		wg:  &sync.WaitGroup{},
	}
	cmd.newAddCmd(s)
	cmd.newListCmd(s)
	cmd.newDeleteCmd(s)
	cmd.newConnectCmd(s)
	return cmd
}

func (c *Command) newAddCmd(s *storage.Storage) {
	var name, address, user string
	var port int

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an SSH connection",
		Run: func(cmd *cobra.Command, args []string) {
			err := s.Create(name, address, user, port)
			if err != nil {
				slog.Error("failed to create connection", "err", err)
				return
			}
			slog.Info("Connection successfully created")
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of connection")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&address, "address", "a", "", "IP address")
	cmd.MarkFlagRequired("address")
	cmd.Flags().StringVarP(&user, "user", "u", "root", "Username")
	cmd.Flags().IntVarP(&port, "port", "p", 22, "Port")
	c.Cmd.AddCommand(cmd)
}

func (c *Command) newListCmd(s *storage.Storage) {
	var withPing bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all SSH connections",
		Run: func(cmd *cobra.Command, args []string) {
			conns, err := s.Records()
			if err != nil {
				slog.Error("failed to list connections", "err", err)
				return
			}

			statuses := make(map[string]utils.PingStatus)
			done := make(chan bool)
			if withPing {
				c.wg.Go(func() {
					components.Spinner(done)
				})
				for _, conn := range conns {
					host := fmt.Sprintf("%s:%d", conn.Address, conn.Port)
					statuses[conn.ID] = utils.PingHost(host)
				}
				close(done)
				c.wg.Wait()
			}
			fmt.Println(components.ConnectionsTable(conns, withPing, statuses))
		},
	}

	cmd.Flags().BoolVarP(&withPing, "ping", "p", false, "Show ping time")
	c.Cmd.AddCommand(cmd)
}

func (c *Command) newDeleteCmd(s *storage.Storage) {
	var name string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an SSH connection",
		Run: func(cmd *cobra.Command, args []string) {
			if err := s.Delete(name); err != nil {
				slog.Error("failed to delete connection", "err", err)
				return
			}
			slog.Info("Connection successfully deleted")
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of connaction")
	cmd.MarkFlagRequired("name")
	c.Cmd.AddCommand(cmd)
}

func (c *Command) newConnectCmd(s *storage.Storage) {
	var name string

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to an SSH connection",
		Run: func(cmd *cobra.Command, args []string) {
			conn, err := s.Find(name)
			if err != nil {
				slog.Error("failed to find connection", "err", err)
				return
			}

			sshCmd := exec.Command("ssh", "-p", strconv.Itoa(conn.Port), conn.User, "@", conn.Address)

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
	c.Cmd.AddCommand(cmd)
}
