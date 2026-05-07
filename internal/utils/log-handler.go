package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type PrettyHandler struct {
	level slog.Level
}

func NewPrettyHandler(level slog.Level) *PrettyHandler {
	return &PrettyHandler{
		level: level,
	}
}

func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := formatLevel(r.Level)
	timestamp := time.Now().Format("15:04:05")

	msg := fmt.Sprintf(
		"%s[%s] %s",
		level,
		timestamp,
		r.Message,
	)

	var attrs []string

	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%s=%+v", a.Key, a.Value.Any()))
		return true
	})

	if len(attrs) > 0 {
		msg += " " + strings.Join(attrs, " ")
	}

	_, err := fmt.Fprintln(os.Stdout, msg)
	return err
}

func (h *PrettyHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *PrettyHandler) WithGroup(_ string) slog.Handler {
	return h
}

func formatLevel(level slog.Level) string {
	switch {
	case level >= 12:
		return colorize("FATA", "88", "160")

	case level >= slog.LevelError:
		return colorize("ERRO", "88", "160")

	case level >= slog.LevelWarn:
		return colorize("WARN", "214", "136")

	case level >= slog.LevelInfo:
		return colorize("INFO", "45", "26")

	default:
		return colorize("DEBG", "54", "92")
	}
}

func colorize(text, darkColor, lightColor string) string {
	text = strings.ToUpper(text)
	if len(text) > 4 {
		text = text[:4]
	}

	style := lipgloss.NewStyle().
		Foreground(
			lipgloss.AdaptiveColor{
				Dark:  darkColor,
				Light: lightColor,
			},
		)

	return style.Render(text)
}
