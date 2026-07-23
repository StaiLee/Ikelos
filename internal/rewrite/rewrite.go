// Package rewrite performs the "surgery" on fetched documents: it points every
// asset reference at its local copy and hands back the list of assets that must
// be downloaded. It holds no crawler state — all outside effects go through the
// Sink interface — which keeps the DOM/CSS logic pure and unit-testable.
package rewrite

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"ikelos/internal/urlx"
)

// Sink receives assets discovered during rewriting and answers where their
// local copy will live. Localize must register absURL for download and return
// the relative path (from base's document) to use in the rewritten reference.
type Sink interface {
	Localize(absURL, base string) string
}

// elementAttrs lists the (selector, attributes) pairs that can carry a fetchable
// asset reference. data-src/data-original cover the lazy-loading case.
var attrCandidates = []string{"src", "href", "data-src", "data-original", "poster"}

// HTML rewrites every asset reference in doc so it points at a local file,
// registering each asset with the sink. pageURL is the absolute URL the
// document was fetched from.
func HTML(doc *goquery.Document, pageURL string, sink Sink) {
	doc.Find("img, source, link, script, video, audio, iframe, embed").Each(func(_ int, s *goquery.Selection) {
		for _, attr := range attrCandidates {
			val, ok := s.Attr(attr)
			if !ok || val == "" {
				continue
			}
			abs := urlx.Resolve(val, pageURL)
			if abs == "" {
				continue
			}
			s.SetAttr(attr, sink.Localize(abs, pageURL))
		}
		if srcset, ok := s.Attr("srcset"); ok && srcset != "" {
			s.SetAttr("srcset", Srcset(srcset, pageURL, sink))
		}
	})

	// Inline style="..." attributes may reference url(...) assets.
	doc.Find("[style]").Each(func(_ int, s *goquery.Selection) {
		style, _ := s.Attr("style")
		if strings.Contains(style, "url") {
			s.SetAttr("style", CSS(style, pageURL, sink))
		}
	})
}

// Srcset rewrites a responsive srcset attribute, preserving each descriptor
// (the "2x" / "640w" suffix) while localizing the URL.
func Srcset(srcset, pageURL string, sink Sink) string {
	parts := strings.Split(srcset, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) == 0 {
			continue
		}
		abs := urlx.Resolve(fields[0], pageURL)
		if abs == "" {
			out = append(out, strings.TrimSpace(part))
			continue
		}
		rel := sink.Localize(abs, pageURL)
		if len(fields) > 1 {
			rel += " " + strings.Join(fields[1:], " ")
		}
		out = append(out, rel)
	}
	return strings.Join(out, ", ")
}

// cssURL captures both url(...) references and @import "..."/@import url(...)
// forms. Group 1 = url(...) target, group 2 = bare @import target.
var cssURL = regexp.MustCompile(`url\(\s*['"]?([^'")]+?)['"]?\s*\)|@import\s+['"]([^'"]+)['"]`)

// CSS rewrites url(...) and @import references in a stylesheet (or inline style
// string). baseURL is the URL of the stylesheet itself, so relative asset paths
// resolve correctly.
func CSS(content, baseURL string, sink Sink) string {
	return cssURL.ReplaceAllStringFunc(content, func(match string) string {
		sub := cssURL.FindStringSubmatch(match)
		raw := sub[1]
		isImport := false
		if raw == "" {
			raw = sub[2]
			isImport = true
		}
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "data:") {
			return match
		}
		abs := urlx.Resolve(raw, baseURL)
		if abs == "" {
			return match
		}
		rel := strings.ReplaceAll(sink.Localize(abs, baseURL), "\\", "/")
		if isImport {
			return `@import "` + rel + `"`
		}
		return `url("` + rel + `")`
	})
}
