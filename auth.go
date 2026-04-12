package main

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

var (
	tokenMu sync.RWMutex
	tokens  = make(map[string]string) // Bearer token → userID
)

// NewToken creates a random token and maps it to the given userID.
func NewToken(userID string) string {
	b := make([]byte, 20)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	tokenMu.Lock()
	tokens[token] = userID
	tokenMu.Unlock()
	return token
}

// ResolveToken returns the userID for a valid token.
func ResolveToken(token string) (string, bool) {
	tokenMu.RLock()
	id, ok := tokens[token]
	tokenMu.RUnlock()
	return id, ok
}

// RevokeToken deletes a token (logout).
func RevokeToken(token string) {
	tokenMu.Lock()
	delete(tokens, token)
	tokenMu.Unlock()
}
