<p align="center">
  <img src="./assets/banner.svg" alt="Ikelos" width="100%"/>
</p>

<p align="center">
  <img src="https://img.shields.io/github/actions/workflow/status/StaiLee/Ikelos/ci.yml?branch=main&style=for-the-badge&logo=githubactions&logoColor=white&label=BUILD" alt="badge"/>
  <img src="https://img.shields.io/github/v/release/StaiLee/Ikelos?style=for-the-badge&logo=github&color=39E0D8&label=RELEASE" alt="badge"/>
  <img src="https://img.shields.io/github/go-mod/go-version/StaiLee/Ikelos?style=for-the-badge&logo=go&logoColor=white&color=00ADD8&label=GO" alt="badge"/>
  <img src="https://img.shields.io/github/license/StaiLee/Ikelos?style=for-the-badge&color=684A95&label=LICENSE" alt="badge"/>
  <img src="https://img.shields.io/github/stars/StaiLee/Ikelos?style=for-the-badge&logo=github&color=e3b341&label=STARS" alt="badge"/>
</p>

# 🧬 IKELOS - THE REALITY SHIFTER



> **The Omniscient Web Cloner.**
> *Absorb. Replicate. Dominate.*


---

## 📖 Overview

**IKELOS** is not just a web scraper; it is a **Tactical Reality Shifting Engine**.

Named after the Greek god of realistic dreams, Ikelos creates perfect, offline-browsable mirrors of any target website. Built on a high-performance concurrent Golang architecture, it bypasses modern protections (Lazy Loading, WAFs) to deliver pixel-perfect replicas.

**Version 8.0 "THE SURGEON"** is a full engineering rebuild: a bounded, deadlock-free worker pool, clean context cancellation, honest error accounting, and a modular package architecture — with the same "God Tier" TUI and stealth toolkit on top.

### ✨ Features (v8.0.0)

* **🩺 Bounded Concurrency Core (REBUILT):** A fixed worker pool draining an unbounded, deadlock-free queue. The `-threads` value now *actually* caps live goroutines — no more RAM blow-ups on large targets — while completion is detected exactly (no `WaitGroup` races).
* **🛑 Clean Cancellation (NEW):** `Q` cancels the crawl for real via `context.Context`; in-flight requests are interrupted and workers exit gracefully.
* **🛡️ Stealth & Proxy Support:** Native HTTP/SOCKS proxies and a user-agent rotation pool to evade IP bans and WAFs.
* **⏸️ Tactical Pause:** Press `P` to freeze the engine, analyze real-time logs, and resume seamlessly.
* **🧠 Intelligent Asset Parsing:**
    * **Srcset Decoding:** Downloads responsive `srcset` images while preserving descriptors.
    * **Deep CSS Surgery:** Rewrites `url(...)` **and** `@import` references (both the `url()` and bare-string forms) so stylesheets resolve fully offline.
* **🤖 Proper robots.txt:** A real matcher — `User-agent` grouping, `Allow`/`Disallow`, `*`/`$` wildcards, longest-match-wins. Override with `-ignore-robots`.
* **💾 MD5 Hashed Storage:** Deterministic URL→path mapping with per-segment sanitization, query hashing, `MAX_PATH` collapse, and a **path-traversal guard** that keeps every write inside the output directory.
* **🕷️ Sitemap Hunter:** Parses `sitemap.xml` (following one level of sitemap-index nesting) to discover unlinked pages.
* **👁️ Lazy-Load Killer:** Forces download of hidden assets (`data-src`, `data-original`).
* **🎨 Dynamic Themes:**
    * **🔵 MIRROR:** Cyberpunk Cyan/Blue (High Fidelity).
    * **🔴 BLITZ:** Aggressive Magma Red (Max Speed).
    * **🟣 SHADOW:** Void Purple (Stealth/Low Profile).

---

## ⚡ Installation

### Prerequisites
* **Go 1.21** or higher installed.
* A terminal with **TrueColor** support.

### Fast Install

```bash
# 1. Clone the repository
git clone [https://github.com/StaiLee/Ikelos.git]
cd Ikelos

# 2. Install dependencies
go mod tidy

# 3. Build the binary (compiles the whole module, not just main.go)
go build -ldflags="-s -w" -o ikelos .

# ...or just run the full pipeline (vet + tests + build)
make
```

---

## 🚀 Usage

Ikelos simplifies complex mirroring tasks into tactical modes, now with advanced routing options.

```bash
./ikelos -url <TARGET> [FLAGS]
```

### 🛡️ Tactical Modes

| Mode | Code | Description | Best Use Case |
| :--- | :--- | :--- | :--- |
| **MIRROR** | `-mode mirror` | **(Default)** High fidelity. Balanced recursion and speed. | UI/UX Theft, Offline Backup. |
| **SHADOW** | `-mode shadow` | **Stealth / Evasion.** Slow, randomized jitter, human emulation. | Protected Sites (Cloudflare), WAFs. |
| **BLITZ** | `-mode blitz` | **Aggressive Dump.** Max threads, minimal timeouts. | Data Hoarding, CTFs, Docs scraping. |

### 🚩 Command Flags

| Flag | Description |
| :--- | :--- |
| `-url` | The target website (e.g., `https://example.com`). |
| `-out` | Output directory for the clone (Default: `./cloned_site`). |
| `-proxy` | **(NEW)** Proxy URL (e.g., `http://127.0.0.1:8080` or `socks5://...`). |
| `-ua` | **(NEW)** Custom User-Agent string (Overrides rotation). |
| `-depth` | Recursion depth. `2` is standard. |
| `-threads` | Number of concurrent workers (Default: `20`). |
| `-nositemap` | Disable the *Sitemap Hunter* module (Strict crawling). |
| `-ignore-robots` | **(NEW)** Ignore `robots.txt` entirely and crawl everything. |
| `-insecure` | Skip TLS certificate verification (Default: `true`). Set `-insecure=false` to enforce valid certs. |

### 🎮 Runtime Controls

* **`P`**: Toggle **Pause/Resume**.
* **`Q`** or **`Ctrl+C`**: Abort mission and save current state.

### 💡 Operational Examples

**1. The "Ghost" Extraction (Proxy + Stealth)**
Clone a protected site through a local Tor proxy with randomized delays.
```bash
./ikelos -url [https://protected-target.com](https://protected-target.com) -mode shadow -proxy socks5://127.0.0.1:9050
```

**2. The Archive (High Fidelity)**
Create a perfect mirror, respecting `robots.txt` but hunting for hidden sitemap links.
```bash
./ikelos -url [https://awwwards.com](https://awwwards.com) -mode mirror -out ./archives/awwwards
```

**3. The Smash & Grab (Max Speed)**
Download a documentation site using 50 threads, ignoring sitemaps for speed.
```bash
./ikelos -url [https://docs.python.org](https://docs.python.org) -mode blitz -threads 50 -nositemap
```

---

## 🏗️ Technical Architecture

Ikelos v8.0 is split into small, independently-testable packages under `internal/`:

```
main.go                 # thin wiring: flags → Config → engine → TUI
internal/
├─ urlx/     # URL resolution & scope checks (pure, tested)
├─ robots/   # robots.txt matcher: User-agent groups, Allow/Disallow, *,$ (tested)
├─ store/    # deterministic URL→disk mapping + writes + traversal guard (tested)
├─ rewrite/  # HTML / CSS / srcset surgery via goquery (pure, tested)
├─ engine/   # bounded worker pool, dispatcher, fetcher, orchestration (tested + integration)
└─ tui/      # Bubble Tea dashboard + interactive manual
```

The crawl pipeline:

1.  **The Hunter** (`engine`): fetches `robots.txt` and `sitemap.xml` *synchronously* before the pool starts, so seeds are counted before any worker can drain the queue — no `WaitGroup` race.
2.  **The Dispatcher** (`engine/dispatcher.go`): an unbounded, mutex+`Cond` queue with an exact *pending* counter. It detects true completion (nothing queued or in flight) and unblocks cleanly on cancellation.
3.  **The Pool** (`engine`): a **fixed** number of workers (`-threads`) pull jobs; deduplication happens at enqueue time via a `sync.Map`, so concurrency and memory are both bounded regardless of site size.
4.  **The Surgeon** (`rewrite`): `goquery`-based DOM rewriting, `srcset` decoding, and CSS `url()`/`@import` localization — with all outside effects behind a `Sink` interface, keeping the logic pure and unit-tested.
5.  **The Overseer** (`tui`): a Bubble Tea loop rendering a live dashboard from a lock-free `Stats` snapshot, never blocking the engine (logs are dropped, not queued, if the UI lags).

Reliability guarantees: context-based cancellation, retry **only** on transient failures (network / 5xx / 429) with jittered backoff, `io.LimitReader`-capped response bodies, real I/O error accounting, and panic isolation per job.

---

## ⚠️ Disclaimer

**Ikelos is intended for educational purposes, digital archiving, and authorized security testing.**
Cloning websites you do not own may violate terms of service or copyright laws. The developers assume no liability for misuse.

---

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.
