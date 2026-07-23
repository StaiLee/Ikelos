package engine

import (
	"context"
	"errors"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

// userAgents is the rotation pool used unless the operator pins one with -ua.
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
}

// botToken is the identity used for robots.txt matching. Keeping it distinct
// from the rotated browser user-agents means only "User-agent: *" rules apply,
// which is the predictable, well-behaved default.
const botToken = "ikelos"

type fetchResult struct {
	body        []byte
	status      int
	contentType string
}

func (e *Engine) userAgent() string {
	if e.cfg.UserAgent != "" {
		return e.cfg.UserAgent
	}
	return userAgents[rand.Intn(len(userAgents))]
}

// fetch retrieves target, retrying only on transient failures (network errors,
// 5xx, 429) with exponential backoff plus jitter. Permanent responses (4xx) and
// context cancellation stop immediately. A zero status means the request never
// completed.
func (e *Engine) fetch(target, referer string) fetchResult {
	var last fetchResult
	for attempt := 0; attempt < e.cfg.MaxRetries; attempt++ {
		if e.ctx.Err() != nil {
			return fetchResult{}
		}
		e.checkPause()

		res, transient := e.doRequest(target, referer)
		if res.status == http.StatusOK {
			return res
		}
		last = res
		if !transient {
			return res
		}
		select {
		case <-e.ctx.Done():
			return fetchResult{}
		case <-time.After(backoff(attempt)):
		}
	}
	return last
}

func (e *Engine) doRequest(target, referer string) (res fetchResult, transient bool) {
	req, err := http.NewRequestWithContext(e.ctx, http.MethodGet, target, nil)
	if err != nil {
		return fetchResult{}, false
	}
	req.Header.Set("User-Agent", e.userAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		atomic.AddUint64(&e.stats.errors, 1)
		// A cancelled context is not a transient error to retry.
		return fetchResult{}, !errors.Is(err, context.Canceled)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, e.cfg.MaxBodyBytes))
	if err != nil {
		return fetchResult{status: resp.StatusCode, contentType: resp.Header.Get("Content-Type")}, true
	}
	atomic.AddUint64(&e.stats.bytes, uint64(len(data)))

	transient = resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests
	return fetchResult{body: data, status: resp.StatusCode, contentType: resp.Header.Get("Content-Type")}, transient
}

// backoff returns the wait before attempt n: 200ms * 2^n, capped, with up to
// 50% random jitter to avoid thundering herds against the same origin.
func backoff(attempt int) time.Duration {
	base := 200 * time.Millisecond * time.Duration(math.Pow(2, float64(attempt)))
	if base > 5*time.Second {
		base = 5 * time.Second
	}
	jitter := time.Duration(rand.Int63n(int64(base)/2 + 1))
	return base + jitter
}
