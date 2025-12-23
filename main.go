package main

import (
	"bytes"
	"crypto/tls"
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
// CONFIGURATION & THEMES
// ---------------------------------------------------------

const (
	AppVersion = "6.1.0 (THE OVERSEER)"
	AppName    = "IKELOS"
)

type Theme struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Dark      lipgloss.Color
	Text      lipgloss.Color
	Success   lipgloss.Color
	Error     lipgloss.Color
	Gradient  []string
}

var (
	ThemeBlitz = Theme{
		Primary:   lipgloss.Color("#FF3300"),
		Secondary: lipgloss.Color("#FF8800"),
		Dark:      lipgloss.Color("#1a0500"),
		Text:      lipgloss.Color("#FFCC00"),
		Success:   lipgloss.Color("#00FF00"),
		Error:     lipgloss.Color("#FF0000"),
		Gradient:  []string{"#FF0000", "#FF4400", "#FF8800"},
	}
	ThemeMirror = Theme{
		Primary:   lipgloss.Color("#00FFFF"),
		Secondary: lipgloss.Color("#0066FF"),
		Dark:      lipgloss.Color("#00051a"),
		Text:      lipgloss.Color("#E0FFFF"),
		Success:   lipgloss.Color("#00FF99"),
		Error:     lipgloss.Color("#FF3333"),
		Gradient:  []string{"#0000FF", "#0088FF", "#00FFFF"},
	}
	ThemeShadow = Theme{
		Primary:   lipgloss.Color("#FFFFFF"),
		Secondary: lipgloss.Color("#666666"),
		Dark:      lipgloss.Color("#111111"),
		Text:      lipgloss.Color("#CCCCCC"),
		Success:   lipgloss.Color("#FFFFFF"),
		Error:     lipgloss.Color("#555555"),
		Gradient:  []string{"#333333", "#888888", "#FFFFFF"},
	}
)

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

	filled := strings.Repeat("█", wFilled)
	empty := strings.Repeat("░", wEmpty)

	return lipgloss.NewStyle().Foreground(color).Render(filled) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#333")).Render(empty)
}

// ---------------------------------------------------------
// UI MODEL: HELP / MANUAL
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
	dotActive := lipgloss.NewStyle().Foreground(ThemeMirror.Primary).Render("●")
	dotInactive := lipgloss.NewStyle().Foreground(lipgloss.Color("#333")).Render("○")
	dots := ""
	for i := 0; i < len(m.pages); i++ {
		if i == m.pageIdx {
			dots += dotActive + " "
		} else {
			dots += dotInactive + " "
		}
	}
	frame := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ThemeMirror.Primary).Padding(1, 3).Width(90).Align(lipgloss.Center)
	nav := fmt.Sprintf("\n%s\n%s", dots, navStyle.Render("[ARROWS] NAVIGATE  •  [Q] EXIT"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, frame.Render(content), nav))
}

func buildPages() []string {
	t := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFF")).Background(ThemeMirror.Secondary).Padding(0, 1).Render
	h := lipgloss.NewStyle().Bold(true).Foreground(ThemeMirror.Primary).MarginTop(1).Render
	c := lipgloss.NewStyle().Foreground(lipgloss.Color("#0F0")).Background(lipgloss.Color("#222")).Padding(0, 1).Render
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render

	p1 := fmt.Sprintf(`
%s

IKELOS v%s
THE OMNISCIENT

%s
%s

"We do not just copy. We absorb."
`, renderGradient("/// TOTAL SITE REPLICATION SYSTEM ///", ThemeMirror.Gradient), "6.1", dim("Classified Tool"), dim("Systems Optimal."))

	p2 := fmt.Sprintf(`
%s

Ikelos v6 introduces the "Omniscient Engine".

%s
1. %s : Checks robots.txt & sitemap.xml to find hidden pages.
2. %s : Analyzes DOM structure & rewrites CSS/JS paths.
3. %s : Handles concurrent downloads with Retry Logic.
4. %s : Unlocks Lazy-Loaded assets automatically.

%s
Features: User-Agent Rotation, Header Spoofing, Sitemap Hunting.
`, t("SYSTEM ARCHITECTURE"), h("WORKFLOW"), c("The Hunter"), c("The Brain"), c("The Swarm"), c("The Surgeon"), h("CAPABILITIES"))

	p3 := fmt.Sprintf(`
%s

%s
%s
Target: High Fidelity Clone.
Specs:  Wait times, Full Recursion, Sitemap Check active.

%s
%s
Target: WAF Protected / Stealth.
Specs:  Slow, Jitter, Human Emulation (Mouse move simulation headers).

%s
%s
Target: Bruteforce Download.
Specs:  MAX THREADS. No Mercy. 3 Retries per fail.

`, t("TACTICAL MODES"), h("1. MIRROR (Standard)"), c("-mode mirror"), h("2. SHADOW (Stealth)"), c("-mode shadow"), h("3. BLITZ (Speed)"), c("-mode blitz"))

	p4 := fmt.Sprintf(`
%s

%s
Target URL (http/https).

%s
Output directory for the clone.

%s
Operational profile: mirror, shadow, blitz.

%s
Recursion depth (Links to follow).

%s
Concurrent workers.

%s
Disable Sitemap hunting (if you want to stay strictly on page).
`,
		t("COMMAND FLAGS"), c("-url <URL>"), c("-out <DIR>"), c("-mode <MODE>"), c("-depth <N>"), c("-threads <N>"), c("-nositemap"))

	return []string{p1, p2, p3, p4}
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

type model struct {
	viewport      viewport.Model
	spinner       spinner.Model
	width, height int
	quitting      bool
	finished      bool
	logBuffer     string

	startTime time.Time
	config    Config
	theme     Theme

	// Stats pointers (read from Engine directly)
	engine *IkelosEngine

	// Display stats
	filesCount uint64
	bytesCount uint64
	errCount   uint64

	countIMG  uint64
	countHTML uint64
	countCODE uint64

	throughput    float64 // MB/s
	sparkline     []int
	ramUsage      string
	goroutines    int
	activeWorkers int

	msgChan  chan logMsg
	doneChan chan bool
}

// ---------------------------------------------------------
// MOTEUR IKELOS
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
}

type IkelosEngine struct {
	Config    Config
	Client    *http.Client
	Visited   sync.Map
	Semaphore chan struct{} // Used to track active workers
	WaitGroup sync.WaitGroup
	MsgChan   chan logMsg

	// Atomic Counters
	Files  uint64
	Bytes  uint64
	Errors uint64

	// Content Type Counters
	TypeIMG  uint64
	TypeHTML uint64
	TypeCODE uint64
}

// ---------------------------------------------------------
// INIT UI (CLONER)
// ---------------------------------------------------------

func initialModel(cfg Config, engine *IkelosEngine, mChan chan logMsg, dChan chan bool) model {
	t := ThemeMirror
	if cfg.Mode == ModeBlitz {
		t = ThemeBlitz
	}
	if cfg.Mode == ModeShadow {
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
		sparkline: make([]int, 30),
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
		if msg.String() == "q" || msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h := m.height - 10
		if h < 10 {
			h = 10
		}
		m.viewport.Width = int(float64(m.width) * 0.60)
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
		// Read Atomic Stats
		m.filesCount = atomic.LoadUint64(&m.engine.Files)
		m.bytesCount = atomic.LoadUint64(&m.engine.Bytes)
		m.errCount = atomic.LoadUint64(&m.engine.Errors)
		m.countIMG = atomic.LoadUint64(&m.engine.TypeIMG)
		m.countHTML = atomic.LoadUint64(&m.engine.TypeHTML)
		m.countCODE = atomic.LoadUint64(&m.engine.TypeCODE)

		// Active workers calculation (Semaphore length)
		m.activeWorkers = len(m.engine.Semaphore)

		// Throughput
		elapsed := time.Since(m.startTime).Seconds()
		if elapsed > 0 {
			m.throughput = (float64(m.bytesCount) / 1024 / 1024) / elapsed
		}

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		m.ramUsage = fmt.Sprintf("%d MB", mem.Alloc/1024/1024)
		m.goroutines = runtime.NumGoroutine()

		// Sparkline logic (visual effect based on active workers)
		activity := m.activeWorkers
		if activity > 8 {
			activity = 8
		}
		m.sparkline = append(m.sparkline[1:], activity)
		return m, tickCmd()

	case logMsg:
		prefix := "  "
		color := m.theme.Text

		// NOUVEAUX TAGS UX
		switch msg.Type {
		case "INFO":
			prefix = "🔎 INDEX"
			color = m.theme.Primary
		case "SUCCESS":
			prefix = "📄 PAGE "
			color = m.theme.Success
		case "ERROR":
			prefix = "❌ ERROR"
			color = m.theme.Error
		case "ASSET":
			prefix = "📦 ASSET"
			color = m.theme.Secondary
		case "HUNT":
			prefix = "🕸️  SITE"
			color = lipgloss.Color("#FF00FF")
		}

		if msg.Text != "" {
			ts := lipgloss.NewStyle().Foreground(lipgloss.Color("#444")).Render(time.Now().Format("15:04:05"))
			pre := lipgloss.NewStyle().Foreground(color).Bold(true).Width(8).Render(prefix) // Fixed width for alignment
			txt := lipgloss.NewStyle().Foreground(lipgloss.Color("#EEE")).Render(msg.Text)
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
		border = border.BorderForeground(cPrim)
	}

	// Header
	headTxt := fmt.Sprintf(" %s %s | MODE: %s | OUT: %s ", AppName, AppVersion, m.config.Mode, filepath.Base(m.config.OutputDir))
	title := renderGradient(headTxt, m.theme.Gradient)
	header := border.Copy().Width(m.width - 2).Align(lipgloss.Center).Render(title)

	// Panels
	leftW := int(float64(m.width) * 0.30)
	rightW := m.width - leftW - 6

	lbl := lipgloss.NewStyle().Foreground(cPrim).Bold(true).Render
	val := lipgloss.NewStyle().Foreground(m.theme.Text).Render
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#444")).Render

	// Sparkline
	spark := ""
	bars := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	for _, v := range m.sparkline {
		if v >= len(bars) {
			v = len(bars) - 1
		}
		spark += bars[v]
	}
	sparkRender := lipgloss.NewStyle().Foreground(cSec).Render(spark)

	// Block 1: Target
	u := m.config.TargetURL
	blockTarget := lipgloss.JoinVertical(lipgloss.Left,
		lbl("TARGET LOCK"),
		val(u.Host),
		dim(u.Scheme+" protocol"),
	)

	// Block 2: Vitals
	blockPerf := lipgloss.JoinVertical(lipgloss.Left,
		lbl("SYSTEM VITALS"),
		val(fmt.Sprintf("RAM: %s", m.ramUsage)),
		val(fmt.Sprintf("GRT: %d", m.goroutines)),
		val(fmt.Sprintf("SPD: %.2f MB/s", m.throughput)),
	)

	// Block 3: Payload Analysis (NEW UX FEATURE)
	blockPayload := lipgloss.JoinVertical(lipgloss.Left,
		lbl("PAYLOAD ANALYSIS"),
		fmt.Sprintf("IMG  : %s", val(fmt.Sprintf("%d", m.countIMG))),
		fmt.Sprintf("HTML : %s", val(fmt.Sprintf("%d", m.countHTML))),
		fmt.Sprintf("CODE : %s", val(fmt.Sprintf("%d", m.countCODE))),
	)

	// Block 4: Metrics
	dataSize := float64(m.bytesCount) / 1024 / 1024
	sizeUnit := "MB"
	if dataSize > 1024 {
		dataSize /= 1024
		sizeUnit = "GB"
	}

	blockData := lipgloss.JoinVertical(lipgloss.Left,
		lbl("TOTAL INGESTION"),
		fmt.Sprintf("FILES : %s", val(fmt.Sprintf("%d", m.filesCount))),
		fmt.Sprintf("SIZE  : %s", val(fmt.Sprintf("%.2f %s", dataSize, sizeUnit))),
		fmt.Sprintf("ERRS  : %s", lipgloss.NewStyle().Foreground(m.theme.Error).Bold(true).Render(fmt.Sprintf("%d", m.errCount))),
	)

	leftContent := lipgloss.JoinVertical(lipgloss.Left,
		blockTarget, "\n", blockPerf, "\n", blockPayload, "\n", blockData, "\n",
		lbl("NET FLUX"), sparkRender,
	)

	leftP := border.Copy().Width(leftW).Height(m.viewport.Height).Background(cDark).Padding(1, 2).Render(leftContent)
	rightP := border.Copy().Width(rightW).Height(m.viewport.Height).Render(m.viewport.View())

	// Footer (PROGRESS BAR)
	spin := m.spinner.View()
	status := "HUNTING ASSETS..."
	if m.finished {
		spin = "✅"
		status = "ACQUISITION COMPLETE."
	}

	// Bar logic: Saturation (Active Workers / Max Configured Workers)
	barWidth := m.width - 45
	if barWidth < 10 {
		barWidth = 10
	}

	// If finished, bar is full 100%
	saturation := 0.0
	if m.finished {
		saturation = 1.0
	} else {
		// Calculate % of workers busy
		saturation = float64(m.activeWorkers) / float64(m.config.Workers)
	}

	bar := renderBar(barWidth, saturation, cPrim)
	pctText := fmt.Sprintf("%.0f%% LOAD", saturation*100)

	inf := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(4).Render(spin),
		lipgloss.NewStyle().Foreground(m.theme.Text).Width(20).Render(status),
		bar,
		lipgloss.NewStyle().Foreground(cPrim).Width(10).Align(lipgloss.Right).Render(pctText),
	)
	foot := border.Copy().Width(m.width - 2).Render(inf)

	return lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinHorizontal(lipgloss.Top, leftP, rightP), foot)
}

// ---------------------------------------------------------
// ENGINE LOGIC
// ---------------------------------------------------------

func NewEngine(cfg Config, mChan chan logMsg) *IkelosEngine {
	t := &http.Transport{
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 200,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DisableCompression:  false,
	}
	return &IkelosEngine{
		Config:    cfg,
		Client:    &http.Client{Transport: t, Timeout: cfg.Timeout},
		Semaphore: make(chan struct{}, cfg.Workers),
		MsgChan:   mChan,
	}
}

func (e *IkelosEngine) Log(text, lType string) {
	select {
	case e.MsgChan <- logMsg{Text: text, Type: lType}:
	default: // Don't block engine if UI is slow
	}
}

func (e *IkelosEngine) Run() {
	if !e.Config.SkipSitemap {
		go e.huntSitemap()
	}
	e.WaitGroup.Add(1)
	go e.crawlPage(e.Config.TargetURL.String(), 0)
	e.WaitGroup.Wait()
}

func (e *IkelosEngine) huntSitemap() {
	sitemapURL := e.Config.TargetURL.Scheme + "://" + e.Config.TargetURL.Host + "/sitemap.xml"
	e.Log("Checking: "+sitemapURL, "INFO")

	data, code := e.fetch(sitemapURL, "")
	if code == 200 {
		e.Log("Sitemap Found! Injecting...", "HUNT")
		re := regexp.MustCompile(`<loc>(.*?)</loc>`)
		matches := re.FindAllStringSubmatch(string(data), -1)
		count := 0
		for _, m := range matches {
			if len(m) > 1 {
				urlStr := m[1]
				if strings.Contains(urlStr, e.Config.TargetURL.Host) {
					e.WaitGroup.Add(1)
					go e.crawlPage(urlStr, 1)
					count++
				}
			}
		}
		e.Log(fmt.Sprintf("Injected %d URLs", count), "HUNT")
	}
}

func (e *IkelosEngine) crawlPage(target string, depth int) {
	defer e.WaitGroup.Done()
	if depth > e.Config.MaxDepth {
		return
	}
	if strings.Contains(target, "#") {
		target = strings.Split(target, "#")[0]
	}

	if _, loaded := e.Visited.LoadOrStore(target, true); loaded {
		return
	}

	if e.Config.Mode == ModeShadow {
		time.Sleep(time.Duration(rand.Intn(800)+200) * time.Millisecond)
	}

	e.Semaphore <- struct{}{} // Worker START
	e.Log("Indexing: "+filepath.Base(target), "INFO")

	var bodyBytes []byte
	var statusCode int
	for retries := 0; retries < 3; retries++ {
		bodyBytes, statusCode = e.fetch(target, "")
		if statusCode == 200 {
			break
		}
		time.Sleep(time.Duration(math.Pow(2, float64(retries))) * 100 * time.Millisecond)
	}
	<-e.Semaphore // Worker STOP

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
				e.WaitGroup.Add(1)
				go e.crawlPage(absURL, depth+1)
			}
		}
	})
}

func (e *IkelosEngine) processAssetsAndRewrite(doc *goquery.Document, pageURL string) {
	doc.Find("img, source, link, script").Each(func(_ int, s *goquery.Selection) {
		attrs := []string{"src", "href", "srcset", "data-src", "data-original"}
		for _, attr := range attrs {
			val, exists := s.Attr(attr)
			if exists && val != "" {
				rawVal := strings.Split(val, " ")[0]
				absURL := e.resolveURL(rawVal, pageURL)
				if absURL != "" {
					e.downloadAsync(absURL, pageURL)
					s.SetAttr(attr, e.getRelPath(pageURL, absURL))
					if attr == "data-src" || attr == "data-original" {
						s.SetAttr("src", e.getRelPath(pageURL, absURL))
						s.RemoveAttr("loading")
					}
				}
			}
		}
	})

	doc.Find("*").Each(func(_ int, s *goquery.Selection) {
		style, exists := s.Attr("style")
		if exists && strings.Contains(style, "url") {
			s.SetAttr("style", e.rewriteCSSString(style, pageURL))
		}
	})
}

func (e *IkelosEngine) downloadAsync(url, referer string) {
	e.WaitGroup.Add(1)
	go e.downloadAsset(url, referer)
}

func (e *IkelosEngine) downloadAsset(target, referer string) {
	defer e.WaitGroup.Done()
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

var cssUrlRegex = regexp.MustCompile(`url\(['"]?([^'"\)]+)['"]?\)`)

func (e *IkelosEngine) rewriteCSSString(content, baseURL string) string {
	return cssUrlRegex.ReplaceAllStringFunc(content, func(match string) string {
		subMatch := cssUrlRegex.FindStringSubmatch(match)
		if len(subMatch) < 2 {
			return match
		}
		rawUrl := subMatch[1]
		if strings.HasPrefix(rawUrl, "data:") {
			return match
		}
		absURL := e.resolveURL(rawUrl, baseURL)
		if absURL == "" {
			return match
		}
		e.downloadAsync(absURL, baseURL)
		relPath := e.getRelPath(baseURL, absURL)
		return fmt.Sprintf("url('%s')", strings.ReplaceAll(relPath, "\\", "/"))
	})
}

func (e *IkelosEngine) fetch(target, referer string) ([]byte, int) {
	req, _ := http.NewRequest("GET", target, nil)
	uas := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/122.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/122.0.0.0 Safari/537.36",
	}
	req.Header.Set("User-Agent", uas[rand.Intn(len(uas))])
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := e.Client.Do(req)
	if err != nil {
		e.Log("Failed: "+filepath.Base(target), "ERROR")
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
		safeQuery := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(u.RawQuery, "_")
		if len(safeQuery) > 20 {
			safeQuery = safeQuery[:20]
		}
		ext := filepath.Ext(path)
		name := strings.TrimSuffix(path, ext)
		path = name + "_" + safeQuery + ext
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
	flag.Parse()

	if *urlStr == "" {
		p := tea.NewProgram(initialHelpModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	targetURL, err := url.Parse(*urlStr)
	if err != nil {
		fmt.Println("Invalid URL")
		return
	}

	mode := Mode(strings.ToUpper(*modeFlag))
	timeout := 30 * time.Second
	if mode == ModeBlitz {
		timeout = 10 * time.Second
	}

	cfg := Config{
		TargetURL: targetURL, OutputDir: *outFlag, Mode: mode, MaxDepth: *depthFlag,
		Workers: *workersFlag, Timeout: timeout, SkipSitemap: *noSitemap,
	}

	msgChan := make(chan logMsg, 2000)
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
	fmt.Printf("\n%s\n", c(fmt.Sprintf("[+] OPERATION COMPLETE. %d files secured.", atomic.LoadUint64(&engine.Files))))
}
