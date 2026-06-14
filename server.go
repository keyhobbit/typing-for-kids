package main

import (
	"fmt"
	"net"
	"net/http"
)

// newServeMux builds the shared HTTP mux used by both the server and GUI binary.
func newServeMux() *http.ServeMux {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/next", handleNext)

	// Auth
	mux.HandleFunc("/api/auth/guest", handleAuthGuest)
	mux.HandleFunc("/api/auth/register", handleRegister)
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/auth/logout", handleLogout)

	// User management
	mux.HandleFunc("/api/user/name", handleRename)

	// Score & ranking
	mux.HandleFunc("/api/score", handleScore)
	mux.HandleFunc("/api/ranking", handleRanking)

	// Bắn Cung v2 (Math Quest) leaderboard
	mux.HandleFunc("/api/bow/score", handleBowScore)
	mux.HandleFunc("/api/bow/ranking", handleBowRanking)

	return mux
}

// freePort returns a random available TCP port on localhost.
func freePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port, nil
}

// startServer starts the HTTP server on the given port in a background goroutine.
func startServer(port int) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	go func() {
		if err := http.ListenAndServe(addr, newServeMux()); err != nil {
			// Non-fatal in GUI mode – window close will end the process anyway.
		}
	}()
}
