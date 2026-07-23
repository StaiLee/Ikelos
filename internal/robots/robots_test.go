package robots

import "testing"

func TestAllowed(t *testing.T) {
	body := []byte(`
User-agent: *
Disallow: /private/
Disallow: /tmp
Allow: /private/public
Disallow: /*.pdf$

User-agent: Googlebot
Disallow: /
`)
	r := Parse(body)

	cases := []struct {
		ua, path string
		want     bool
	}{
		{"ikelos", "/", true},
		{"ikelos", "/index.html", true},
		{"ikelos", "/private/secret", false},
		{"ikelos", "/private/public/page", true}, // Allow is longer than Disallow
		{"ikelos", "/tmp", false},
		{"ikelos", "/tmpfile", false}, // prefix match
		{"ikelos", "/report.pdf", false},
		{"ikelos", "/report.pdf.html", true}, // $ anchor: only real .pdf endings
		{"Mozilla Googlebot/2.1", "/anything", false},
	}
	for _, c := range cases {
		if got := r.Allowed(c.ua, c.path); got != c.want {
			t.Errorf("Allowed(%q, %q) = %v, want %v", c.ua, c.path, got, c.want)
		}
	}
}

func TestEmptyAndNil(t *testing.T) {
	if !Parse(nil).Allowed("x", "/y") {
		t.Error("nil body should allow everything")
	}
	if !Parse([]byte("# just a comment\n")).Allowed("x", "/y") {
		t.Error("comment-only body should allow everything")
	}
	var r *Rules
	if !r.Allowed("x", "/y") {
		t.Error("nil *Rules should allow everything")
	}
}

func TestWildcard(t *testing.T) {
	cases := []struct {
		pat, s string
		want   bool
	}{
		{"abc", "abc", true},
		{"abc", "abcd", false},
		{"a*c", "abbbc", true},
		{"a*c", "ac", true},
		{"*", "anything", true},
		{"a*", "a", true},
		{"*z", "abz", true},
		{"*z", "abzq", false},
	}
	for _, c := range cases {
		if got := wildcard(c.pat, c.s); got != c.want {
			t.Errorf("wildcard(%q, %q) = %v, want %v", c.pat, c.s, got, c.want)
		}
	}
}
