package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// ── Data types ─────────────────────────────────────────────────────────────

// User represents a player (guest or registered).
type User struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"device_id"`
	Username  string    `json:"username,omitempty"`
	Name      string    `json:"name"`
	PassHash  string    `json:"pass_hash,omitempty"`
	IsGuest   bool      `json:"is_guest"`
	CreatedAt time.Time `json:"created_at"`
}

// Score records one submission (1 correct answer).
type Score struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Correct  int       `json:"correct"`
	Total    int       `json:"total"`
	Level    int       `json:"level"`
	ScoredAt time.Time `json:"scored_at"`
}

// RankEntry is one row in the leaderboard.
type RankEntry struct {
	Rank    int    `json:"rank"`
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
	Score   int    `json:"score"`
	IsGuest bool   `json:"is_guest"`
}

// ── Persistent store ────────────────────────────────────────────────────────

type appData struct {
	Users  map[string]*User `json:"users"`
	Scores []Score          `json:"scores"`
}

var (
	dataMu sync.RWMutex
	appDB  = &appData{
		Users:  make(map[string]*User),
		Scores: []Score{},
	}
)

const storeFile = "data.json"

func init() {
	b, err := os.ReadFile(storeFile)
	if err != nil {
		if !os.IsNotExist(err) {
			panic("read store: " + err.Error())
		}
		return // fresh start
	}
	if err := json.Unmarshal(b, appDB); err != nil {
		panic("corrupt store: " + err.Error())
	}
	if appDB.Users == nil {
		appDB.Users = make(map[string]*User)
	}
	if appDB.Scores == nil {
		appDB.Scores = []Score{}
	}
}

func persist() {
	b, _ := json.Marshal(appDB)
	_ = os.WriteFile(storeFile, b, 0600)
}

// ── User operations ──────────────────────────────────────────────────────────

// GetOrCreateGuest returns the existing guest with deviceID, or creates one.
func GetOrCreateGuest(deviceID string) *User {
	dataMu.Lock()
	defer dataMu.Unlock()
	for _, u := range appDB.Users {
		if u.DeviceID == deviceID {
			return u
		}
	}
	id := newUUID()
	u := &User{
		ID:        id,
		DeviceID:  deviceID,
		Name:      id[:10],
		IsGuest:   true,
		CreatedAt: time.Now(),
	}
	appDB.Users[id] = u
	persist()
	return u
}

// RegisterUser upgrades a guest (same deviceID) or creates a new registered user.
func RegisterUser(deviceID, username, password string) (*User, error) {
	dataMu.Lock()
	defer dataMu.Unlock()

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
	for _, u := range appDB.Users {
		if strings.EqualFold(u.Username, uname) {
			return nil, fmt.Errorf("tên đăng nhập đã được sử dụng")
		}
	}

	hash, err := hashPass(password)
	if err != nil {
		return nil, fmt.Errorf("lỗi hệ thống")
	}

	// Upgrade existing guest with same device, or create fresh
	var user *User
	for _, u := range appDB.Users {
		if u.DeviceID == deviceID && u.IsGuest {
			user = u
			break
		}
	}
	if user == nil {
		id := newUUID()
		user = &User{ID: id, DeviceID: deviceID, CreatedAt: time.Now()}
		appDB.Users[id] = user
	}
	user.Username = uname
	user.Name = uname
	user.PassHash = hash
	user.IsGuest = false
	persist()
	return user, nil
}

// LoginUser authenticates by username + password.
func LoginUser(username, password string) (*User, error) {
	dataMu.RLock()
	defer dataMu.RUnlock()
	for _, u := range appDB.Users {
		if strings.EqualFold(u.Username, username) {
			if !checkPass(password, u.PassHash) {
				return nil, fmt.Errorf("sai mật khẩu")
			}
			return u, nil
		}
	}
	return nil, fmt.Errorf("không tìm thấy tài khoản")
}

// RenameUser changes the display name of the user.
func RenameUser(userID, newName string) error {
	dataMu.Lock()
	defer dataMu.Unlock()
	name := strings.TrimSpace(newName)
	if name == "" {
		return fmt.Errorf("tên không được để trống")
	}
	if len([]rune(name)) > 20 {
		return fmt.Errorf("tên tối đa 20 ký tự")
	}
	u, ok := appDB.Users[userID]
	if !ok {
		return fmt.Errorf("không tìm thấy người dùng")
	}
	u.Name = name
	persist()
	return nil
}

// GetUserByID returns the user with the given ID.
func GetUserByID(id string) (*User, bool) {
	dataMu.RLock()
	defer dataMu.RUnlock()
	u, ok := appDB.Users[id]
	return u, ok
}

// ── Score operations ───────────────────────────────────────────────────────

// AddScore records a score entry for the user.
func AddScore(userID string, correct, total, level int) {
	dataMu.Lock()
	defer dataMu.Unlock()
	appDB.Scores = append(appDB.Scores, Score{
		ID:       newUUID(),
		UserID:   userID,
		Correct:  correct,
		Total:    total,
		Level:    level,
		ScoredAt: time.Now(),
	})
	persist()
}

// GetRanking returns the leaderboard for the given time period.
// period: "day" | "week" | "month" | "year"
func GetRanking(period string) []RankEntry {
	now := time.Now()
	var cutoff time.Time
	switch period {
	case "week":
		cutoff = now.AddDate(0, 0, -7)
	case "month":
		cutoff = now.AddDate(0, -1, 0)
	case "year":
		cutoff = now.AddDate(-1, 0, 0)
	default: // day – from midnight today
		y, m, d := now.Date()
		cutoff = time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	}

	dataMu.RLock()
	defer dataMu.RUnlock()

	totals := make(map[string]int)
	for _, s := range appDB.Scores {
		if s.ScoredAt.After(cutoff) {
			totals[s.UserID] += s.Correct
		}
	}

	type kv struct {
		id    string
		score int
	}
	kvs := make([]kv, 0, len(totals))
	for id, score := range totals {
		kvs = append(kvs, kv{id, score})
	}
	sort.Slice(kvs, func(i, j int) bool { return kvs[i].score > kvs[j].score })

	entries := make([]RankEntry, 0, len(kvs))
	for i, kv := range kvs {
		name, isGuest := kv.id[:min(10, len(kv.id))], true
		if u, ok := appDB.Users[kv.id]; ok {
			name = u.Name
			isGuest = u.IsGuest
		}
		entries = append(entries, RankEntry{
			Rank:    i + 1,
			UserID:  kv.id,
			Name:    name,
			Score:   kv.score,
			IsGuest: isGuest,
		})
	}
	return entries
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
