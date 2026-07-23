package engine

import "sync"

type jobKind int

const (
	jobCrawl jobKind = iota
	jobAsset
)

type job struct {
	url     string
	referer string
	depth   int
	kind    jobKind
}

// dispatcher is an unbounded work queue drained by a fixed pool of workers. It
// tracks the number of outstanding jobs ("pending") so it can detect true
// completion — the moment no job is queued or in flight — without a WaitGroup
// race, and without spawning one goroutine per URL. New work may be submitted
// from inside a job (crawling a page enqueues its assets and links); as long as
// that happens before the parent job completes, pending never hits zero early.
type dispatcher struct {
	mu      sync.Mutex
	cond    *sync.Cond
	queue   []job
	head    int
	pending int
	done    bool
}

func newDispatcher() *dispatcher {
	d := &dispatcher{}
	d.cond = sync.NewCond(&d.mu)
	return d
}

// submit adds a job. It is a no-op once the dispatcher is done (drained or shut
// down), so late submissions after cancellation are silently dropped.
func (d *dispatcher) submit(j job) {
	d.mu.Lock()
	if d.done {
		d.mu.Unlock()
		return
	}
	d.queue = append(d.queue, j)
	d.pending++
	d.mu.Unlock()
	d.cond.Signal()
}

// pop blocks until a job is available, returning ok=false when all work is done
// or the dispatcher was shut down.
func (d *dispatcher) pop() (job, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for d.head >= len(d.queue) && !d.done {
		d.cond.Wait()
	}
	if d.head >= len(d.queue) {
		return job{}, false
	}
	j := d.queue[d.head]
	d.queue[d.head] = job{} // drop reference so the string can be GC'd
	d.head++
	// Periodically compact the backing array so a long crawl doesn't retain
	// every job ever queued.
	if d.head > 1024 && d.head*2 >= len(d.queue) {
		n := copy(d.queue, d.queue[d.head:])
		d.queue = d.queue[:n]
		d.head = 0
	}
	return j, true
}

// complete marks one popped job as fully processed. When the last outstanding
// job completes, the dispatcher transitions to done and wakes every worker.
func (d *dispatcher) complete() {
	d.mu.Lock()
	d.pending--
	finished := d.pending == 0 && !d.done
	if finished {
		d.done = true
	}
	d.mu.Unlock()
	if finished {
		d.cond.Broadcast()
	}
}

// shutdown forces completion regardless of pending work (used on cancellation).
func (d *dispatcher) shutdown() {
	d.mu.Lock()
	d.done = true
	d.mu.Unlock()
	d.cond.Broadcast()
}

// outstanding returns the number of jobs queued or in flight.
func (d *dispatcher) outstanding() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.pending
}
