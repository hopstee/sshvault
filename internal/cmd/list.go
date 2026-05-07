package cmd

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/hopstee/sshvault/internal/components"
	"github.com/hopstee/sshvault/internal/utils"
	"github.com/spf13/cobra"
)

func (c *Command) listCmd() {
	var withPing bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all SSH connections",
		Run: func(cmd *cobra.Command, args []string) {
			conns, err := c.storage.Records()
			if err != nil {
				slog.Error("failed to list connections", "err", err)
				return
			}

			statuses := make(map[string]utils.PingStatus)
			semaphore := make(chan struct{}, 15)
			done := make(chan bool)
			if withPing {
				spinnerWg := sync.WaitGroup{}
				spinnerWg.Go(func() {
					components.Spinner(done)
				})
				for _, conn := range conns {
					c.wg.Go(func() {
						semaphore <- struct{}{}
						defer func() { <-semaphore }()

						host := fmt.Sprintf("%s:%d", conn.Address, conn.Port)
						c.mu.Lock()
						statuses[conn.ID] = utils.PingHost(host)
						c.mu.Unlock()
					})
				}
				c.wg.Wait()
				close(done)
				spinnerWg.Wait()
			}
			fmt.Println(components.ConnectionsTable(conns, withPing, statuses))
		},
	}

	cmd.Flags().BoolVarP(&withPing, "ping", "p", false, "Show ping time")
	c.Cmd.AddCommand(cmd)
}
