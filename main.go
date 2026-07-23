// Command ikelos is a concurrent website mirror / "reality shifter". It fetches
// a target site, rewrites every asset and link to point at local copies, and
// renders a live Bubble Tea dashboard while it works.
//
// See README.md for the full field manual. main.go is intentionally thin: it
// parses flags, resolves configuration, and wires the engine to the TUI.
package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ikelos/internal/engine"
	"ikelos/internal/tui"
)

const (
	defaultMaxBody = 64 << 20 // 64 MiB per resource
	defaultRetries = 3
)

func main() {
	var (
		urlStr       = flag.String("url", "", "Target URL")
		outFlag      = flag.String("out", "./cloned_site", "Output directory")
		modeFlag     = flag.String("mode", "MIRROR", "Mode: MIRROR, SHADOW, BLITZ")
		proxyFlag    = flag.String("proxy", "", "Proxy URL (e.g. http://127.0.0.1:8080)")
		uaFlag       = flag.String("ua", "", "Custom User-Agent (overrides rotation)")
		workersFlag  = flag.Int("threads", 20, "Concurrent workers")
		depthFlag    = flag.Int("depth", 2, "Recursion depth")
		noSitemap    = flag.Bool("nositemap", false, "Disable sitemap hunting")
		ignoreRobots = flag.Bool("ignore-robots", false, "Ignore robots.txt (crawl everything)")
		insecure     = flag.Bool("insecure", true, "Skip TLS certificate verification")
	)
	flag.Parse()

	// No URL: show the interactive manual and exit.
	if *urlStr == "" {
		if _, err := tea.NewProgram(tui.NewHelp(), tea.WithAltScreen()).Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error starting manual:", err)
			os.Exit(1)
		}
		return
	}

	target, err := normalizeURL(*urlStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid URL:", err)
		os.Exit(2)
	}

	mode := engine.ParseMode(*modeFlag)
	timeout := 45 * time.Second
	if mode == engine.ModeBlitz {
		timeout = 15 * time.Second
	}

	workers := *workersFlag
	if workers < 1 {
		workers = 1
	}

	cfg := engine.Config{
		TargetURL:    target,
		OutputDir:    *outFlag,
		Mode:         mode,
		MaxDepth:     *depthFlag,
		Workers:      workers,
		Timeout:      timeout,
		SkipSitemap:  *noSitemap,
		IgnoreRobots: *ignoreRobots,
		Proxy:        *proxyFlag,
		UserAgent:    *uaFlag,
		InsecureTLS:  *insecure,
		MaxBodyBytes: defaultMaxBody,
		MaxRetries:   defaultRetries,
	}

	eng := engine.New(cfg)
	done := make(chan bool, 1)
	go func() {
		eng.Run()
		done <- true
	}()

	program := tea.NewProgram(tui.NewDashboard(cfg, eng, done), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	// If the UI exited before the crawl (user pressed Q), make sure the engine
	// is torn down so its goroutines stop.
	eng.Cancel()

	stats := eng.Snapshot()
	ok := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).Render
	fmt.Printf("\n%s\n", ok(fmt.Sprintf("[+] ARCHIVE COMPLETE. %d files downloaded (%.1f MB).",
		stats.Files, float64(stats.Bytes)/1024/1024)))
	if stats.Errors > 0 {
		fmt.Printf("[!] Encountered %d errors during the run.\n", stats.Errors)
	}
}

// normalizeURL parses a user-supplied target, defaulting to https:// when no
// scheme is given, and validates it is an http(s) URL with a host.
func normalizeURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("missing host")
	}
	return u, nil
}
