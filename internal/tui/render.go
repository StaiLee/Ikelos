package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderGradient paints text across a colour ramp, one rune at a time. It
// indexes by rune position (not byte) so multi-byte characters render correctly.
func renderGradient(text string, colors []string) string {
	runes := []rune(text)
	if len(runes) == 0 || len(colors) == 0 {
		return text
	}
	var s strings.Builder
	for i, r := range runes {
		idx := 0
		if len(runes) > 1 {
			idx = i * (len(colors) - 1) / (len(runes) - 1)
		}
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colors[idx])).Render(string(r)))
	}
	return s.String()
}

// renderBar draws a proportional progress bar of the given cell width.
func renderBar(width int, pct float64, color lipgloss.Color) string {
	if width < 0 {
		width = 0
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := int(float64(width) * pct)
	if filled > width {
		filled = width
	}
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("━", filled)) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#333")).Render(strings.Repeat("┄", width-filled))
}

// truncate shortens s to at most max runes, appending an ellipsis. It is
// rune-safe and never panics, even for max <= 0 or width smaller than the
// ellipsis.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}

// truncateLeft keeps the tail of s (useful for long URLs where the end matters).
func truncateLeft(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[len(runes)-max:])
	}
	return "…" + string(runes[len(runes)-(max-1):])
}
