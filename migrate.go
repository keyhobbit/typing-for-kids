package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// migrateFromJSON imports data.json into SQLite if the users table is empty.
func migrateFromJSON() {
	const legacyFile = "data.json"

	// Only migrate if DB is empty AND legacy file exists
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count > 0 {
		return
	}
	b, err := os.ReadFile(legacyFile)
	if err != nil {
		return // no legacy file, fresh start
	}

	var legacy struct {
		Users map[string]*struct {
			ID        string    `json:"id"`
			DeviceID  string    `json:"device_id"`
			Username  string    `json:"username"`
			Name      string    `json:"name"`
			PassHash  string    `json:"pass_hash"`
			IsGuest   bool      `json:"is_guest"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"users"`
		Scores []struct {
			ID       string    `json:"id"`
			UserID   string    `json:"user_id"`
			Correct  int       `json:"correct"`
			Level    int       `json:"level"`
			ScoredAt time.Time `json:"scored_at"`
		} `json:"scores"`
	}
	if err := json.Unmarshal(b, &legacy); err != nil {
		log.Printf("migrate: cannot parse data.json: %v", err)
		return
	}

	tx, _ := db.Begin()
	for _, u := range legacy.Users {
		isGuest := 0
		if u.IsGuest {
			isGuest = 1
		}
		tx.Exec(
			`INSERT OR IGNORE INTO users(id,device_id,username,name,pass_hash,is_guest,created_at)
			 VALUES(?,?,?,?,?,?,?)`,
			u.ID, u.DeviceID, nullStr(u.Username), u.Name, u.PassHash, isGuest, u.CreatedAt.UTC(),
		)
	}
	for _, s := range legacy.Scores {
		if s.Correct <= 0 {
			continue
		}
		// Expand each legacy score record into individual 1-point rows
		for i := 0; i < s.Correct; i++ {
			tx.Exec(
				`INSERT OR IGNORE INTO scores(id,user_id,points,level,scored_at)
				 VALUES(?,?,1,?,?)`,
				s.ID+"-"+itoa(i), s.UserID, s.Level, s.ScoredAt.UTC(),
			)
		}
	}
	if err := tx.Commit(); err != nil {
		log.Printf("migrate commit: %v", err)
		return
	}
	log.Printf("migrate: imported %d users and %d score entries from data.json", len(legacy.Users), len(legacy.Scores))
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
func itoa(i int) string {
	return string(rune('0' + i%10)) // good enough for small i
}
