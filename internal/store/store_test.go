package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathFor(t *testing.T) {
	m := New("/out")
	root, _ := filepath.Abs("/out")

	cases := []struct {
		url  string
		want string // relative to root, slash form
	}{
		{"https://example.com/", "example.com/index.html"},
		{"https://example.com/about", "example.com/about.html"},
		{"https://example.com/blog/", "example.com/blog/index.html"},
		{"https://example.com/assets/app.css", "example.com/assets/app.css"},
		{"https://example.com:8443/x.js", "example.com_8443/x.js"},
	}
	for _, c := range cases {
		got, err := m.PathFor(c.url)
		if err != nil {
			t.Fatalf("PathFor(%q) error: %v", c.url, err)
		}
		rel, _ := filepath.Rel(root, got)
		if filepath.ToSlash(rel) != c.want {
			t.Errorf("PathFor(%q) = %q, want %q", c.url, filepath.ToSlash(rel), c.want)
		}
	}
}

func TestPathForQueryDistinct(t *testing.T) {
	m := New("/out")
	a, _ := m.PathFor("https://example.com/api?page=1")
	b, _ := m.PathFor("https://example.com/api?page=2")
	if a == b {
		t.Errorf("distinct queries mapped to same path: %q", a)
	}
}

func TestPathForDeterministic(t *testing.T) {
	m := New("/out")
	a, _ := m.PathFor("https://example.com/a/b/c?x=1")
	b, _ := m.PathFor("https://example.com/a/b/c?x=1")
	if a != b {
		t.Errorf("mapping not deterministic: %q vs %q", a, b)
	}
}

func TestRelPath(t *testing.T) {
	m := New("/out")
	// From a page to a sibling asset.
	got := m.RelPath("https://example.com/blog/post", "https://example.com/img/x.png")
	if got != "../img/x.png" {
		t.Errorf("RelPath = %q, want ../img/x.png", got)
	}
	// From root index to a nested asset.
	got = m.RelPath("https://example.com/", "https://example.com/css/app.css")
	if got != "css/app.css" {
		t.Errorf("RelPath = %q, want css/app.css", got)
	}
}

func TestSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	m := New(dir)
	if err := m.Save("https://example.com/note.txt", []byte("hi")); err != nil {
		t.Fatalf("Save error: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "example.com", "note.txt"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != "hi" {
		t.Errorf("content = %q, want hi", got)
	}
}

func TestLongSegmentCollapses(t *testing.T) {
	m := New("/out")
	long := "https://example.com/" + strings.Repeat("a", 300) + ".html"
	got, err := m.PathFor(long)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(filepath.Base(got)) > 140 {
		t.Errorf("segment not collapsed: %d chars", len(filepath.Base(got)))
	}
}
