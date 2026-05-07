package cmd

import (
	"fmt"
	"sync"

	"github.com/hopstee/sshvault/internal/storage"

	"github.com/spf13/cobra"
)

type Command struct {
	Cmd *cobra.Command
	wg  *sync.WaitGroup
	mu  *sync.Mutex

	storage *storage.Storage
	keyring *storage.Keyring
}

var rootCmd = &cobra.Command{
	Use:   "sshvault <command> [arguments]",
	Short: "SSH Vault is a CLI for managing your SSH connections",
	Long:  `SSH Vault is a usefull CLI for managing your SSH connections. You can store and run your SSH keys from a central location.`,
}

func NewCommand(s *storage.Storage, k *storage.Keyring, version string) *Command {
	cmd := &Command{
		Cmd: rootCmd,
		wg:  &sync.WaitGroup{},
		mu:  &sync.Mutex{},

		storage: s,
		keyring: k,
	}
	cmd.versionCmd(version)
	cmd.addCmd()
	cmd.listCmd()
	cmd.deleteCmd()
	cmd.connectCmd()
	return cmd
}

func (c *Command) versionCmd(version string) {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of SSH Vault",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
	c.Cmd.AddCommand(cmd)
}
