//go:build !gui

package main

import (
	"log"
	"net/http"
)

func main() {
	const port = 11100
	initDB()
	StartRanker()
	log.Printf("KidTyping VN – Listening on http://localhost:%d\n", port)
	if err := http.ListenAndServe(":11100", newServeMux()); err != nil {
		log.Fatal(err)
	}
}
