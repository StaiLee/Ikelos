// Package urlx contains small, dependency-free URL helpers shared across the
// engine. Keeping them isolated makes the tricky resolution logic trivially
// testable without spinning up the crawler.
package urlx

import (
	"net/url"
	"strings"
)

// Resolve turns a possibly-relative href into an absolute URL, using base as
// the reference document. It returns "" when either input is unparseable or
// when the result is not an http(s) resource (mailto:, javascript:, tel:, ...),
// which the crawler must never try to fetch.
func Resolve(href, base string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	// Cheap rejects before touching the parser.
	switch {
	case strings.HasPrefix(href, "#"):
		return ""
	case strings.HasPrefix(href, "data:"),
		strings.HasPrefix(href, "javascript:"),
		strings.HasPrefix(href, "mailto:"),
		strings.HasPrefix(href, "tel:"),
		strings.HasPrefix(href, "blob:"):
		return ""
	}

	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	abs := baseURL.ResolveReference(ref)
	if abs.Scheme != "http" && abs.Scheme != "https" {
		return ""
	}
	// Drop the fragment: it never changes what we fetch or store.
	abs.Fragment = ""
	return abs.String()
}

// SameHost reports whether rawURL points at the given host (case-insensitive,
// port-insensitive). Used to decide whether a link is in-scope for crawling.
func SameHost(rawURL, host string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Hostname(), stripPort(host))
}

func stripPort(host string) string {
	if i := strings.IndexByte(host, ':'); i >= 0 {
		return host[:i]
	}
	return host
}
