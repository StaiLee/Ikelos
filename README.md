# 🧬 IKELOS - THE REALITY SHIFTER

> **The Omniscient Web Cloner.**
> *Absorb. Replicate. Dominate.*

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)

---

## 📖 Overview

**IKELOS** is not just a web scraper; it is a **Tactical Reality Shifting Engine**.

Named after the Greek god of realistic dreams, Ikelos creates perfect, offline-browsable mirrors of any target website. Built on a high-performance concurrent Golang architecture, it bypasses modern protections (Lazy Loading, WAFs) to deliver pixel-perfect replicas.

**Version 7.0 "THE ARCHITECT"** introduces a redefined "God Tier" TUI, advanced stealth capabilities with proxy support, and intelligent asset parsing for the modern web.

### ✨ OMNISCIENT Features (v7.0.0)

* **🛡️ Stealth & Proxy Support (NEW):** Native support for HTTP/SOCKS proxies and a massive user-agent rotation pool to evade IP bans and WAFs.
* **⏸️ Tactical Pause (NEW):** Press `P` at any time to freeze the engine, analyze real-time logs, and resume operations seamlessly.
* **🧠 Intelligent Asset Parsing (NEW):**
    * **Srcset Decoding:** Downloads high-resolution images defined in responsive `srcset` attributes.
    * **Deep CSS Analysis:** Detects and downloads assets hidden within `@import` rules and complex `url()` definitions.
* **💾 MD5 Hashed Storage:** Smart file management uses MD5 hashing for complex URLs and query strings, eliminating "File name too long" errors and filesystem corruption.
* **🕷️ Sitemap Hunter:** Automatically detects and parses `sitemap.xml` and `robots.txt` to discover hidden pages not linked on the homepage.
* **⚡ The Swarm Engine:** A massive concurrent downloader capable of saturating bandwidth with hundreds of micro-threads.
* **👁️ Lazy-Load Killer:** Detects and forces the download of hidden assets (`data-src`, `data-original`), ensuring no broken images or "grey squares".
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

# 3. Build the binary
go build -ldflags="-s -w" -o ikelos main.go
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

Ikelos v7.0 is an engineering lesson in **Go Concurrency** and **HTML Surgery**:

1.  **The Hunter:** Scans `robots.txt` and `sitemap.xml` to build an initial target map.
2.  **The Brain (Crawler):** Manages a thread-safe `Visited` map, handles pause states, and distributes jobs via a semaphore-controlled queue.
3.  **The Swarm (Workers):** Hundreds of Goroutines fetch assets simultaneously via a persistent HTTP Transport.
4.  **The Surgeon (Parser):** Uses `goquery` to parse DOM, inject local MD5-hashed paths, resolve CSS `@import`/`url()`, and decrypt `srcset` attributes.
5.  **The Overseer (TUI):** A separate Bubble Tea event loop renders the dashboard at 60fps, providing a real-time "Live Feed" and segmented progress visualization without blocking the engine.

---

## ⚠️ Disclaimer

**Ikelos is intended for educational purposes, digital archiving, and authorized security testing.**
Cloning websites you do not own may violate terms of service or copyright laws. The developers assume no liability for misuse.

---

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.