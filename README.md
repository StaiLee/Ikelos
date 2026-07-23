<div align="center">

<img src="./assets/banner.svg" alt="Ikelos — concurrent website mirroring engine" width="100%"/>

<br/>

**A fast, concurrent website mirroring engine with a live terminal dashboard.**
Point it at a site — it crawls, downloads every asset, rewrites the links, and leaves you a fully offline-browsable copy.

<br/>

![Build](https://img.shields.io/github/actions/workflow/status/StaiLee/Ikelos/ci.yml?branch=main&style=for-the-badge&logo=githubactions&logoColor=white&label=BUILD)
![Release](https://img.shields.io/github/v/release/StaiLee/Ikelos?style=for-the-badge&logo=github&color=39E0D8&label=RELEASE)
![Go](https://img.shields.io/github/go-mod/go-version/StaiLee/Ikelos?style=for-the-badge&logo=go&logoColor=white&color=00ADD8&label=GO)
![License](https://img.shields.io/github/license/StaiLee/Ikelos?style=for-the-badge&color=684A95&label=LICENSE)
![Stars](https://img.shields.io/github/stars/StaiLee/Ikelos?style=for-the-badge&logo=github&color=e3b341&label=STARS)

<br/>

![web-scraper](https://img.shields.io/badge/web--mirroring-12182b?style=flat-square)
![golang](https://img.shields.io/badge/Go-12182b?style=flat-square&logo=go)
![concurrency](https://img.shields.io/badge/concurrency-12182b?style=flat-square)
![TUI](https://img.shields.io/badge/TUI-12182b?style=flat-square)
![bubbletea](https://img.shields.io/badge/Bubble%20Tea-12182b?style=flat-square)
![goquery](https://img.shields.io/badge/goquery-12182b?style=flat-square)
![offline-archive](https://img.shields.io/badge/offline--archive-12182b?style=flat-square)

**[Features](#-features) · [Quick start](#-quick-start) · [Usage](#-usage) · [Architecture](#-architecture) · [Controls](#-runtime-controls)**

</div>

---

## 📖 About

**Ikelos** is a command-line tool that produces complete, offline-browsable mirrors of a website.
Built on a bounded concurrent Go architecture, it crawls a target, downloads every referenced asset
(images, stylesheets, scripts, fonts), rewrites all links to point at the local copies, and stores
everything on disk with a filesystem-safe layout. A live [Bubble Tea](https://github.com/charmbracelet/bubbletea)
dashboard shows progress, throughput and a real-time log while it works.

It is named after *Ikelos*, the Greek daemon of realistic dreams — the tool builds a faithful replica
of the original.

---

## ✦ Features

### ⚙️ Concurrency engine
- **Bounded worker pool** draining a lock-free, deadlock-free queue — `-threads` caps the actual number of live goroutines, so memory stays flat regardless of site size.
- **Exact completion detection** (a pending counter, not a racy `WaitGroup`) and **clean cancellation** via `context.Context`: pressing `Q` aborts in-flight requests and exits gracefully.
- **Smart retries** on transient failures only (network / 5xx / 429) with jittered exponential backoff; permanent 4xx responses are never retried.
- Persistent HTTP transport with connection pooling, per-response size cap, and per-job panic isolation.

### 🔬 HTML & CSS rewriting
- DOM rewriting with [goquery](https://github.com/PuerkitoBio/goquery): `src`, `href`, `poster`, lazy-load (`data-src`, `data-original`) and responsive `srcset` (descriptors preserved).
- Stylesheet surgery: both `url(...)` and `@import` references are localized, including inline `style="..."` attributes.

### 🗂️ Storage
- Deterministic URL → path mapping with per-segment sanitization (Windows-safe), query-string hashing, and `MAX_PATH` collapse via MD5 — no "file name too long" or illegal-character failures.
- **Path-traversal guard**: every write is verified to stay inside the output directory.
- Honest I/O accounting — a file only counts as saved when the write actually succeeds.

### 🕸️ Discovery & stealth
- **Sitemap hunter**: parses `sitemap.xml` (one level of sitemap-index nesting) to reach unlinked pages.
- **robots.txt matcher**: `User-agent` groups, `Allow`/`Disallow`, `*`/`$` wildcards, longest-match-wins — or bypass it entirely with `-ignore-robots`.
- HTTP/SOCKS **proxy** support and a rotating **user-agent** pool.

### 🖥️ Live dashboard
- Three-panel TUI: target/metrics, data-ingestion counters, and a scrolling live feed.
- Real-time throughput, memory, goroutine count, worker saturation and an activity sparkline.
- **Pause/resume** (`P`) at any time; three colour themes tied to the operational mode.

---

## ⚡ Quick start

**Requirements:** Go 1.24+ and a terminal with TrueColor support.

```bash
git clone https://github.com/StaiLee/Ikelos.git
cd Ikelos

# Build (compiles the whole module)
go build -ldflags="-s -w" -o ikelos .

# ...or run the full pipeline: vet + tests + build
make
```

Then run a mirror:

```bash
./ikelos -url https://example.com -out ./cloned_site
```

Running `ikelos` with no `-url` opens an interactive manual.

---

## 🚀 Usage

```bash
./ikelos -url <TARGET> [flags]
```

### Operational modes

| Mode | Flag | Behaviour | Best for |
|:--|:--|:--|:--|
| **Mirror** *(default)* | `-mode mirror` | Balanced recursion and speed, high fidelity | Offline backup, UI reference |
| **Shadow** | `-mode shadow` | Slow, randomized jitter, low profile | Rate-limited or protected sites |
| **Blitz** | `-mode blitz` | Max throughput, short timeouts | Docs / data hoarding, CTFs |

### Flags

| Flag | Default | Description |
|:--|:--|:--|
| `-url` | — | Target URL (scheme optional; defaults to `https://`). |
| `-out` | `./cloned_site` | Output directory for the mirror. |
| `-mode` | `mirror` | Operational profile: `mirror`, `shadow`, `blitz`. |
| `-depth` | `2` | Recursion depth (links to follow). |
| `-threads` | `20` | Number of concurrent workers. |
| `-proxy` | — | Proxy URL, e.g. `http://127.0.0.1:8080` or `socks5://127.0.0.1:9050`. |
| `-ua` | — | Custom User-Agent (overrides rotation). |
| `-nositemap` | `false` | Disable the sitemap hunter. |
| `-ignore-robots` | `false` | Ignore `robots.txt` and crawl everything. |
| `-insecure` | `true` | Skip TLS certificate verification (`-insecure=false` to enforce). |

### Examples

```bash
# High-fidelity archive
./ikelos -url https://example.com -mode mirror -out ./archives/example

# Stealth crawl through a local SOCKS proxy
./ikelos -url https://example.com -mode shadow -proxy socks5://127.0.0.1:9050

# Fast documentation dump, 50 workers, sitemap off
./ikelos -url https://docs.example.com -mode blitz -threads 50 -nositemap
```

---

## 🏗️ Architecture

Ikelos is split into small, independently-tested packages under `internal/`.

```
main.go              thin wiring: flags → Config → engine → TUI
internal/
├─ urlx/     URL resolution & scope checks
├─ robots/   robots.txt matcher (User-agent groups, Allow/Disallow, * $)
├─ store/    deterministic URL→disk mapping, writes, traversal guard
├─ rewrite/  HTML / CSS / srcset surgery (goquery)
├─ engine/   worker pool, dispatcher, fetcher, orchestration
└─ tui/      Bubble Tea dashboard + interactive manual
```

The crawl pipeline:

1. **Hunt** — fetch `robots.txt` and `sitemap.xml` *before* the pool starts, so seed pages are counted before any worker can drain the queue.
2. **Dispatch** — an unbounded mutex + `Cond` queue with an exact *pending* counter detects true completion and unblocks cleanly on cancellation.
3. **Pool** — a fixed number of workers pull jobs; deduplication happens at enqueue time via a `sync.Map`, keeping both concurrency and memory bounded.
4. **Rewrite** — DOM/CSS rewriting with all side effects behind a `Sink` interface, keeping the logic pure and unit-tested.
5. **Render** — the TUI reads a lock-free `Stats` snapshot and never blocks the engine (log lines are dropped, not queued, if the UI lags).

| Package | Responsibility |
|:--|:--|
| `internal/engine` | Concurrency core, HTTP fetching, retries, orchestration. |
| `internal/rewrite` | Point every asset/link reference at its local copy. |
| `internal/store` | Map URLs to safe on-disk paths and write files. |
| `internal/robots` | Parse and evaluate `robots.txt` rules. |
| `internal/urlx` | Absolute-URL resolution and same-host scoping. |
| `internal/tui` | Live dashboard and interactive manual. |

Run the test suite:

```bash
go test ./...
```

---

## 🎮 Runtime controls

| Key | Action |
|:--|:--|
| `P` | Pause / resume the engine |
| `Q` / `Ctrl+C` | Abort — cancels in-flight requests and exits cleanly |
| `↑` / `↓` | Scroll the live feed |

---

## ⚠️ Disclaimer

Ikelos is intended for **digital archiving, offline reading, and authorized testing**. Mirroring
sites you do not own may violate their terms of service or copyright. You are responsible for how
you use it; the author assumes no liability for misuse.

---

## 📜 License

Distributed under the **MIT License** — © 2026 Ilyes Staili (StaiLee). See [LICENSE](LICENSE).

<div align="center">
<br/>

**Ikelos** — *absorb. replicate. browse offline.*

</div>
