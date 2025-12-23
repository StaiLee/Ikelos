package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------
// CONSTANTS & THEMES
// ---------------------------------------------------------

const (
	AppVersion = "7.0.0 (THE ARCHITECT)"
	AppName    = "IKELOS"
)

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

// ---------------------------------------------------------
// UTILS
// ---------------------------------------------------------

func renderGradient(text string, colors []string) string {
	if len(text) == 0 {
		return ""
	}
	var s strings.Builder
	for i, char := range text {
		idx := int(float64(i) / float64(len(text)) * float64(len(colors)-1))
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colors[idx])).Render(string(char)))
	}
	return s.String()
}

func renderBar(width int, pct float64, color lipgloss.Color) string {
	if pct > 1.0 {
		pct = 1.0
	}
	wFilled := int(float64(width) * pct)
	if wFilled < 0 {
		wFilled = 0
	}
	wEmpty := width - wFilled
	if wEmpty < 0 {
		wEmpty = 0
	}
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("━", wFilled)) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#333")).Render(strings.Repeat("┄", wEmpty))
}

// ---------------------------------------------------------
// UI MODEL: INTERACTIVE MANUAL (RESTORED & UPDATED)
// ---------------------------------------------------------

type helpModel struct {
	pages         []string
	pageIdx       int
	width, height int
	quitting      bool
}

func initialHelpModel() helpModel { return helpModel{pages: buildPages(), pageIdx: 0} }
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
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m helpModel) View() string {
	if m.quitting {
		return ""
	}
	content := m.pages[m.pageIdx]

	navStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555"))
	activeDot := lipgloss.NewStyle().Foreground(ThemeMirror.Primary).Render("●")
	inactiveDot := lipgloss.NewStyle().Foreground(lipgloss.Color("#333")).Render("○")

	dots := ""
	for i := 0; i < len(m.pages); i++ {
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
		lipgloss.JoinVertical(lipgloss.Center, frame.Render(content), nav),
	)
}

func buildPages() []string {
	t := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFF")).Background(ThemeMirror.Secondary).Padding(0, 1).Render
	h := lipgloss.NewStyle().Bold(true).Foreground(ThemeMirror.Primary).MarginTop(1).Render
	c := lipgloss.NewStyle().Foreground(lipgloss.Color("#0F0")).Background(lipgloss.Color("#222")).Padding(0, 1).Render
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render

	p1 := fmt.Sprintf(`
%s

IKELOS v%s
THE ARCHITECT

%s
%s

"We do not just copy. We absorb."
`, renderGradient("/// TOTAL REALITY SHIFTING ENGINE ///", ThemeMirror.Gradient), "7.0", dim("Classified Tool"), dim("Systems Optimal."))

	p2 := fmt.Sprintf(`
%s

Ikelos v7 "The Architect" introduces intelligent structural analysis.

%s
1. %s : Scans robots.txt & sitemap.xml for target mapping.
2. %s : Analyzes DOM, detects Lazy-Load & Srcset images.
3. %s : Routes traffic via Proxies (if enabled) with random UAs.
4. %s : Hashing engine that rewrites filenames using MD5 to prevent OS errors.

%s
New features: Proxy Support, Tactical Pause, MD5 File Hashing.
`, t("SYSTEM ARCHITECTURE"), h("WORKFLOW"), c("The Hunter"), c("The Brain"), c("The Ghost"), c("The Vault"), h("CAPABILITIES"))

	p3 := fmt.Sprintf(`
%s

%s
%s
Target: High Fidelity Clone.
Specs:  Wait times, Full Recursion, Sitemap Check active.

%s
%s
Target: WAF Protected / Stealth.
Specs:  Slow, Jitter, Human Emulation. Requires Proxy for max effect.

%s
%s
Target: Bruteforce Download.
Specs:  MAX THREADS. No Mercy. 3 Retries per fail.
`, t("TACTICAL PROFILES"), h("1. MIRROR (Standard)"), c("-mode mirror"), h("2. SHADOW (Stealth)"), c("-mode shadow"), h("3. BLITZ (Speed)"), c("-mode blitz"))

	p4 := fmt.Sprintf(`
%s

%s
Target URL (http/https).

%s
Output directory for the clone.

%s
Proxy URL (HTTP/SOCKS5). Essential for "Ghost" operations.
Ex: http://127.0.0.1:8080

%s
Operational profile: mirror, shadow, blitz.

%s
Manual User-Agent override (Bypasses rotation).
`,
		t("COMMAND FLAGS (1/2)"), c("-url <URL>"), c("-out <DIR>"), c("-proxy <URL>"), c("-mode <MODE>"), c("-ua <STRING>"))

	p5 := fmt.Sprintf(`
%s

%s
Recursion depth (Links to follow).

%s
Concurrent workers.

%s
Disable Sitemap hunting.

%s
%s
Press 'P' during scan to Freeze/Resume the engine.
Perfect for analyzing real-time logs without losing context.
`,
		t("COMMAND FLAGS (2/2) & CONTROLS"), c("-depth <N>"), c("-threads <N>"), c("-nositemap"), h("RUNTIME CONTROLS"), c("[P] PAUSE / RESUME"))

	return []string{p1, p2, p3, p4, p5}
}

// ---------------------------------------------------------
// UI MODEL: CLONER DASHBOARD
// ---------------------------------------------------------

type tickMsg time.Time
type logMsg struct {
	Text string
	Type string
}
type finishedMsg struct{}
type pauseMsg bool

type model struct {
	viewport      viewport.Model
	spinner       spinner.Model
	width, height int
	quitting      bool
	finished      bool
	paused        bool
	logBuffer     string

	startTime time.Time
	config    Config
	theme     Theme
	engine    *IkelosEngine

	// UI Stats
	filesCount uint64
	bytesCount uint64
	errCount   uint64
	qSize      int

	countIMG  uint64
	countHTML uint64
	countCODE uint64

	throughput    float64
	sparkline     []int
	ramUsage      string
	goroutines    int
	activeWorkers int
	lastURL       string

	msgChan  chan logMsg
	doneChan chan bool
}

// ---------------------------------------------------------
// ENGINE CONFIG & TYPES
// ---------------------------------------------------------

type Mode string

const (
	ModeBlitz  Mode = "BLITZ"
	ModeShadow Mode = "SHADOW"
	ModeMirror Mode = "MIRROR"
)

type Config struct {
	TargetURL   *url.URL
	OutputDir   string
	Mode        Mode
	MaxDepth    int
	Workers     int
	Timeout     time.Duration
	SkipSitemap bool
	Proxy       string
	UserAgent   string
}

type IkelosEngine struct {
	Config    Config
	Client    *http.Client
	Visited   sync.Map
	Semaphore chan struct{}
	WaitGroup sync.WaitGroup
	MsgChan   chan logMsg
	IsPaused  atomic.Bool

	// Stats
	Files    uint64
	Bytes    uint64
	Errors   uint64
	TypeIMG  uint64
	TypeHTML uint64
	TypeCODE uint64
	QueueLen int64

	// Robots
	RobotsDisallowed []string
}

// ---------------------------------------------------------
// INIT UI (DASHBOARD)
// ---------------------------------------------------------

func initialModel(cfg Config, engine *IkelosEngine, mChan chan logMsg, dChan chan bool) model {
	t := ThemeMirror
	switch cfg.Mode {
	case ModeBlitz:
		t = ThemeBlitz
	case ModeShadow:
		t = ThemeShadow
	}

	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(t.Primary)

	vp := viewport.New(0, 0)

	return model{
		spinner:   s,
		viewport:  vp,
		startTime: time.Now(),
		config:    cfg,
		engine:    engine,
		theme:     t,
		sparkline: make([]int, 40),
		msgChan:   mChan,
		doneChan:  dChan,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitForLog(m.msgChan), waitForDone(m.doneChan), tickCmd())
}
func waitForLog(sub chan logMsg) tea.Cmd { return func() tea.Msg { return <-sub } }
func waitForDone(sub chan bool) tea.Cmd  { return func() tea.Msg { <-sub; return finishedMsg{} } }
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// ---------------------------------------------------------
// UPDATE LOOP
// ---------------------------------------------------------

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "p": // Pause Feature
			m.paused = !m.paused
			m.engine.IsPaused.Store(m.paused)
			if m.paused {
				m.logBuffer += lipgloss.NewStyle().Foreground(m.theme.Warning).Render("\n[!] SYSTEM PAUSED BY USER [!]\n")
			} else {
				m.logBuffer += lipgloss.NewStyle().Foreground(m.theme.Success).Render("\n[>] SYSTEM RESUMED\n")
			}
			m.viewport.SetContent(m.logBuffer)
			m.viewport.GotoBottom()
			return m, nil
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h := m.height - 12
		if h < 10 {
			h = 10
		}
		m.viewport.Width = int(float64(m.width) * 0.65)
		m.viewport.Height = h
		return m, nil

	case spinner.TickMsg:
		if m.finished {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		m.filesCount = atomic.LoadUint64(&m.engine.Files)
		m.bytesCount = atomic.LoadUint64(&m.engine.Bytes)
		m.errCount = atomic.LoadUint64(&m.engine.Errors)
		m.countIMG = atomic.LoadUint64(&m.engine.TypeIMG)
		m.countHTML = atomic.LoadUint64(&m.engine.TypeHTML)
		m.countCODE = atomic.LoadUint64(&m.engine.TypeCODE)
		m.qSize = int(atomic.LoadInt64(&m.engine.QueueLen))

		m.activeWorkers = len(m.engine.Semaphore)

		elapsed := time.Since(m.startTime).Seconds()
		if elapsed > 0 {
			m.throughput = (float64(m.bytesCount) / 1024 / 1024) / elapsed
		}

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		m.ramUsage = fmt.Sprintf("%d MB", mem.Alloc/1024/1024)
		m.goroutines = runtime.NumGoroutine()

		activity := m.activeWorkers
		if activity > 10 {
			activity = 10
		}
		m.sparkline = append(m.sparkline[1:], activity)
		return m, tickCmd()

	case logMsg:
		prefix := "  "
		color := m.theme.Text

		switch msg.Type {
		case "INFO":
			prefix = "🔍 SCAN"
			color = m.theme.Primary
		case "SUCCESS":
			prefix = "💾 SAVE"
			color = m.theme.Success
		case "ERROR":
			prefix = "💀 FAIL"
			color = m.theme.Error
		case "ASSET":
			prefix = "📦 PACK"
			color = m.theme.Secondary
		case "HUNT":
			prefix = "👁️ SEER"
			color = lipgloss.Color("#FF00FF")
		case "WARN":
			prefix = "⚠️ WARN"
			color = m.theme.Warning
		}

		if msg.Text != "" {
			m.lastURL = msg.Text
			ts := lipgloss.NewStyle().Foreground(lipgloss.Color("#555")).Render(time.Now().Format("15:04:05"))
			pre := lipgloss.NewStyle().Foreground(color).Bold(true).Width(7).Render(prefix)
			txt := lipgloss.NewStyle().Foreground(lipgloss.Color("#EEE")).Render(msg.Text)
			if len(txt) > m.viewport.Width-20 {
				txt = txt[:m.viewport.Width-23] + "..."
			}
			m.logBuffer += fmt.Sprintf("%s %s %s\n", ts, pre, txt)
			m.viewport.SetContent(m.logBuffer)
			m.viewport.GotoBottom()
		}
		return m, waitForLog(m.msgChan)

	case finishedMsg:
		m.finished = true
		return m, nil
	}
	return m, nil
}

// ---------------------------------------------------------
// VIEW (RENDER)
// ---------------------------------------------------------

func (m model) View() string {
	if m.quitting {
		return ""
	}
	cPrim := m.theme.Primary
	cSec := m.theme.Secondary
	cDark := m.theme.Dark

	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(cSec)
	if m.finished {
		border = border.BorderForeground(m.theme.Success)
	} else if m.paused {
		border = border.BorderForeground(m.theme.Warning)
	}

	// Header
	modeStatus := string(m.config.Mode)
	if m.paused {
		modeStatus += " [PAUSED]"
	}
	headTxt := fmt.Sprintf(" %s %s | OP: %s | PROXY: %v ", AppName, AppVersion, modeStatus, m.config.Proxy != "")
	title := renderGradient(headTxt, m.theme.Gradient)
	header := border.Copy().Width(m.width - 2).Align(lipgloss.Center).Render(title)

	// Layout
	leftW := int(float64(m.width) * 0.35)
	rightW := m.width - leftW - 6

	lbl := lipgloss.NewStyle().Foreground(cPrim).Bold(true).Render
	val := lipgloss.NewStyle().Foreground(m.theme.Text).Render
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#555")).Render

	spark := ""
	bars := []string{" ", " ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	for _, v := range m.sparkline {
		if v >= len(bars) {
			v = len(bars) - 1
		}
		spark += bars[v]
	}
	sparkRender := lipgloss.NewStyle().Foreground(cSec).Render(spark)

	// Panels
	blockTarget := lipgloss.JoinVertical(lipgloss.Left,
		lbl("🎯 TARGET ACQUISITION"),
		val(m.config.TargetURL.Host),
		dim(fmt.Sprintf("Depth: %d | Threads: %d", m.config.MaxDepth, m.config.Workers)),
	)

	blockPerf := lipgloss.JoinVertical(lipgloss.Left,
		lbl("⚡ SYSTEM METRICS"),
		fmt.Sprintf("MEM : %s", val(m.ramUsage)),
		fmt.Sprintf("GRT : %s", val(fmt.Sprintf("%d", m.goroutines))),
		fmt.Sprintf("NET : %s", val(fmt.Sprintf("%.2f MB/s", m.throughput))),
		fmt.Sprintf("ACT : %s", val(fmt.Sprintf("%d/%d", m.activeWorkers, m.config.Workers))),
	)

	blockStats := lipgloss.JoinVertical(lipgloss.Left,
		lbl("📊 DATA INGESTION"),
		fmt.Sprintf("Files: %s", val(fmt.Sprintf("%d", m.filesCount))),
		fmt.Sprintf("Err  : %s", lipgloss.NewStyle().Foreground(m.theme.Error).Render(fmt.Sprintf("%d", m.errCount))),
		fmt.Sprintf("HTML : %s", val(fmt.Sprintf("%d", m.countHTML))),
		fmt.Sprintf("IMG  : %s", val(fmt.Sprintf("%d", m.countIMG))),
		fmt.Sprintf("CODE : %s", val(fmt.Sprintf("%d", m.countCODE))),
	)

	lastUrlDisplay := m.lastURL
	if len(lastUrlDisplay) > 35 {
		lastUrlDisplay = "..." + lastUrlDisplay[len(lastUrlDisplay)-35:]
	}
	blockLive := lipgloss.JoinVertical(lipgloss.Left,
		lbl("📡 LIVE FEED"),
		dim(lastUrlDisplay),
		sparkRender,
	)

	leftContent := lipgloss.JoinVertical(lipgloss.Left,
		blockTarget, "\n", blockPerf, "\n", blockStats, "\n", blockLive,
	)
	leftP := border.Copy().Width(leftW).Height(m.viewport.Height).Background(cDark).Padding(1, 2).Render(leftContent)
	rightP := border.Copy().Width(rightW).Height(m.viewport.Height).Render(m.viewport.View())

	// Footer
	spin := m.spinner.View()
	status := "PROWL IN PROGRESS..."
	if m.finished {
		spin = "💎"
		status = "ARCHIVE SECURED."
	} else if m.paused {
		spin = "⏸️"
		status = "SYSTEM HALTED."
	}

	saturation := 0.0
	if m.finished {
		saturation = 1.0
	} else {
		saturation = float64(m.activeWorkers) / float64(m.config.Workers)
	}

	barWidth := m.width - 50
	if barWidth < 10 {
		barWidth = 10
	}
	bar := renderBar(barWidth, saturation, cPrim)
	pctText := fmt.Sprintf("%.0f%% LOAD", saturation*100)

	inf := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(4).Render(spin),
		lipgloss.NewStyle().Foreground(m.theme.Text).Width(22).Render(status),
		bar,
		lipgloss.NewStyle().Foreground(cPrim).Width(12).Align(lipgloss.Right).Render(pctText),
	)

	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#555")).Render("[P] PAUSE/RESUME  •  [Q] ABORT MISSION")

	foot := border.Copy().Width(m.width - 2).Render(
		lipgloss.JoinVertical(lipgloss.Center, inf, hint),
	)

	return lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinHorizontal(lipgloss.Top, leftP, rightP), foot)
}

// ---------------------------------------------------------
// ENGINE LOGIC (THE BRAIN)
// ---------------------------------------------------------

func NewEngine(cfg Config, mChan chan logMsg) *IkelosEngine {
	// Configure Transport with Proxy Support
	t := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 500,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DisableCompression:  false,
	}

	// Proxy Setup
	if cfg.Proxy != "" {
		proxyUrl, err := url.Parse(cfg.Proxy)
		if err == nil {
			t.Proxy = http.ProxyURL(proxyUrl)
		}
	}

	e := &IkelosEngine{
		Config:    cfg,
		Client:    &http.Client{Transport: t, Timeout: cfg.Timeout},
		Semaphore: make(chan struct{}, cfg.Workers),
		MsgChan:   mChan,
	}
	e.IsPaused.Store(false)
	return e
}

func (e *IkelosEngine) Log(text, lType string) {
	select {
	case e.MsgChan <- logMsg{Text: text, Type: lType}:
	default:
	}
}

func (e *IkelosEngine) checkPause() {
	for e.IsPaused.Load() {
		time.Sleep(500 * time.Millisecond)
	}
}

// ROBOTS.TXT PARSER
func (e *IkelosEngine) fetchRobotsTxt() {
	robotsURL := e.Config.TargetURL.Scheme + "://" + e.Config.TargetURL.Host + "/robots.txt"
	data, code := e.fetch(robotsURL, "")
	if code == 200 {
		e.Log("Parsing robots.txt...", "INFO")
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Disallow:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					rule := strings.TrimSpace(parts[1])
					if rule != "" {
						e.RobotsDisallowed = append(e.RobotsDisallowed, rule)
					}
				}
			}
		}
	}
}

func (e *IkelosEngine) isAllowed(target string) bool {
	u, err := url.Parse(target)
	if err != nil {
		return false
	}
	for _, rule := range e.RobotsDisallowed {
		if strings.HasPrefix(u.Path, rule) {
			return false
		}
	}
	return true
}

func (e *IkelosEngine) Run() {
	e.fetchRobotsTxt()

	if !e.Config.SkipSitemap {
		go e.huntSitemap()
	}

	e.WaitGroup.Add(1)
	go e.crawlPage(e.Config.TargetURL.String(), 0)
	e.WaitGroup.Wait()
}

func (e *IkelosEngine) huntSitemap() {
	sitemapURL := e.Config.TargetURL.Scheme + "://" + e.Config.TargetURL.Host + "/sitemap.xml"
	e.Log("Scanning Sitemap: "+sitemapURL, "HUNT")

	data, code := e.fetch(sitemapURL, "")
	if code == 200 {
		re := regexp.MustCompile(`<loc>(.*?)</loc>`)
		matches := re.FindAllStringSubmatch(string(data), -1)
		count := 0
		for _, m := range matches {
			if len(m) > 1 {
				urlStr := strings.TrimSpace(m[1])
				if strings.Contains(urlStr, e.Config.TargetURL.Host) {
					e.WaitGroup.Add(1)
					go e.crawlPage(urlStr, 1)
					count++
				}
			}
		}
		e.Log(fmt.Sprintf("Sitemap extracted: %d entries", count), "HUNT")
	}
}

func (e *IkelosEngine) crawlPage(target string, depth int) {
	defer e.WaitGroup.Done()
	e.checkPause()

	if depth > e.Config.MaxDepth {
		return
	}
	if strings.Contains(target, "#") {
		target = strings.Split(target, "#")[0]
	}

	if !e.isAllowed(target) {
		e.Log("Blocked by Robots.txt: "+filepath.Base(target), "WARN")
		return
	}

	if _, loaded := e.Visited.LoadOrStore(target, true); loaded {
		return
	}

	if e.Config.Mode == ModeShadow {
		time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)
	}

	atomic.AddInt64(&e.QueueLen, 1)
	e.Semaphore <- struct{}{}
	defer func() {
		<-e.Semaphore
		atomic.AddInt64(&e.QueueLen, -1)
	}()

	e.Log("Crawling: "+filepath.Base(target), "INFO")

	var bodyBytes []byte
	var statusCode int
	for retries := 0; retries < 3; retries++ {
		e.checkPause()
		bodyBytes, statusCode = e.fetch(target, "")
		if statusCode == 200 {
			break
		}
		time.Sleep(time.Duration(math.Pow(2, float64(retries))) * 200 * time.Millisecond)
	}

	if bodyBytes == nil || statusCode != 200 {
		return
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
	if err != nil {
		return
	}

	e.processAssetsAndRewrite(doc, target)

	finalHTML, err := doc.Html()
	if err == nil {
		e.saveFile(target, []byte(finalHTML))
		atomic.AddUint64(&e.TypeHTML, 1)
		e.Log(filepath.Base(target), "SUCCESS")
	}

	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			absURL := e.resolveURL(href, target)
			if absURL != "" && strings.Contains(absURL, e.Config.TargetURL.Host) {
				ext := strings.ToLower(filepath.Ext(absURL))
				if ext == "" || ext == ".html" || ext == ".htm" || ext == ".php" || ext == ".asp" {
					e.WaitGroup.Add(1)
					go e.crawlPage(absURL, depth+1)
				} else {
					e.downloadAsync(absURL, target)
				}
			}
		}
	})
}

func (e *IkelosEngine) processAssetsAndRewrite(doc *goquery.Document, pageURL string) {
	doc.Find("img, source, link, script, video, audio").Each(func(_ int, s *goquery.Selection) {
		attrs := []string{"src", "href", "data-src", "data-original", "poster"}
		for _, attr := range attrs {
			val, exists := s.Attr(attr)
			if exists && val != "" {
				absURL := e.resolveURL(val, pageURL)
				if absURL != "" {
					e.downloadAsync(absURL, pageURL)
					s.SetAttr(attr, e.getRelPath(pageURL, absURL))
				}
			}
		}
		srcset, exists := s.Attr("srcset")
		if exists && srcset != "" {
			newSrcset := e.rewriteSrcset(srcset, pageURL)
			s.SetAttr("srcset", newSrcset)
		}
	})

	doc.Find("*").Each(func(_ int, s *goquery.Selection) {
		style, exists := s.Attr("style")
		if exists && strings.Contains(style, "url") {
			s.SetAttr("style", e.rewriteCSSString(style, pageURL))
		}
	})
}

func (e *IkelosEngine) rewriteSrcset(srcset string, pageURL string) string {
	parts := strings.Split(srcset, ",")
	var newParts []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		fields := strings.Fields(part)
		if len(fields) > 0 {
			rawURL := fields[0]
			absURL := e.resolveURL(rawURL, pageURL)
			if absURL != "" {
				e.downloadAsync(absURL, pageURL)
				relPath := e.getRelPath(pageURL, absURL)
				if len(fields) > 1 {
					newParts = append(newParts, relPath+" "+strings.Join(fields[1:], " "))
				} else {
					newParts = append(newParts, relPath)
				}
			} else {
				newParts = append(newParts, part)
			}
		}
	}
	return strings.Join(newParts, ", ")
}

func (e *IkelosEngine) downloadAsync(url, referer string) {
	e.WaitGroup.Add(1)
	go e.downloadAsset(url, referer)
}

func (e *IkelosEngine) downloadAsset(target, referer string) {
	defer e.WaitGroup.Done()
	e.checkPause()

	if _, loaded := e.Visited.LoadOrStore(target, true); loaded {
		return
	}

	e.Semaphore <- struct{}{}
	data, code := e.fetch(target, referer)
	<-e.Semaphore

	if data == nil || code != 200 {
		return
	}

	if strings.HasSuffix(strings.Split(target, "?")[0], ".css") {
		data = []byte(e.rewriteCSSString(string(data), target))
		atomic.AddUint64(&e.TypeCODE, 1)
	} else if strings.HasSuffix(target, ".js") {
		atomic.AddUint64(&e.TypeCODE, 1)
	} else {
		atomic.AddUint64(&e.TypeIMG, 1)
	}

	e.saveFile(target, data)
	e.Log(filepath.Base(target), "ASSET")
}

var cssUrlRegex = regexp.MustCompile(`(?:url\(['"]?|@import\s+['"]?)([^'"\)]+)['"]?\)`)

func (e *IkelosEngine) rewriteCSSString(content, baseURL string) string {
	return cssUrlRegex.ReplaceAllStringFunc(content, func(match string) string {
		subMatch := cssUrlRegex.FindStringSubmatch(match)
		if len(subMatch) < 2 {
			return match
		}
		rawUrl := strings.TrimSpace(subMatch[1])
		if strings.HasPrefix(rawUrl, "data:") {
			return match
		}
		absURL := e.resolveURL(rawUrl, baseURL)
		if absURL == "" {
			return match
		}
		e.downloadAsync(absURL, baseURL)
		relPath := e.getRelPath(baseURL, absURL)

		if strings.HasPrefix(match, "@import") {
			return fmt.Sprintf("@import '%s')", strings.ReplaceAll(relPath, "\\", "/"))
		}
		return fmt.Sprintf("url('%s')", strings.ReplaceAll(relPath, "\\", "/"))
	})
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0",
}

func (e *IkelosEngine) fetch(target, referer string) ([]byte, int) {
	req, _ := http.NewRequest("GET", target, nil)

	if e.Config.UserAgent != "" {
		req.Header.Set("User-Agent", e.Config.UserAgent)
	} else {
		req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
	}

	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := e.Client.Do(req)
	if err != nil {
		atomic.AddUint64(&e.Errors, 1)
		return nil, 0
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode
	}

	atomic.AddUint64(&e.Bytes, uint64(len(data)))
	return data, resp.StatusCode
}

func (e *IkelosEngine) saveFile(urlStr string, data []byte) {
	fullPath := e.getFilePath(urlStr)
	os.MkdirAll(filepath.Dir(fullPath), 0755)
	os.WriteFile(fullPath, data, 0644)
	atomic.AddUint64(&e.Files, 1)
}

func (e *IkelosEngine) getFilePath(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return filepath.Join(e.Config.OutputDir, "unknown")
	}

	path := u.Path
	if path == "" || path == "/" {
		path = "/index.html"
	} else if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	} else if filepath.Ext(path) == "" {
		path = path + ".html"
	}

	path = strings.ReplaceAll(path, ":", "_")

	if u.RawQuery != "" {
		hasher := md5.New()
		hasher.Write([]byte(u.RawQuery))
		hash := hex.EncodeToString(hasher.Sum(nil))[:8]

		ext := filepath.Ext(path)
		name := strings.TrimSuffix(path, ext)
		path = fmt.Sprintf("%s_%s%s", name, hash, ext)
	}

	if len(filepath.Base(path)) > 200 {
		ext := filepath.Ext(path)
		hasher := md5.New()
		hasher.Write([]byte(filepath.Base(path)))
		hash := hex.EncodeToString(hasher.Sum(nil))
		path = filepath.Join(filepath.Dir(path), hash+ext)
	}

	return filepath.Join(e.Config.OutputDir, u.Host, path)
}

func (e *IkelosEngine) resolveURL(href, base string) string {
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(u).String()
}

func (e *IkelosEngine) getRelPath(sourceURL, targetURL string) string {
	sourcePath := e.getFilePath(sourceURL)
	targetPath := e.getFilePath(targetURL)
	rel, err := filepath.Rel(filepath.Dir(sourcePath), targetPath)
	if err != nil {
		return targetURL
	}
	return filepath.ToSlash(rel)
}

// ---------------------------------------------------------
// MAIN
// ---------------------------------------------------------

func main() {
	urlStr := flag.String("url", "", "Target URL")
	outFlag := flag.String("out", "./cloned_site", "Output Dir")
	modeFlag := flag.String("mode", "MIRROR", "Mode: BLITZ, SHADOW, MIRROR")
	workersFlag := flag.Int("threads", 20, "Threads")
	depthFlag := flag.Int("depth", 2, "Depth")
	noSitemap := flag.Bool("nositemap", false, "Disable Sitemap Hunt")
	proxyFlag := flag.String("proxy", "", "Proxy URL (e.g. http://127.0.0.1:8080)")
	uaFlag := flag.String("ua", "", "Custom User-Agent")
	flag.Parse()

	// INTERACTIVE MANUAL LAUNCH (if no URL provided)
	if *urlStr == "" {
		p := tea.NewProgram(initialHelpModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Println("Error starting manual:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	targetURL, err := url.Parse(*urlStr)
	if err != nil {
		fmt.Println("Invalid URL")
		return
	}

	mode := Mode(strings.ToUpper(*modeFlag))
	timeout := 45 * time.Second
	if mode == ModeBlitz {
		timeout = 15 * time.Second
	}

	cfg := Config{
		TargetURL: targetURL, OutputDir: *outFlag, Mode: mode, MaxDepth: *depthFlag,
		Workers: *workersFlag, Timeout: timeout, SkipSitemap: *noSitemap,
		Proxy: *proxyFlag, UserAgent: *uaFlag,
	}

	msgChan := make(chan logMsg, 5000)
	doneChan := make(chan bool)

	engine := NewEngine(cfg, msgChan)
	uiModel := initialModel(cfg, engine, msgChan, doneChan)

	go func() {
		engine.Run()
		doneChan <- true
	}()

	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	c := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).Render
	fmt.Printf("\n%s\n", c(fmt.Sprintf("[+] ARCHIVE COMPLETE. %d files downloaded.", atomic.LoadUint64(&engine.Files))))
	if engine.Errors > 0 {
		fmt.Printf("[!] Encountered %d errors (check logs).\n", atomic.LoadUint64(&engine.Errors))
	}
}
