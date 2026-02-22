# ⌨️ KidTyping VN

A lightweight, high-performance typing-practice app for Vietnamese primary-school children.
Runs as either a **native desktop window** (no browser needed) or a classic **web server**.

---

## Features

- **3 difficulty levels** with 1,497 hand-curated Vietnamese lessons:

  | Level | Content | Entries |
  |-------|---------|---------|
  | Cấp 1 | Single words – covers all 6 tones, all vowel groups (a/ă/â/e/ê/i/o/ô/ơ/u/ư/y), all final consonants | 964 |
  | Cấp 2 | Common phrases (2–4 words) | 379 |
  | Cấp 3 | Full sentences, proverbs, and short narratives | 154 |

- **✍️ Free-typing mode** – no target, just a blank canvas for open practice
- **Real-time character-match progress bar** as you type
- **Enter to check** or click the check button; green/red border + "Bravô!" / "Thử lại nhé!" feedback
- **Score counter** (correct / total / current level)
- **No client-side Telex engine** – Vietnamese input is handled by the system IME (**Unikey**, **iBus-Bamboo**, etc.) with zero conflict
- **Native GUI** via WebKit2GTK – same UI as the web version, zero browser dependency
- **Desktop shortcut** for the GNOME Activities launcher

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22+ · `net/http` · `html/template` |
| Frontend | HTML5 · CSS3 (Tailwind-inspired custom CSS) · Vanilla JavaScript |
| Native GUI | [`webview_go`](https://github.com/webview/webview_go) (WebKit2GTK) |
| Port (server mode) | 8700 (server) / random localhost port (GUI mode) |

---

## Project Structure

```
/typing-for-kids
├── main.go              # Server entry point  (build tag: !gui)
├── main_gui.go          # GUI entry point     (build tag:  gui)
├── server.go            # Shared HTTP mux, freePort(), startServer()
├── handlers.go          # HTTP handlers: page render + /api/next JSON
├── dictionary.go        # Vietnamese lesson bank + GetRandomLesson()
├── manage.sh            # CLI tool: start / stop / restart / gui / install
├── kidtyping-vn.desktop # GNOME desktop entry
├── pkgconfig/
│   └── webkit2gtk-4.0.pc  # pkg-config shim (4.0 → 4.1)
├── static/
│   ├── css/style.css    # Kid-friendly UI styling
│   └── js/telex.js      # (unused – kept for reference)
└── templates/
    └── index.html       # Main application layout
```

---

## Prerequisites

### Server mode only
- Go 1.22+

### GUI mode (native window)
```bash
sudo apt install build-essential libwebkit2gtk-4.1-dev libgtk-3-dev pkg-config
```

---

## Quick Start

```bash
# Clone / enter the project
cd typing-for-kids

# ── Option A: Native desktop window (recommended) ──
./manage.sh gui          # builds & opens a native window automatically

# ── Option B: Web server (open http://localhost:8700 in any browser) ──
./manage.sh start
```

### Install desktop shortcut (GNOME)
```bash
./manage.sh install      # adds KidTyping to the Activities launcher
```

---

## manage.sh Reference

```
Server commands (headless – browser on port 8700):
  start      – Build if needed, then start server in background
  stop       – Gracefully stop the running server
  restart    – Stop → rebuild → start
  status     – Show PID, URL, and log path
  logs       – Tail the last 50 lines of .app.log
  build      – Compile the server binary only

GUI commands (native window – no browser / fixed port):
  gui        – Build GUI binary if needed, then launch native window
  build-gui  – Compile kidtyping-gui binary only
  install    – Install .desktop shortcut → ~/.local/share/applications
```

---

## How the GUI Works

`main_gui.go` picks a **random free TCP port**, starts the HTTP server bound to
`127.0.0.1:<port>` (never exposed externally), waits 150 ms, then opens a
**WebKit2GTK native window** pointed at that URL.  
The UI is byte-for-byte identical to the browser version.

---

## Vietnamese Input

The app has **no client-side input method engine**. All Vietnamese diacritics are
produced by the OS-level IME:

| IME | Install |
|-----|---------|
| **Unikey** | `sudo apt install unikey` or from https://www.unikey.org |
| **iBus-Bamboo** | `sudo apt install ibus-bamboo` |
| **iBus-Unikey** | `sudo apt install ibus-unikey` |

Set the input method to **Telex** or **VNI** in your IME settings before practicing.

---

## API

| Method | Path | Query | Response |
|--------|------|-------|----------|
| `GET` | `/` | `?level=1\|2\|3` | HTML page with initial lesson |
| `GET` | `/api/next` | `?level=1\|2\|3` | `{"content":"…","level":N}` |

---

## License

MIT – free to use, modify, and distribute.
