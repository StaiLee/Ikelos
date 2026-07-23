package urlx

import "testing"

func TestResolve(t *testing.T) {
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
		{"strips fragment", "page.html#section", "https://example.com/", "https://example.com/page.html"},
		{"protocol relative", "//cdn.example.com/x.js", "https://example.com/", "https://cdn.example.com/x.js"},
		{"pure fragment rejected", "#top", "https://example.com/", ""},
		{"mailto rejected", "mailto:a@b.com", "https://example.com/", ""},
		{"javascript rejected", "javascript:void(0)", "https://example.com/", ""},
		{"data uri rejected", "data:image/png;base64,AAAA", "https://example.com/", ""},
		{"empty rejected", "", "https://example.com/", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Resolve(c.href, c.base); got != c.want {
				t.Errorf("Resolve(%q, %q) = %q, want %q", c.href, c.base, got, c.want)
			}
		})
	}
}

func TestSameHost(t *testing.T) {
	cases := []struct {
		url, host string
		want      bool
	}{
		{"https://example.com/a", "example.com", true},
		{"https://example.com:443/a", "example.com", true},
		{"https://example.com/a", "example.com:8080", true},
		{"https://cdn.example.com/a", "example.com", false},
		{"https://EXAMPLE.com/a", "example.com", true},
	}
	for _, c := range cases {
		if got := SameHost(c.url, c.host); got != c.want {
			t.Errorf("SameHost(%q, %q) = %v, want %v", c.url, c.host, got, c.want)
		}
	}
}
