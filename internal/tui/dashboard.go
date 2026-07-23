package tui

import (
	"fmt"
	"runtime"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ikelos/internal/engine"
)

type tickMsg time.Time
type finishedMsg struct{}

type dashboard struct {
	viewport      viewport.Model
	spinner       spinner.Model
	width, height int
	ready         bool
	quitting      bool
	finished      bool
	paused        bool
	logBuffer     string

	startTime time.Time
	cfg       engine.Config
	theme     Theme
	eng       *engine.Engine

	stats      engine.Stats
	throughput float64
	sparkline  []int
	ramUsage   string
	goroutines int
	lastLine   string

	logs <-chan engine.LogEntry
	done chan bool
}

// NewDashboard builds the live crawl UI. done is signalled by the caller when
// the engine's Run returns.
func NewDashboard(cfg engine.Config, eng *engine.Engine, done chan bool) tea.Model {
	theme := ThemeFor(cfg.Mode)
	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	return &dashboard{
		spinner:   s,
		viewport:  viewport.New(0, 0),
		startTime: time.Now(),
		cfg:       cfg,
		theme:     theme,
		eng:       eng,
		sparkline: make([]int, 40),
		logs:      eng.Logs(),
		done:      done,
	}
}

func (m *dashboard) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitForLog(m.logs), waitForDone(m.done), tickCmd())
}

func waitForLog(sub <-chan engine.LogEntry) tea.Cmd {
	return func() tea.Msg { return <-sub }
}
func waitForDone(sub chan bool) tea.Cmd {
	return func() tea.Msg { <-sub; return finishedMsg{} }
}
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m *dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			m.eng.Cancel() // stop the crawl, not just the UI
			return m, tea.Quit
		case "p":
			m.paused = !m.paused
			m.eng.SetPaused(m.paused)
			if m.paused {
				m.appendLine(lipgloss.NewStyle().Foreground(m.theme.Warning).Render("\n[!] SYSTEM PAUSED BY USER [!]\n"))
			} else {
				m.appendLine(lipgloss.NewStyle().Foreground(m.theme.Success).Render("\n[>] SYSTEM RESUMED\n"))
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		h := m.height - 12
		if h < 10 {
			h = 10
		}
		w := int(float64(m.width) * 0.65)
		if w < 20 {
			w = 20
		}
		m.viewport.Width, m.viewport.Height = w, h
		m.ready = true
		return m, nil

	case spinner.TickMsg:
		if m.finished {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		m.refreshStats()
		return m, tickCmd()

	case engine.LogEntry:
		m.appendLog(msg)
		return m, waitForLog(m.logs)

	case finishedMsg:
		m.finished = true
		return m, nil
	}
	return m, nil
}

func (m *dashboard) refreshStats() {
	m.stats = m.eng.Snapshot()
	if elapsed := time.Since(m.startTime).Seconds(); elapsed > 0 {
		m.throughput = (float64(m.stats.Bytes) / 1024 / 1024) / elapsed
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	m.ramUsage = fmt.Sprintf("%d MB", mem.Alloc/1024/1024)
	m.goroutines = runtime.NumGoroutine()

	activity := m.stats.Active
	if activity > 10 {
		activity = 10
	}
	m.sparkline = append(m.sparkline[1:], activity)
}

func (m *dashboard) appendLog(e engine.LogEntry) {
	prefix, color := prefixFor(e.Level, m.theme)
	if e.Text == "" {
		return
	}
	m.lastLine = e.Text
	ts := lipgloss.NewStyle().Foreground(lipgloss.Color("#555")).Render(time.Now().Format("15:04:05"))
	pre := lipgloss.NewStyle().Foreground(color).Bold(true).Width(7).Render(prefix)
	limit := m.viewport.Width - 20
	txt := lipgloss.NewStyle().Foreground(lipgloss.Color("#EEE")).Render(truncate(e.Text, limit))
	m.appendLine(fmt.Sprintf("%s %s %s\n", ts, pre, txt))
}

func (m *dashboard) appendLine(line string) {
	m.logBuffer += line
	m.viewport.SetContent(m.logBuffer)
	m.viewport.GotoBottom()
}

func prefixFor(level engine.LogLevel, t Theme) (string, lipgloss.Color) {
	switch level {
	case engine.LogInfo:
		return "🔍 SCAN", t.Primary
	case engine.LogSuccess:
		return "💾 SAVE", t.Success
	case engine.LogError:
		return "💀 FAIL", t.Error
	case engine.LogAsset:
		return "📦 PACK", t.Secondary
	case engine.LogHunt:
		return "👁️ SEER", lipgloss.Color("#FF00FF")
	case engine.LogWarn:
		return "⚠️ WARN", t.Warning
	default:
		return "  ", t.Text
	}
}

func (m *dashboard) View() string {
	if m.quitting {
		return ""
	}
	if !m.ready {
		return "\n  Initialising Ikelos…"
	}
	cPrim, cSec, cDark := m.theme.Primary, m.theme.Secondary, m.theme.Dark

	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(cSec)
	if m.finished {
		border = border.BorderForeground(m.theme.Success)
	} else if m.paused {
		border = border.BorderForeground(m.theme.Warning)
	}

	modeStatus := string(m.cfg.Mode)
	if m.paused {
		modeStatus += " [PAUSED]"
	}
	headTxt := fmt.Sprintf(" %s %s | OP: %s | PROXY: %v ", engine.Name, engine.Version, modeStatus, m.cfg.Proxy != "")
	header := border.Width(m.width - 2).Align(lipgloss.Center).Render(renderGradient(headTxt, m.theme.Gradient))

	leftW := int(float64(m.width) * 0.35)
	rightW := m.width - leftW - 6

	lbl := lipgloss.NewStyle().Foreground(cPrim).Bold(true).Render
	val := lipgloss.NewStyle().Foreground(m.theme.Text).Render
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#555")).Render

	spark := ""
	bars := []string{" ", " ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	for _, v := range m.sparkline {
		if v < 0 {
			v = 0
		}
		if v >= len(bars) {
			v = len(bars) - 1
		}
		spark += bars[v]
	}

	blockTarget := lipgloss.JoinVertical(lipgloss.Left,
		lbl("🎯 TARGET ACQUISITION"),
		val(m.cfg.TargetURL.Host),
		dim(fmt.Sprintf("Depth: %d | Threads: %d", m.cfg.MaxDepth, m.cfg.Workers)),
	)
	blockPerf := lipgloss.JoinVertical(lipgloss.Left,
		lbl("⚡ SYSTEM METRICS"),
		fmt.Sprintf("MEM : %s", val(m.ramUsage)),
		fmt.Sprintf("GRT : %s", val(fmt.Sprintf("%d", m.goroutines))),
		fmt.Sprintf("NET : %s", val(fmt.Sprintf("%.2f MB/s", m.throughput))),
		fmt.Sprintf("ACT : %s", val(fmt.Sprintf("%d/%d", m.stats.Active, m.cfg.Workers))),
		fmt.Sprintf("QUE : %s", val(fmt.Sprintf("%d", m.stats.Queue))),
	)
	blockStats := lipgloss.JoinVertical(lipgloss.Left,
		lbl("📊 DATA INGESTION"),
		fmt.Sprintf("Files: %s", val(fmt.Sprintf("%d", m.stats.Files))),
		fmt.Sprintf("Err  : %s", lipgloss.NewStyle().Foreground(m.theme.Error).Render(fmt.Sprintf("%d", m.stats.Errors))),
		fmt.Sprintf("HTML : %s", val(fmt.Sprintf("%d", m.stats.HTML))),
		fmt.Sprintf("IMG  : %s", val(fmt.Sprintf("%d", m.stats.IMG))),
		fmt.Sprintf("CODE : %s", val(fmt.Sprintf("%d", m.stats.CODE))),
	)
	blockLive := lipgloss.JoinVertical(lipgloss.Left,
		lbl("📡 LIVE FEED"),
		dim(truncateLeft(m.lastLine, 35)),
		lipgloss.NewStyle().Foreground(cSec).Render(spark),
	)

	leftContent := lipgloss.JoinVertical(lipgloss.Left, blockTarget, "\n", blockPerf, "\n", blockStats, "\n", blockLive)
	leftP := border.Width(leftW).Height(m.viewport.Height).Background(cDark).Padding(1, 2).Render(leftContent)
	rightP := border.Width(rightW).Height(m.viewport.Height).Render(m.viewport.View())

	spin := m.spinner.View()
	status := "PROWL IN PROGRESS..."
	if m.finished {
		spin, status = "💎", "ARCHIVE SECURED."
	} else if m.paused {
		spin, status = "⏸️", "SYSTEM HALTED."
	}

	saturation := 0.0
	if m.finished {
		saturation = 1.0
	} else if m.cfg.Workers > 0 {
		saturation = float64(m.stats.Active) / float64(m.cfg.Workers)
	}

	barWidth := m.width - 50
	if barWidth < 10 {
		barWidth = 10
	}
	inf := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(4).Render(spin),
		lipgloss.NewStyle().Foreground(m.theme.Text).Width(22).Render(status),
		renderBar(barWidth, saturation, cPrim),
		lipgloss.NewStyle().Foreground(cPrim).Width(12).Align(lipgloss.Right).Render(fmt.Sprintf("%.0f%% LOAD", saturation*100)),
	)
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#555")).Render("[P] PAUSE/RESUME  •  [Q] ABORT MISSION")
	foot := border.Width(m.width - 2).Render(lipgloss.JoinVertical(lipgloss.Center, inf, hint))

	return lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinHorizontal(lipgloss.Top, leftP, rightP), foot)
}
