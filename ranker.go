package main

import (
	"log"
	"time"
)

// StartRanker runs an initial rebuild immediately, then rebuilds the
// ranking_cache table every 5 minutes in the background.
func StartRanker() {
	go func() {
		RebuildRankingCache()
		for range time.NewTicker(5 * time.Minute).C {
			RebuildRankingCache()
		}
	}()
}

// RebuildRankingCache recalculates all four leaderboards (day/week/month/year)
// and atomically replaces the rows in ranking_cache.
func RebuildRankingCache() {
	now := time.Now().UTC()
	cutoffs := map[string]time.Time{
		"day":   time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		"week":  now.AddDate(0, 0, -7),
		"month": now.AddDate(0, -1, 0),
		"year":  now.AddDate(-1, 0, 0),
	}

	for period, cutoff := range cutoffs {
		if err := rebuildPeriod(period, cutoff); err != nil {
			log.Printf("ranker: rebuild %s: %v", period, err)
		}
		if err := rebuildBowPeriod(period, cutoff); err != nil {
			log.Printf("ranker: rebuild bow-%s: %v", period, err)
		}
	}
}

// rebuildBowPeriod materialises the Bắn Cung v2 leaderboard: each player's
// BEST "lực chiến" (MAX power) within the window, cached under "bow-<period>".
func rebuildBowPeriod(period string, cutoff time.Time) error {
	rows, err := db.Query(`
		SELECT u.id, u.name, u.is_guest, COALESCE(MAX(b.power),0) AS score
		FROM users u
		JOIN bow_scores b ON b.user_id = u.id
		WHERE b.scored_at >= ?
		GROUP BY u.id
		ORDER BY score DESC
		LIMIT 100
	`, cutoff.Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}
	defer rows.Close()

	type row struct {
		userID  string
		name    string
		isGuest int
		score   int
	}
	var entries []row
	for rows.Next() {
		var e row
		rows.Scan(&e.userID, &e.name, &e.isGuest, &e.score)
		entries = append(entries, e)
	}
	rows.Close()

	cacheKey := "bow-" + period
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	tx.Exec(`DELETE FROM ranking_cache WHERE period=?`, cacheKey)
	updatedAt := time.Now().UTC().Format("2006-01-02T15:04:05")
	for rank, e := range entries {
		tx.Exec(
			`INSERT INTO ranking_cache(period,rank,user_id,name,score,is_guest,updated_at)
			 VALUES(?,?,?,?,?,?,?)`,
			cacheKey, rank+1, e.userID, e.name, e.score, e.isGuest, updatedAt,
		)
	}
	return tx.Commit()
}

func rebuildPeriod(period string, cutoff time.Time) error {
	rows, err := db.Query(`
		SELECT u.id, u.name, u.is_guest, COALESCE(SUM(s.points),0) AS score
		FROM users u
		JOIN scores s ON s.user_id = u.id
		WHERE s.scored_at >= ?
		GROUP BY u.id
		ORDER BY score DESC
		LIMIT 100
	`, cutoff.Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}
	defer rows.Close()

	type row struct {
		userID  string
		name    string
		isGuest int
		score   int
	}
	var entries []row
	for rows.Next() {
		var e row
		rows.Scan(&e.userID, &e.name, &e.isGuest, &e.score)
		entries = append(entries, e)
	}
	rows.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	tx.Exec(`DELETE FROM ranking_cache WHERE period=?`, period)
	updatedAt := time.Now().UTC().Format("2006-01-02T15:04:05")
	for rank, e := range entries {
		tx.Exec(
			`INSERT INTO ranking_cache(period,rank,user_id,name,score,is_guest,updated_at)
			 VALUES(?,?,?,?,?,?,?)`,
			period, rank+1, e.userID, e.name, e.score, e.isGuest, updatedAt,
		)
	}
	return tx.Commit()
}
