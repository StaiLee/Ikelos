package main

import "testing"

// resolveURL has no engine state dependency, so a zero-value engine is enough.
func TestResolveURL(t *testing.T) {
	e := &IkelosEngine{}
	cases := []struct {
		name string
		href string
		base string
		want string
	}{
		{"relative file", "page.html", "https://example.com/dir/", "https://example.com/dir/page.html"},
		{"absolute path", "/assets/app.css", "https://example.com/dir/page.html", "https://example.com/assets/app.css"},
		{"already absolute", "https://cdn.example.com/x.js", "https://example.com/", "https://cdn.example.com/x.js"},
		{"parent traversal", "../up.png", "https://example.com/a/b/", "https://example.com/a/up.png"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := e.resolveURL(c.href, c.base); got != c.want {
				t.Errorf("resolveURL(%q, %q) = %q, want %q", c.href, c.base, got, c.want)
			}
		})
	}
}
