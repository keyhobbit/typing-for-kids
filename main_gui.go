//go:build gui

package main

import (
	"fmt"
	"log"
	"time"

	webview "github.com/webview/webview_go"
)

func main() {
	port, err := freePort()
	if err != nil {
		log.Fatal("cannot find free port:", err)
	}

	startServer(port)
	// Give the HTTP server a moment to be ready before the webview tries to load
	time.Sleep(150 * time.Millisecond)

	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	w := webview.New(false) // false = no devtools
	defer w.Destroy()

	w.SetTitle("⌨️ KidTyping VN – Luyện Gõ Tiếng Việt")
	w.SetSize(1080, 780, webview.HintMin)
	w.Navigate(url)
	w.Run()
}
