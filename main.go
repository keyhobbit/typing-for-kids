//go:build !gui

package main

import (
	"log"
	"net/http"
)

func main() {
	const port = 8700
	log.Printf("KidTyping VN – Listening on http://localhost:%d\n", port)
	if err := http.ListenAndServe(":8700", newServeMux()); err != nil {
		log.Fatal(err)
	}
}
