// Package tui renders Ikelos's Bubble Tea interfaces: the interactive manual
// and the live crawl dashboard. It depends on the engine for stats and control
// but the engine has no knowledge of the UI.
package tui

import (
	"github.com/charmbracelet/lipgloss"
	"ikelos/internal/engine"
)

// Theme is a coordinated colour set for one operational mode.
type Theme struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Dark      lipgloss.Color
	Text      lipgloss.Color
	Success   lipgloss.Color
	Error     lipgloss.Color
	Warning   lipgloss.Color
	Gradient  []string
}

var (
	ThemeBlitz = Theme{
		Primary:   lipgloss.Color("#FF3300"), // Magma Red
		Secondary: lipgloss.Color("#FF8800"),
		Dark:      lipgloss.Color("#1a0500"),
		Text:      lipgloss.Color("#FFCC00"),
		Success:   lipgloss.Color("#00FF00"),
		Error:     lipgloss.Color("#FF0000"),
		Warning:   lipgloss.Color("#FFFF00"),
		Gradient:  []string{"#FF0000", "#FF4400", "#FF8800"},
	}
	ThemeMirror = Theme{
		Primary:   lipgloss.Color("#00F0FF"), // Cyberpunk Cyan
		Secondary: lipgloss.Color("#7df9ff"),
		Dark:      lipgloss.Color("#000a14"),
		Text:      lipgloss.Color("#E0FFFF"),
		Success:   lipgloss.Color("#00FF99"),
		Error:     lipgloss.Color("#FF3333"),
		Warning:   lipgloss.Color("#FFD700"),
		Gradient:  []string{"#0099FF", "#00CCFF", "#00FFFF"},
	}
	ThemeShadow = Theme{
		Primary:   lipgloss.Color("#A020F0"), // Void Purple
		Secondary: lipgloss.Color("#D8BFD8"),
		Dark:      lipgloss.Color("#0a000a"),
		Text:      lipgloss.Color("#E6E6FA"),
		Success:   lipgloss.Color("#00FA9A"),
		Error:     lipgloss.Color("#FF4500"),
		Warning:   lipgloss.Color("#FFA500"),
		Gradient:  []string{"#4B0082", "#800080", "#9932CC"},
	}
)

// ThemeFor returns the palette for a given operational mode.
func ThemeFor(mode engine.Mode) Theme {
	switch mode {
	case engine.ModeBlitz:
		return ThemeBlitz
	case engine.ModeShadow:
		return ThemeShadow
	default:
		return ThemeMirror
	}
}
