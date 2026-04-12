package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
)

// NewToken creates a random token, persists it to the sessions table, and returns it.
func NewToken(userID string) string {
	b := make([]byte, 20)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	db.Exec(
		`INSERT OR IGNORE INTO sessions(token,user_id,created_at) VALUES(?,?,datetime('now'))`,
		token, userID,
	)
	return token
}

// ResolveToken returns the userID for a valid token, or ("", false) if unknown.
func ResolveToken(token string) (string, bool) {
	var userID string
	err := db.QueryRow(`SELECT user_id FROM sessions WHERE token=?`, token).Scan(&userID)
	if err == sql.ErrNoRows || err != nil {
		return "", false
	}
	return userID, true
}

// RevokeToken removes a session (logout).
func RevokeToken(token string) {
	db.Exec(`DELETE FROM sessions WHERE token=?`, token)
}
