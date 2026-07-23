package rewrite

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// fakeSink records every asset it is asked to localize and returns a
// predictable local path so assertions are easy.
type fakeSink struct {
	seen []string
}

func (f *fakeSink) Localize(absURL, base string) string {
	f.seen = append(f.seen, absURL)
	return "local/" + lastSeg(absURL)
}

func lastSeg(u string) string {
	if i := strings.LastIndexByte(u, '/'); i >= 0 {
		return u[i+1:]
	}
	return u
}

func TestHTMLRewritesAssets(t *testing.T) {
	html := `<html><body>
		<img src="/img/a.png">
		<img data-src="/img/lazy.png">
		<script src="app.js"></script>
		<link href="style.css">
	</body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sink := &fakeSink{}
	HTML(doc, "https://example.com/page", sink)

	want := map[string]bool{
		"https://example.com/img/a.png":    true,
		"https://example.com/img/lazy.png": true,
		"https://example.com/app.js":       true,
		"https://example.com/style.css":    true,
	}
	if len(sink.seen) != len(want) {
		t.Fatalf("registered %d assets, want %d: %v", len(sink.seen), len(want), sink.seen)
	}
	for _, u := range sink.seen {
		if !want[u] {
			t.Errorf("unexpected asset registered: %q", u)
		}
	}
	out, _ := doc.Html()
	if strings.Contains(out, `src="/img/a.png"`) {
		t.Error("img src was not localized")
	}
}

func TestSrcsetPreservesDescriptors(t *testing.T) {
	sink := &fakeSink{}
	got := Srcset("/a.png 1x, /b.png 2x", "https://example.com/", sink)
	if !strings.Contains(got, "1x") || !strings.Contains(got, "2x") {
		t.Errorf("descriptors lost: %q", got)
	}
	if len(sink.seen) != 2 {
		t.Errorf("registered %d assets, want 2", len(sink.seen))
	}
}

func TestCSSUrlAndImport(t *testing.T) {
	sink := &fakeSink{}
	css := `@import "base.css";
		body { background: url('/img/bg.png'); }
		div { background: url(icon.svg); }
		p { background: url(data:image/png;base64,AAAA); }`
	got := CSS(css, "https://example.com/css/main.css", sink)

	if strings.Contains(got, `@import "base.css"`) == false {
		// It should still be an @import, but pointing at a local path.
		if !strings.Contains(got, "@import") {
			t.Errorf("@import removed entirely: %q", got)
		}
	}
	if strings.Contains(got, "@import") && strings.Contains(got, `)`) && strings.Contains(got, `@import "local/base.css")`) {
		t.Errorf("stray paren after @import: %q", got)
	}
	if !strings.Contains(got, "data:image/png") {
		t.Error("data URI should be left untouched")
	}
	// base.css, bg.png, icon.svg registered; data: skipped.
	if len(sink.seen) != 3 {
		t.Errorf("registered %d assets, want 3: %v", len(sink.seen), sink.seen)
	}
}
