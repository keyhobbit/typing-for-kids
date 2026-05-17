package main

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func dbFile() string {
	if p := os.Getenv("DB_PATH"); p != "" {
		return p
	}
	return "kidtyping.db"
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", dbFile()+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1) // SQLite: single writer
	if err := createSchema(); err != nil {
		log.Fatalf("create schema: %v", err)
	}
	migrateFromJSON()
}

func createSchema() error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id         TEXT PRIMARY KEY,
		device_id  TEXT UNIQUE,
		username   TEXT UNIQUE,
		name       TEXT NOT NULL,
		pass_hash  TEXT NOT NULL DEFAULT '',
		is_guest   INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token      TEXT PRIMARY KEY,
		user_id    TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS scores (
		id         TEXT PRIMARY KEY,
		user_id    TEXT NOT NULL,
		points     INTEGER NOT NULL DEFAULT 1,
		level      INTEGER NOT NULL DEFAULT 1,
		scored_at  DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_scores_user_at ON scores(user_id, scored_at);
	CREATE INDEX IF NOT EXISTS idx_sessions_user  ON sessions(user_id);

	CREATE TABLE IF NOT EXISTS ranking_cache (
		period     TEXT NOT NULL,
		rank       INTEGER NOT NULL,
		user_id    TEXT NOT NULL,
		name       TEXT NOT NULL,
		score      INTEGER NOT NULL DEFAULT 0,
		is_guest   INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL,
		PRIMARY KEY(period, user_id)
	);
	`)
	return err
}
