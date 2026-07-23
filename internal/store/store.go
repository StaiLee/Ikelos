// Package store maps remote URLs to safe on-disk paths and writes files. The
// mapping is deterministic (same URL -> same path) so that link rewriting and
// asset saving always agree, and hardened against the two classic failure modes
// of a naive web mirror: filesystem-forbidden characters and path traversal
// outside the output directory.
package store

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maxSegmentLen = 100 // per path component, before MD5 collapse
	maxNameKeep   = 50  // kept prefix when appending a query hash
)

// forbidden matches characters illegal in Windows path segments (and generally
// unsafe). The forward slash is handled separately as the segment separator.
var forbidden = regexp.MustCompile(`[<>:"|?*\x00-\x1f]`)

// Mapper turns URLs into paths rooted at OutputDir.
type Mapper struct {
	OutputDir string
}

// New returns a Mapper writing under outputDir.
func New(outputDir string) *Mapper { return &Mapper{OutputDir: outputDir} }

// PathFor returns the absolute on-disk path for rawURL. The result is always
// contained within OutputDir; a traversal attempt yields an error rather than
// escaping the sandbox.
func (m *Mapper) PathFor(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse %q: %w", rawURL, err)
	}

	host := sanitizeSegment(u.Host)
	if host == "" {
		host = "unknown_host"
	}

	path := u.Path
	switch {
	case path == "" || path == "/":
		path = "/index.html"
	case strings.HasSuffix(path, "/"):
		path += "index.html"
	case filepath.Ext(path) == "":
		path += ".html"
	}

	// Sanitize each segment independently so we never destroy the directory
	// structure, then collapse overly long segments to an MD5 digest.
	segs := strings.Split(path, "/")
	for i, s := range segs {
		if s == "" {
			continue
		}
		s = sanitizeSegment(s)
		if len(s) > maxSegmentLen {
			ext := filepath.Ext(s)
			s = shortHash(s) + ext
		}
		segs[i] = s
	}
	path = strings.Join(segs, "/")

	// Distinct query strings must map to distinct files.
	if u.RawQuery != "" {
		ext := filepath.Ext(path)
		name := strings.TrimSuffix(path, ext)
		if len(name) > maxNameKeep {
			name = name[:maxNameKeep]
		}
		path = fmt.Sprintf("%s_%s%s", name, shortHash(u.RawQuery)[:8], ext)
	}

	root, err := filepath.Abs(m.OutputDir)
	if err != nil {
		return "", err
	}
	full := filepath.Clean(filepath.Join(root, host, filepath.FromSlash(path)))

	// Traversal guard: the cleaned path must stay under root.
	rel, err := filepath.Rel(root, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes output dir: %q", rawURL)
	}
	return full, nil
}

// RelPath returns the slash-separated relative path from the file for fromURL's
// directory to the file for toURL, suitable for embedding in HTML/CSS. On any
// error it falls back to the absolute toURL so the document still resolves
// online.
func (m *Mapper) RelPath(fromURL, toURL string) string {
	from, err := m.PathFor(fromURL)
	if err != nil {
		return toURL
	}
	to, err := m.PathFor(toURL)
	if err != nil {
		return toURL
	}
	rel, err := filepath.Rel(filepath.Dir(from), to)
	if err != nil {
		return toURL
	}
	return filepath.ToSlash(rel)
}

// Save writes data for rawURL, creating parent directories. It returns any I/O
// error instead of swallowing it, so callers can account for real failures.
func (m *Mapper) Save(rawURL string, data []byte) error {
	full, err := m.PathFor(rawURL)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return fmt.Errorf("mkdir for %q: %w", rawURL, err)
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		return fmt.Errorf("write %q: %w", rawURL, err)
	}
	return nil
}

func sanitizeSegment(s string) string {
	s = strings.ReplaceAll(s, ":", "_")
	return forbidden.ReplaceAllString(s, "_")
}

func shortHash(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
