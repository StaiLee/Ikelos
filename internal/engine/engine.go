// Package engine is Ikelos's crawling core: a bounded worker pool that fetches,
// rewrites and stores a website mirror. It exposes cancellation, pause/resume
// and a lock-free stats snapshot for the TUI, and depends only on the small
// leaf packages (urlx, robots, store, rewrite) so its behaviour is testable in
// isolation.
package engine

import (
	"bytes"
	"context"
	"crypto/tls"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
	"ikelos/internal/rewrite"
	"ikelos/internal/robots"
	"ikelos/internal/store"
	"ikelos/internal/urlx"
)

// Product identity, surfaced by the TUI and the final summary.
const (
	Name    = "IKELOS"
	Version = "8.0.0 (THE SURGEON)"
)

// Mode selects an operational profile.
type Mode string

const (
	ModeMirror Mode = "MIRROR"
	ModeShadow Mode = "SHADOW"
	ModeBlitz  Mode = "BLITZ"
)

// ParseMode normalises a user-supplied mode string, defaulting to MIRROR.
func ParseMode(s string) Mode {
	switch Mode(strings.ToUpper(strings.TrimSpace(s))) {
	case ModeShadow:
		return ModeShadow
	case ModeBlitz:
		return ModeBlitz
	default:
		return ModeMirror
	}
}

// Config is the fully-resolved runtime configuration.
type Config struct {
	TargetURL    *url.URL
	OutputDir    string
	Mode         Mode
	MaxDepth     int
	Workers      int
	Timeout      time.Duration
	SkipSitemap  bool
	IgnoreRobots bool
	Proxy        string
	UserAgent    string
	InsecureTLS  bool
	MaxBodyBytes int64
	MaxRetries   int
}

// LogLevel classifies a live-feed message for colouring in the TUI.
type LogLevel int

const (
	LogInfo LogLevel = iota
	LogSuccess
	LogError
	LogAsset
	LogHunt
	LogWarn
)

// LogEntry is one line of the live feed.
type LogEntry struct {
	Text  string
	Level LogLevel
}

type stats struct {
	files    uint64
	bytes    uint64
	errors   uint64
	typeIMG  uint64
	typeHTML uint64
	typeCODE uint64
	active   int64
}

// Stats is an immutable snapshot for the UI.
type Stats struct {
	Files, Bytes, Errors uint64
	IMG, HTML, CODE      uint64
	Active, Queue        int
}

// Engine owns the crawl. Construct with New, then call Run (blocking).
type Engine struct {
	cfg    Config
	client *http.Client
	disp   *dispatcher
	logs   chan LogEntry

	ctx    context.Context
	cancel context.CancelFunc

	visited sync.Map // url string -> struct{}
	robots  *robots.Rules
	store   *store.Mapper

	paused atomic.Bool
	stats  stats
}

// New builds an Engine from cfg. It does not start any work.
func New(cfg Config) *Engine {
	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 256,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: cfg.InsecureTLS},
	}
	if cfg.Proxy != "" {
		if proxyURL, err := url.Parse(cfg.Proxy); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		cfg:    cfg,
		client: &http.Client{Transport: transport, Timeout: cfg.Timeout},
		disp:   newDispatcher(),
		logs:   make(chan LogEntry, 4096),
		ctx:    ctx,
		cancel: cancel,
		store:  store.New(cfg.OutputDir),
	}
}

// Logs is the live-feed stream. The TUI consumes it; it is never closed so a
// slow consumer simply drops messages (Log is non-blocking).
func (e *Engine) Logs() <-chan LogEntry { return e.logs }

// Cancel aborts the crawl. In-flight requests are interrupted and workers exit.
func (e *Engine) Cancel() {
	e.cancel()
	e.disp.shutdown()
}

// SetPaused freezes or resumes the crawl. Paused workers hold their slot but do
// no network I/O.
func (e *Engine) SetPaused(p bool) { e.paused.Store(p) }

// Snapshot returns current counters for display.
func (e *Engine) Snapshot() Stats {
	return Stats{
		Files:  atomic.LoadUint64(&e.stats.files),
		Bytes:  atomic.LoadUint64(&e.stats.bytes),
		Errors: atomic.LoadUint64(&e.stats.errors),
		IMG:    atomic.LoadUint64(&e.stats.typeIMG),
		HTML:   atomic.LoadUint64(&e.stats.typeHTML),
		CODE:   atomic.LoadUint64(&e.stats.typeCODE),
		Active: int(atomic.LoadInt64(&e.stats.active)),
		Queue:  e.disp.outstanding(),
	}
}

// Run seeds the crawl and blocks until it completes or is cancelled.
func (e *Engine) Run() {
	if !e.cfg.IgnoreRobots {
		e.loadRobots()
	}
	if !e.cfg.SkipSitemap {
		e.huntSitemap()
	}

	e.enqueue(job{url: e.cfg.TargetURL.String(), depth: 0, kind: jobCrawl})

	var wg sync.WaitGroup
	for i := 0; i < e.cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				j, ok := e.disp.pop()
				if !ok {
					return
				}
				e.runJob(j)
				e.disp.complete()
			}
		}()
	}
	wg.Wait()
}

// enqueue submits a job unless its URL was already seen. Deduplication happens
// here so the dispatcher's pending counter stays exact.
func (e *Engine) enqueue(j job) {
	if e.ctx.Err() != nil {
		return
	}
	if _, loaded := e.visited.LoadOrStore(j.url, struct{}{}); loaded {
		return
	}
	e.disp.submit(j)
}

// runJob processes one job, guaranteeing it never panics out (which would leak
// the dispatcher's pending count and hang the crawl).
func (e *Engine) runJob(j job) {
	defer func() {
		if r := recover(); r != nil {
			atomic.AddUint64(&e.stats.errors, 1)
			e.log(LogError, "recovered from panic")
		}
	}()

	e.checkPause()
	if e.ctx.Err() != nil {
		return
	}
	atomic.AddInt64(&e.stats.active, 1)
	defer atomic.AddInt64(&e.stats.active, -1)

	switch j.kind {
	case jobCrawl:
		e.crawl(j)
	case jobAsset:
		e.download(j)
	}
}

func (e *Engine) crawl(j job) {
	if j.depth > e.cfg.MaxDepth {
		return
	}
	if e.robots != nil {
		if u, err := url.Parse(j.url); err == nil && !e.robots.Allowed(botToken, u.Path) {
			e.log(LogWarn, "Blocked by robots.txt: "+shortName(j.url))
			return
		}
	}
	if e.cfg.Mode == ModeShadow {
		e.jitter()
	}

	e.log(LogInfo, "Crawling: "+shortName(j.url))
	res := e.fetch(j.url, j.referer)
	if res.body == nil || res.status != http.StatusOK {
		return
	}

	// A "page" URL that turns out to be a binary/asset is saved as-is.
	if !isHTML(res.contentType) {
		e.saveAsset(j.url, res.body, res.contentType)
		return
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(res.body))
	if err != nil {
		return
	}
	rewrite.HTML(doc, j.url, e)

	out, err := doc.Html()
	if err != nil {
		return
	}
	if err := e.store.Save(j.url, []byte(out)); err != nil {
		atomic.AddUint64(&e.stats.errors, 1)
		e.log(LogError, "save failed: "+err.Error())
		return
	}
	atomic.AddUint64(&e.stats.files, 1)
	atomic.AddUint64(&e.stats.typeHTML, 1)
	e.log(LogSuccess, shortName(j.url))

	e.followLinks(doc, j)
}

// followLinks enqueues in-scope anchors: page-like URLs as deeper crawls,
// everything else as assets.
func (e *Engine) followLinks(doc *goquery.Document, parent job) {
	host := e.cfg.TargetURL.Host
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		abs := urlx.Resolve(href, parent.url)
		if abs == "" || !urlx.SameHost(abs, host) {
			return
		}
		if isPageURL(abs) {
			if parent.depth < e.cfg.MaxDepth {
				e.enqueue(job{url: abs, referer: parent.url, depth: parent.depth + 1, kind: jobCrawl})
			}
		} else {
			e.enqueue(job{url: abs, referer: parent.url, kind: jobAsset})
		}
	})
}

func (e *Engine) download(j job) {
	res := e.fetch(j.url, j.referer)
	if res.body == nil || res.status != http.StatusOK {
		return
	}
	e.saveAsset(j.url, res.body, res.contentType)
}

// saveAsset classifies, optionally rewrites (CSS), and stores a downloaded
// asset. Localizing a stylesheet enqueues the assets it references.
func (e *Engine) saveAsset(target string, body []byte, contentType string) {
	switch {
	case isCSS(contentType, target):
		body = []byte(rewrite.CSS(string(body), target, e))
		atomic.AddUint64(&e.stats.typeCODE, 1)
	case isJS(contentType, target):
		atomic.AddUint64(&e.stats.typeCODE, 1)
	default:
		atomic.AddUint64(&e.stats.typeIMG, 1)
	}
	if err := e.store.Save(target, body); err != nil {
		atomic.AddUint64(&e.stats.errors, 1)
		e.log(LogError, "save failed: "+err.Error())
		return
	}
	atomic.AddUint64(&e.stats.files, 1)
	e.log(LogAsset, shortName(target))
}

// Localize implements rewrite.Sink: it registers an asset for download and
// returns the on-disk relative path to reference it by.
func (e *Engine) Localize(absURL, base string) string {
	e.enqueue(job{url: absURL, referer: base, kind: jobAsset})
	return e.store.RelPath(base, absURL)
}

func (e *Engine) loadRobots() {
	robotsURL := e.cfg.TargetURL.Scheme + "://" + e.cfg.TargetURL.Host + "/robots.txt"
	res := e.fetch(robotsURL, "")
	if res.status == http.StatusOK && res.body != nil {
		e.log(LogInfo, "Parsing robots.txt")
		e.robots = robots.Parse(res.body)
	}
}

var locRegex = regexp.MustCompile(`<loc>\s*(.*?)\s*</loc>`)

// huntSitemap fetches sitemap.xml (following one level of sitemap-index
// nesting) and seeds in-scope pages. Done synchronously before the pool starts,
// so seeds are counted before any worker can drain the queue.
func (e *Engine) huntSitemap() {
	root := e.cfg.TargetURL.Scheme + "://" + e.cfg.TargetURL.Host + "/sitemap.xml"
	e.log(LogHunt, "Scanning sitemap: "+root)
	count := e.ingestSitemap(root, true)
	if count > 0 {
		e.log(LogHunt, "Sitemap seeded "+itoa(count)+" pages")
	}
}

func (e *Engine) ingestSitemap(sitemapURL string, allowNested bool) int {
	res := e.fetch(sitemapURL, "")
	if res.status != http.StatusOK || res.body == nil {
		return 0
	}
	count := 0
	for _, m := range locRegex.FindAllStringSubmatch(string(res.body), -1) {
		loc := strings.TrimSpace(m[1])
		if loc == "" {
			continue
		}
		if allowNested && strings.HasSuffix(strings.ToLower(loc), ".xml") {
			count += e.ingestSitemap(loc, false) // one level of nesting only
			continue
		}
		if urlx.SameHost(loc, e.cfg.TargetURL.Host) {
			e.enqueue(job{url: loc, depth: 1, kind: jobCrawl})
			count++
		}
	}
	return count
}

func (e *Engine) log(level LogLevel, text string) {
	select {
	case e.logs <- LogEntry{Text: text, Level: level}:
	default: // drop rather than block the crawl on a slow UI
	}
}

func (e *Engine) checkPause() {
	for e.paused.Load() {
		select {
		case <-e.ctx.Done():
			return
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func (e *Engine) jitter() {
	d := time.Duration(rand.Intn(1000)+400) * time.Millisecond
	select {
	case <-e.ctx.Done():
	case <-time.After(d):
	}
}

// ---- content-type helpers -------------------------------------------------

func isHTML(contentType string) bool {
	ct := strings.ToLower(contentType)
	// Empty content-type: assume HTML (we only reach here for crawl jobs, whose
	// URLs were selected as page-like).
	return contentType == "" || strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml")
}

func isCSS(contentType, target string) bool {
	if strings.Contains(strings.ToLower(contentType), "text/css") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(pathOnly(target)), ".css")
}

func isJS(contentType, target string) bool {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "javascript") || strings.Contains(ct, "ecmascript") {
		return true
	}
	ext := strings.ToLower(path.Ext(pathOnly(target)))
	return ext == ".js" || ext == ".mjs"
}

var pageExts = map[string]bool{
	"": true, ".html": true, ".htm": true, ".php": true,
	".asp": true, ".aspx": true, ".jsp": true, ".xhtml": true,
}

func isPageURL(rawURL string) bool {
	return pageExts[strings.ToLower(path.Ext(pathOnly(rawURL)))]
}

func pathOnly(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil {
		return u.Path
	}
	if i := strings.IndexAny(rawURL, "?#"); i >= 0 {
		return rawURL[:i]
	}
	return rawURL
}

func shortName(rawURL string) string {
	p := pathOnly(rawURL)
	base := path.Base(p)
	if base == "" || base == "/" || base == "." {
		if u, err := url.Parse(rawURL); err == nil {
			return u.Host + "/"
		}
	}
	return base
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
