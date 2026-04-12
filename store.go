package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
)

// ── Data types ─────────────────────────────────────────────────────────────

// User represents a player (guest or registered).
type User struct {
	ID       string
	DeviceID string
	Username string
	Name     string
	PassHash string
	IsGuest  bool
}

// RankEntry is one row in the leaderboard.
type RankEntry struct {
	Rank    int    `json:"rank"`
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
	Score   int    `json:"score"`
	IsGuest bool   `json:"is_guest"`
}

// ── User operations ──────────────────────────────────────────────────────────

// GetOrCreateGuest returns the existing user with deviceID, or creates a guest.
func GetOrCreateGuest(deviceID string) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, device_id, COALESCE(username,''), name, pass_hash, is_guest
		 FROM users WHERE device_id = ?`, deviceID,
	).Scan(&u.ID, &u.DeviceID, &u.Username, &u.Name, &u.PassHash, &u.IsGuest)
	if err == nil {
		return u, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	id := newUUID()
	_, err = db.Exec(
		`INSERT INTO users(id,device_id,name,is_guest,created_at)
		 VALUES(?,?,?,1,datetime('now'))`,
		id, deviceID, id[:10],
	)
	if err != nil {
		return nil, err
	}
	return &User{ID: id, DeviceID: deviceID, Name: id[:10], IsGuest: true}, nil
}

// RegisterUser upgrades or creates a registered account.
func RegisterUser(deviceID, username, password string) (*User, error) {
	uname := strings.TrimSpace(username)
	if uname == "" || strings.ContainsAny(uname, " \t\n\r") {
		return nil, fmt.Errorf("tên đăng nhập không hợp lệ (không có khoảng trắng)")
	}
	if len([]rune(uname)) > 20 {
		return nil, fmt.Errorf("tên đăng nhập tối đa 20 ký tự")
	}
	if len(password) < 4 {
		return nil, fmt.Errorf("mật khẩu phải có ít nhất 4 ký tự")
	}

	var exists int
	db.QueryRow(`SELECT COUNT(*) FROM users WHERE LOWER(username)=LOWER(?)`, uname).Scan(&exists)
	if exists > 0 {
		return nil, fmt.Errorf("tên đăng nhập đã được sử dụng")
	}

	hash, err := hashPass(password)
	if err != nil {
		return nil, fmt.Errorf("lỗi hệ thống")
	}

	// Try to upgrade existing guest for this device, otherwise create fresh
	var userID string
	db.QueryRow(`SELECT id FROM users WHERE device_id=? AND is_guest=1`, deviceID).Scan(&userID)

	if userID != "" {
		_, err = db.Exec(
			`UPDATE users SET username=?, name=?, pass_hash=?, is_guest=0 WHERE id=?`,
			uname, uname, hash, userID,
		)
	} else {
		userID = newUUID()
		_, err = db.Exec(
			`INSERT INTO users(id,device_id,username,name,pass_hash,is_guest,created_at)
			 VALUES(?,?,?,?,?,0,datetime('now'))`,
			userID, deviceID, uname, uname, hash,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("lỗi hệ thống")
	}
	return &User{ID: userID, DeviceID: deviceID, Username: uname, Name: uname, IsGuest: false}, nil
}

// LoginUser authenticates by username + password.
func LoginUser(username, password string) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, device_id, COALESCE(username,''), name, pass_hash, is_guest
		 FROM users WHERE LOWER(username)=LOWER(?)`, username,
	).Scan(&u.ID, &u.DeviceID, &u.Username, &u.Name, &u.PassHash, &u.IsGuest)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("không tìm thấy tài khoản")
	}
	if err != nil {
		return nil, fmt.Errorf("lỗi hệ thống")
	}
	if !checkPass(password, u.PassHash) {
		return nil, fmt.Errorf("sai mật khẩu")
	}
	return u, nil
}

// RenameUser changes the display name.
func RenameUser(userID, newName string) error {
	name := strings.TrimSpace(newName)
	if name == "" {
		return fmt.Errorf("tên không được để trống")
	}
	if len([]rune(name)) > 20 {
		return fmt.Errorf("tên tối đa 20 ký tự")
	}
	res, err := db.Exec(`UPDATE users SET name=? WHERE id=?`, name, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("không tìm thấy người dùng")
	}
	return nil
}

// GetUserByID returns the user with the given ID.
func GetUserByID(id string) (*User, bool) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, device_id, COALESCE(username,''), name, pass_hash, is_guest
		 FROM users WHERE id=?`, id,
	).Scan(&u.ID, &u.DeviceID, &u.Username, &u.Name, &u.PassHash, &u.IsGuest)
	if err != nil {
		return nil, false
	}
	return u, true
}

// ── Score operations ───────────────────────────────────────────────────────

// AddScore records one correct word for the user (1 point per call).
func AddScore(userID string, level int) error {
	_, err := db.Exec(
		`INSERT INTO scores(id,user_id,points,level,scored_at)
		 VALUES(?,?,1,?,datetime('now'))`,
		newUUID(), userID, level,
	)
	return err
}

// ── Ranking (from cache) ───────────────────────────────────────────────────

// GetRanking returns the latest cached leaderboard for the given period.
// period: "day" | "week" | "month" | "year"
func GetRanking(period string) []RankEntry {
	switch period {
	case "day", "week", "month", "year":
	default:
		period = "day"
	}
	rows, err := db.Query(
		`SELECT rank, user_id, name, score, is_guest
		 FROM ranking_cache WHERE period=? ORDER BY rank ASC`,
		period,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []RankEntry
	for rows.Next() {
		var e RankEntry
		rows.Scan(&e.Rank, &e.UserID, &e.Name, &e.Score, &e.IsGuest)
		e.IsGuest = e.IsGuest // already bool-ified by scanner
		out = append(out, e)
	}
	return out
}

// ── Crypto helpers ─────────────────────────────────────────────────────────

func hashPass(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	saltHex := hex.EncodeToString(salt)
	h := sha256.Sum256([]byte(saltHex + password))
	return saltHex + ":" + hex.EncodeToString(h[:]), nil
}

func checkPass(password, stored string) bool {
	parts := strings.SplitN(stored, ":", 2)
	if len(parts) != 2 {
		return false
	}
	h := sha256.Sum256([]byte(parts[0] + password))
	return hex.EncodeToString(h[:]) == parts[1]
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

