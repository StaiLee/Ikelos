package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ikelos/internal/engine"
)

type helpModel struct {
	pages         []string
	pageIdx       int
	width, height int
	quitting      bool
}

// NewHelp returns the interactive manual model (shown when no URL is given).
func NewHelp() tea.Model { return helpModel{pages: buildPages()} }

func (m helpModel) Init() tea.Cmd { return nil }

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "right", "l", "enter", " ":
			if m.pageIdx < len(m.pages)-1 {
				m.pageIdx++
			}
		case "left", "h", "backspace":
			if m.pageIdx > 0 {
				m.pageIdx--
			}
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	}
	return m, nil
}

func (m helpModel) View() string {
	if m.quitting {
		return ""
	}
	navStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555"))
	activeDot := lipgloss.NewStyle().Foreground(ThemeMirror.Primary).Render("●")
	inactiveDot := lipgloss.NewStyle().Foreground(lipgloss.Color("#333")).Render("○")

	dots := ""
	for i := range m.pages {
		if i == m.pageIdx {
			dots += activeDot + " "
		} else {
			dots += inactiveDot + " "
		}
	}
	nav := fmt.Sprintf("\n%s\n%s", dots, navStyle.Render("[ARROWS] NEXT PAGE  •  [Q] EXIT SYSTEM"))

	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ThemeMirror.Primary).
		Padding(1, 3).
		Width(90).
		Align(lipgloss.Center)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, frame.Render(m.pages[m.pageIdx]), nav),
	)
}

func buildPages() []string {
	t := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFF")).Background(ThemeMirror.Secondary).Padding(0, 1).Render
	h := lipgloss.NewStyle().Bold(true).Foreground(ThemeMirror.Primary).MarginTop(1).Render
	c := lipgloss.NewStyle().Foreground(lipgloss.Color("#0F0")).Background(lipgloss.Color("#222")).Padding(0, 1).Render
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render

	p1 := fmt.Sprintf(`
%s

%s v%s
THE SURGEON

%s
%s

"We do not just copy. We absorb."
`, renderGradient("/// TOTAL REALITY SHIFTING ENGINE ///", ThemeMirror.Gradient), engine.Name, engine.Version, dim("Classified Tool"), dim("Systems Optimal."))

	p2 := fmt.Sprintf(`
%s

Ikelos performs intelligent structural analysis of a target.

%s
1. %s : Scans robots.txt & sitemap.xml for target mapping.
2. %s : Analyses DOM, detects lazy-load & srcset images.
3. %s : Routes traffic via proxies (if enabled) with rotating UAs.
4. %s : Hashing engine that rewrites filenames to prevent OS errors.

%s
Bounded worker pool, context cancellation, real error accounting.
`, t("SYSTEM ARCHITECTURE"), h("WORKFLOW"), c("The Hunter"), c("The Brain"), c("The Ghost"), c("The Vault"), h("CAPABILITIES"))

	p3 := fmt.Sprintf(`
%s

%s
%s
Target: High-fidelity clone.
Specs:  Balanced recursion, sitemap check active.

%s
%s
Target: WAF-protected / stealth.
Specs:  Slow, jitter, human emulation. Pair with a proxy.

%s
%s
Target: Bulk download.
Specs:  Max threads, short timeouts, retry on transient errors.
`, t("TACTICAL PROFILES"), h("1. MIRROR (Standard)"), c("-mode mirror"), h("2. SHADOW (Stealth)"), c("-mode shadow"), h("3. BLITZ (Speed)"), c("-mode blitz"))

	p4 := fmt.Sprintf(`
%s

%s
Target URL (http/https).

%s
Output directory for the clone.

%s
Proxy URL (HTTP/SOCKS5). Essential for "Ghost" operations.

%s
Operational profile: mirror, shadow, blitz.

%s
Manual User-Agent override (bypasses rotation).
`, t("COMMAND FLAGS (1/2)"), c("-url <URL>"), c("-out <DIR>"), c("-proxy <URL>"), c("-mode <MODE>"), c("-ua <STRING>"))

	p5 := fmt.Sprintf(`
%s

%s
Recursion depth (links to follow).

%s
Concurrent workers.

%s
Disable sitemap hunting / ignore robots.txt.

%s
%s
Press 'P' during a scan to freeze/resume the engine.
Press 'Q' to abort cleanly (in-flight requests are cancelled).
`, t("COMMAND FLAGS (2/2) & CONTROLS"), c("-depth <N>"), c("-threads <N>"), c("-nositemap / -ignore-robots"), h("RUNTIME CONTROLS"), c("[P] PAUSE / RESUME"))

	return []string{p1, p2, p3, p4, p5}
}
