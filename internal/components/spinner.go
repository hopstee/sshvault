package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var spinnerStyle = lipgloss.NewStyle().
	Foreground(ColorAccent)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func Spinner(done <-chan bool) {
	for {
		select {
		case <-done:
			fmt.Print("\r" + strings.Repeat(" ", 40) + "\r")
			return
		default:
			for _, frame := range spinnerFrames {
				fmt.Printf("\r%s Loading...", spinnerStyle.Render(frame))
				time.Sleep(80 * time.Millisecond)
			}
		}
	}
}
