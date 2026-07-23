package engine

import (
	"sync"
	"sync/atomic"
	"testing"
)

// A fan-out crawl simulation: each of the first N jobs enqueues two children,
// exactly once. The dispatcher must process every job and then terminate
// cleanly with all workers exiting.
func TestDispatcherFanOutTerminates(t *testing.T) {
	d := newDispatcher()
	var processed int64
	var seen sync.Map

	enqueue := func(id string, depth int) {
		if _, loaded := seen.LoadOrStore(id, true); loaded {
			return
		}
		d.submit(job{url: id, depth: depth})
	}

	const workers = 8
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				j, ok := d.pop()
				if !ok {
					return
				}
				atomic.AddInt64(&processed, 1)
				if j.depth < 3 {
					enqueue(j.url+"L", j.depth+1)
					enqueue(j.url+"R", j.depth+1)
				}
				d.complete()
			}
		}()
	}

	enqueue("root", 0)
	wg.Wait()

	// depth 0..3 full binary tree = 1+2+4+8 = 15 nodes.
	if got := atomic.LoadInt64(&processed); got != 15 {
		t.Errorf("processed %d jobs, want 15", got)
	}
	if d.outstanding() != 0 {
		t.Errorf("outstanding = %d, want 0", d.outstanding())
	}
}

func TestDispatcherShutdownUnblocks(t *testing.T) {
	d := newDispatcher()
	done := make(chan struct{})
	go func() {
		_, ok := d.pop() // blocks: queue empty, not done
		if ok {
			t.Error("pop returned ok after shutdown")
		}
		close(done)
	}()
	d.shutdown()
	<-done // must not hang
}
