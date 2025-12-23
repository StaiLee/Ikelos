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