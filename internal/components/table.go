package components

import (
	"fmt"

	"github.com/hopstee/sshvault/internal/storage"
	"github.com/hopstee/sshvault/internal/utils"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func ConnectionsTable(conns []storage.Record, withStatus bool, statuses map[string]utils.PingStatus) string {
	t := table.New()
	if withStatus {
		t.Headers("Name", "Connection", "Auth Type", "Status")
	} else {
		t.Headers("Name", "Connection", "Auth Type")
	}

	t.Border(lipgloss.RoundedBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			base := lipgloss.NewStyle().
				Padding(0, 2)
			if row == -1 {
				// return base.Bold(true).Foreground(lipgloss.Color("61"))
				return base.Bold(true).Foreground(ColorAccent)
			}
			if col == 1 && row != -1 {
				// return base.Foreground(lipgloss.Color("239"))
				return base.Foreground(ColorMuted)
			}
			return lipgloss.NewStyle().
				Padding(0, 2).
				Foreground(ColorPrimary)
		})

	for _, conn := range conns {
		cmd := fmt.Sprintf("ssh %s@%s -p %d", conn.User, conn.Address, conn.Port)

		name := conn.Name
		authType := mutedText.Render("unknown")
		if conn.AuthType != "" {
			authType = string(conn.AuthType)
		}

		if withStatus {
			var statusFormattedTest string

			status := statuses[conn.ID]
			switch status {
			case utils.PingUp:
				statusText := fmt.Sprintf("%s %s", status, ArrowUp)
				statusFormattedTest = upStatusStyle.Render(statusText)
			case utils.PingDown:
				statusText := fmt.Sprintf("%s %s", status, ArrowDown)
				statusFormattedTest = downStatusStyle.Render(statusText)
			}

			t = t.Row(name, cmd, authType, statusFormattedTest)
		} else {
			t = t.Row(name, cmd, authType)
		}
	}

	return t.Render()
}
