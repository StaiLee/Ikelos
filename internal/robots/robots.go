// Package robots implements a pragmatic robots.txt matcher close to Google's
// spec: User-agent grouping, Allow/Disallow, "*" wildcards, "$" end-anchors,
// and longest-match-wins with Allow breaking ties. It is deliberately
// dependency-free and pure so it can be unit-tested exhaustively.
package robots

import (
	"strings"
)

type rule struct {
	pattern string
	allow   bool
}

// Rules is a parsed robots.txt, indexed by the (lower-cased) user-agent token
// each group applies to.
type Rules struct {
	groups map[string][]rule
}

// Parse reads a robots.txt body. A nil or empty body yields a permissive
// ruleset (everything allowed).
func Parse(data []byte) *Rules {
	r := &Rules{groups: map[string][]rule{}}
	var agents []string // agents the current group applies to
	sawRuleForGroup := false

	for _, raw := range strings.Split(string(data), "\n") {
		line := raw
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, val, ok := splitField(line)
		if !ok {
			continue
		}
		switch key {
		case "user-agent":
			// A User-agent line after rules starts a fresh group.
			if sawRuleForGroup {
				agents = nil
				sawRuleForGroup = false
			}
			agents = append(agents, strings.ToLower(val))
		case "disallow", "allow":
			if len(agents) == 0 {
				continue // rule before any User-agent: ignore
			}
			sawRuleForGroup = true
			for _, a := range agents {
				r.groups[a] = append(r.groups[a], rule{pattern: val, allow: key == "allow"})
			}
		}
	}
	return r
}

func splitField(line string) (key, val string, ok bool) {
	i := strings.IndexByte(line, ':')
	if i < 0 {
		return "", "", false
	}
	return strings.ToLower(strings.TrimSpace(line[:i])), strings.TrimSpace(line[i+1:]), true
}

// Allowed reports whether userAgent may fetch path (the URL path, e.g.
// "/private/x"). Selection picks the most specific matching group, falling back
// to "*"; within a group the longest matching rule wins, Allow breaking ties.
func (r *Rules) Allowed(userAgent, path string) bool {
	if r == nil || len(r.groups) == 0 {
		return true
	}
	if path == "" {
		path = "/"
	}
	rules := r.selectGroup(userAgent)
	if len(rules) == 0 {
		return true
	}

	bestLen := -1
	allowed := true
	for _, ru := range rules {
		if !ruleMatch(ru.pattern, path) {
			continue
		}
		l := len(ru.pattern)
		if l > bestLen || (l == bestLen && ru.allow) {
			bestLen = l
			allowed = ru.allow
		}
	}
	return allowed
}

// selectGroup returns the rule set for the group whose token best matches the
// user-agent, or the "*" group otherwise.
func (r *Rules) selectGroup(userAgent string) []rule {
	ua := strings.ToLower(userAgent)
	best := ""
	bestLen := -1
	for agent := range r.groups {
		if agent == "*" {
			continue
		}
		if strings.Contains(ua, agent) && len(agent) > bestLen {
			best = agent
			bestLen = len(agent)
		}
	}
	if best != "" {
		return r.groups[best]
	}
	return r.groups["*"]
}

// ruleMatch applies robots path-matching semantics: patterns are prefix matches
// unless anchored with "$"; "*" matches any run of characters.
func ruleMatch(pattern, path string) bool {
	if pattern == "" {
		return false // empty Disallow/Allow matches nothing
	}
	if strings.HasSuffix(pattern, "$") {
		return wildcard(strings.TrimSuffix(pattern, "$"), path)
	}
	// Non-anchored patterns match a prefix of the path.
	return wildcard(pattern+"*", path)
}

// wildcard reports whether s matches pattern, where '*' matches any (possibly
// empty) run of characters. Iterative backtracking, O(len(pattern)*len(s)) worst
// case, no recursion.
func wildcard(pattern, s string) bool {
	var px, sx int
	nextPx, nextSx := -1, -1
	for sx < len(s) || px < len(pattern) {
		if px < len(pattern) {
			c := pattern[px]
			if c == '*' {
				nextPx, nextSx = px, sx+1
				px++
				continue
			}
			if sx < len(s) && s[sx] == c {
				px++
				sx++
				continue
			}
		}
		if nextSx >= 0 && nextSx <= len(s) {
			px, sx = nextPx+1, nextSx
			nextSx++
			continue
		}
		return false
	}
	return true
}
