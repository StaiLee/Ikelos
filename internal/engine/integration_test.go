package engine

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buildTestSite returns an httptest server serving a small interlinked site:
// two pages, a stylesheet referencing an image, two images, a robots.txt that
// forbids /private, and a sitemap seeding an otherwise-unlinked page.
func buildTestSite() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head>
			<link rel="stylesheet" href="/style.css">
		</head><body>
			<img src="/img/a.png">
			<a href="/about">About</a>
			<a href="/private/secret">Secret</a>
		</body></html>`))
	})
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>About</h1></body></html>`))
	})
	mux.HandleFunc("/sitemap-page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>From sitemap</h1></body></html>`))
	})
	mux.HandleFunc("/private/secret", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>TOP SECRET</body></html>`))
	})
	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(`body { background: url("/img/bg.png"); }`))
	})
	mux.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("\x89PNG\r\n\x1a\n fake image bytes"))
	})
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User-agent: *\nDisallow: /private/\n"))
	})
	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		// Build the <loc> from the live request host so no placeholder is needed.
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<?xml version="1.0"?><urlset><url><loc>http://` +
			r.Host + `/sitemap-page</loc></url></urlset>`))
	})

	return httptest.NewServer(mux)
}

func TestEngineMirrorsSite(t *testing.T) {
	srv := buildTestSite()
	defer srv.Close()

	base := srv.URL
	out := t.TempDir()
	target, _ := url.Parse(base)
	cfg := Config{
		TargetURL:    target,
		OutputDir:    out,
		Mode:         ModeMirror,
		MaxDepth:     2,
		Workers:      4,
		Timeout:      10 * time.Second,
		MaxBodyBytes: 1 << 20,
		MaxRetries:   3,
	}

	eng := New(cfg)
	// Drain logs so the buffered channel never blocks the crawl.
	go func() {
		for range eng.Logs() {
		}
	}()

	done := make(chan struct{})
	go func() { eng.Run(); close(done) }()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("engine did not finish within 20s (possible deadlock)")
	}

	host := strings.ReplaceAll(target.Host, ":", "_")
	mustExist := []string{
		filepath.Join(out, host, "index.html"),
		filepath.Join(out, host, "about.html"),
		filepath.Join(out, host, "sitemap-page.html"),
		filepath.Join(out, host, "style.css"),
		filepath.Join(out, host, "img", "a.png"),
		filepath.Join(out, host, "img", "bg.png"),
	}
	for _, p := range mustExist {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected file missing: %s (%v)", p, err)
		}
	}

	// robots.txt must have blocked the private page.
	if _, err := os.Stat(filepath.Join(out, host, "private", "secret.html")); err == nil {
		t.Error("private page was fetched despite robots.txt Disallow")
	}

	// The index must reference the stylesheet locally, not by absolute URL.
	index, err := os.ReadFile(filepath.Join(out, host, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(index), `href="/style.css"`) || strings.Contains(string(index), base+"/style.css") {
		t.Error("stylesheet link was not localized in index.html")
	}
	if !strings.Contains(string(index), "style.css") {
		t.Error("index.html lost its stylesheet reference entirely")
	}

	// The stylesheet's url() must have been rewritten to the local image.
	css, err := os.ReadFile(filepath.Join(out, host, "style.css"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(css), `url("/img/bg.png")`) {
		t.Error("CSS url() was not localized")
	}
	if !strings.Contains(string(css), "bg.png") {
		t.Error("CSS lost its background image reference")
	}
}

func TestEngineCancel(t *testing.T) {
	srv := buildTestSite()
	defer srv.Close()

	target, _ := url.Parse(srv.URL)
	eng := New(Config{
		TargetURL:    target,
		OutputDir:    t.TempDir(),
		Mode:         ModeMirror,
		MaxDepth:     5,
		Workers:      2,
		Timeout:      10 * time.Second,
		MaxBodyBytes: 1 << 20,
		MaxRetries:   3,
	})
	go func() {
		for range eng.Logs() {
		}
	}()

	done := make(chan struct{})
	go func() { eng.Run(); close(done) }()
	eng.Cancel() // cancel almost immediately

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("engine did not stop after Cancel")
	}
}
