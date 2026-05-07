package components

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary = lipgloss.AdaptiveColor{
		Light: "0",
		Dark:  "255",
	}
	ColorMuted = lipgloss.AdaptiveColor{
		Light: "240",
		Dark:  "245",
	}
	ColorAccent = lipgloss.AdaptiveColor{
		Light: "32",
		Dark:  "110",
	}
	ColorError = lipgloss.AdaptiveColor{
		Light: "9",
		Dark:  "203",
	}
	ColorSuccess = lipgloss.AdaptiveColor{
		Light: "10",
		Dark:  "114",
	}
)

var upStatusStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorSuccess)

var downStatusStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorError)

var mutedText = lipgloss.NewStyle().
	Foreground(ColorMuted)
