package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var nameStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("51"))

var connStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("240"))

func RenderConnectionRow(name string, cmd string, maxNameLength int) string {
	name = padString(name, maxNameLength)
	name = nameStyle.
		Width(maxNameLength + 2).
		Render("[" + name + "]")
	cmd = connStyle.Render(cmd)
	return fmt.Sprintf("%s %s", name, cmd)
}
func padString(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return fmt.Sprintf("%-*s", length, s)
}
