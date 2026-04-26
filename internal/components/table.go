package components

import (
	"fmt"

	"github.com/hopstee/sshvault/internal/storage"
	"github.com/hopstee/sshvault/internal/utils"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var upStatusStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorSuccess)

var downStatusStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorError)

func ConnectionsTable(conns []storage.Record, withStatus bool, statuses map[string]utils.PingStatus) string {
	t := table.New()
	if withStatus {
		t.Headers("Name", "Connection", "Status")
	} else {
		t.Headers("Name", "Connection")
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
		if withStatus {
			status := statuses[conn.ID]
			switch status {
			case utils.PingUp:
				t = t.Row(conn.Name, cmd, upStatusStyle.Render(string(status)))
			case utils.PingDown:
				t = t.Row(conn.Name, cmd, downStatusStyle.Render(string(status)))
			default:
				t = t.Row(conn.Name, cmd, string(status))
			}
		} else {
			t = t.Row(conn.Name, cmd)
		}
	}

	return t.Render()
}
