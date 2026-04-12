#!/usr/bin/env bash
# manage.sh – Start / stop / restart KidTyping VN
# Usage: ./manage.sh [start|stop|restart|status|logs]

set -euo pipefail

APP_NAME="kidtyping-vn"
APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="$APP_DIR/.app.pid"
LOG_FILE="$APP_DIR/.app.log"
BINARY="$APP_DIR/$APP_NAME"
GUI_BINARY="$APP_DIR/kidtyping-gui"
DESKTOP_FILE="$APP_DIR/kidtyping-vn.desktop"
PKG_CFG="$APP_DIR/pkgconfig"
PORT=11100

# ── Colours ────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; RESET='\033[0m'

info()    { echo -e "${CYAN}[info]${RESET}  $*"; }
success() { echo -e "${GREEN}[ok]${RESET}    $*"; }
warn()    { echo -e "${YELLOW}[warn]${RESET}  $*"; }
error()   { echo -e "${RED}[error]${RESET} $*"; }

# ── Helpers ────────────────────────────────────────────────────────────────
is_running() {
    [[ -f "$PID_FILE" ]] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null
}

build() {
    info "Building $APP_NAME (server) …"
    cd "$APP_DIR"
    if go build -o "$BINARY" .; then
        success "Build successful → $BINARY"
    else
        error "Build failed. Fix errors and try again."
        exit 1
    fi
}

build_gui() {
    info "Building kidtyping-gui (native window) …"
    cd "$APP_DIR"
    if CGO_ENABLED=1 PKG_CONFIG_PATH="$PKG_CFG" go build -tags gui -o "$GUI_BINARY" .; then
        success "GUI build successful → $GUI_BINARY"
    else
        error "GUI build failed. Ensure libwebkit2gtk-4.1-dev and build-essential are installed."
        exit 1
    fi
}

do_start() {
    if is_running; then
        warn "$APP_NAME is already running (PID $(cat "$PID_FILE"))."
        return
    fi

    [[ ! -f "$BINARY" ]] && build

    info "Starting $APP_NAME on port $PORT …"
    cd "$APP_DIR"
    nohup "$BINARY" >> "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"
    sleep 0.5

    if is_running; then
        success "$APP_NAME started (PID $(cat "$PID_FILE"))."
        success "Open → http://localhost:$PORT"
    else
        error "Process exited immediately. Check logs: ./manage.sh logs"
        rm -f "$PID_FILE"
        exit 1
    fi
}

do_stop() {
    if ! is_running; then
        warn "$APP_NAME is not running."
        return
    fi
    local pid
    pid=$(cat "$PID_FILE")
    info "Stopping $APP_NAME (PID $pid) …"
    kill "$pid"
    # Wait up to 5 s for graceful exit
    for _ in {1..10}; do
        kill -0 "$pid" 2>/dev/null || break
        sleep 0.5
    done
    if kill -0 "$pid" 2>/dev/null; then
        warn "Process still alive – sending SIGKILL …"
        kill -9 "$pid"
    fi
    rm -f "$PID_FILE"
    success "$APP_NAME stopped."
}

do_restart() {
    info "Restarting $APP_NAME …"
    do_stop
    build
    do_start
}

do_status() {
    echo -e "${BOLD}── KidTyping VN status ──────────────────${RESET}"
    if is_running; then
        local pid
        pid=$(cat "$PID_FILE")
        success "Running  (PID $pid)"
        echo -e "   URL  : http://localhost:$PORT"
        echo -e "   Log  : $LOG_FILE"
    else
        error "Not running"
    fi
}

do_logs() {
    if [[ ! -f "$LOG_FILE" ]]; then
        warn "No log file found at $LOG_FILE"
        return
    fi
    echo -e "${BOLD}── $LOG_FILE (last 50 lines) ──${RESET}"
    tail -n 50 "$LOG_FILE"
}

do_gui() {
    [[ ! -f "$GUI_BINARY" ]] && build_gui
    info "Launching KidTyping VN (native window) …"
    cd "$APP_DIR"
    exec "$GUI_BINARY"
}

install_desktop() {
    local dest="$HOME/.local/share/applications/kidtyping-vn.desktop"
    cp "$DESKTOP_FILE" "$dest"
    chmod +x "$dest"
    # Update-desktop-database if available (non-fatal)
    update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true
    success "Desktop shortcut installed → $dest"
    info  "You can now search 'KidTyping' in the Activities launcher."
}

# ── Dispatch ───────────────────────────────────────────────────────────────
case "${1:-help}" in
    start)         do_start       ;;
    stop)          do_stop        ;;
    restart)       do_restart     ;;
    status)        do_status      ;;
    logs)          do_logs        ;;
    build)         build          ;;
    gui)           do_gui         ;;
    build-gui)     build_gui      ;;
    install)       install_desktop ;;
    *)
        echo -e "${BOLD}KidTyping VN – App Manager${RESET}"
        echo ""
        echo "  Usage: ./manage.sh <command>"
        echo ""
        echo "  Server commands (headless, browser on port $PORT):"
        echo "    start      – Build (if needed) and start the server"
        echo "    stop       – Stop the running server"
        echo "    restart    – Stop, rebuild, and start"
        echo "    status     – Show running status and PID"
        echo "    logs       – Tail the last 50 lines of the log"
        echo "    build      – Compile the server binary only"
        echo ""
        echo "  GUI commands (native window, no browser/port needed):"
        echo "    gui        – Build (if needed) and launch native window"
        echo "    build-gui  – Compile the GUI binary only"
        echo "    install    – Install desktop shortcut (Activities launcher)"
        echo ""
        ;;
esac
