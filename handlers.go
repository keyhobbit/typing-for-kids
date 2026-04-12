package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// jsStr JSON-encodes s and marks it template.JS so html/template does NOT
// apply a second round of JS escaping.  Use it to embed Go strings safely
// as JS string literals:  let x = {{jsStr .SomeField}};
func jsStr(s string) template.JS {
	b, err := json.Marshal(s)
	if err != nil {
		return template.JS(`""`)
	}
	return template.JS(b)
}

var tmpl = template.Must(
	template.New("index.html").
		Funcs(template.FuncMap{"jsStr": jsStr}).
		ParseFiles("templates/index.html"),
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	levelStr := r.URL.Query().Get("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 || level > 3 {
		level = 1
	}
	lesson := GetRandomLesson(level)
	data := struct {
		Lesson Lesson
		Level  int
	}{
		Lesson: lesson,
		Level:  level,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleNext(w http.ResponseWriter, r *http.Request) {
	levelStr := r.URL.Query().Get("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 || level > 3 {
		level = 1
	}
	lesson := GetRandomLesson(level)
	writeJSON(w, lesson)
}

// ── Auth helpers ───────────────────────────────────────────────────────────

func authUser(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", false
	}
	return ResolveToken(strings.TrimPrefix(auth, "Bearer "))
}

// ── POST /api/auth/guest ───────────────────────────────────────────────────

func handleAuthGuest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.DeviceID) == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	u := GetOrCreateGuest(body.DeviceID)
	token := NewToken(u.ID)
	writeJSON(w, map[string]any{"user": publicUser(u), "token": token})
}

// ── POST /api/auth/register ────────────────────────────────────────────────

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		DeviceID string `json:"device_id"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	u, err := RegisterUser(body.DeviceID, body.Username, body.Password)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	token := NewToken(u.ID)
	writeJSON(w, map[string]any{"user": publicUser(u), "token": token})
}

// ── POST /api/auth/login ───────────────────────────────────────────────────

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	u, err := LoginUser(body.Username, body.Password)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusUnauthorized)
		return
	}
	token := NewToken(u.ID)
	writeJSON(w, map[string]any{"user": publicUser(u), "token": token})
}

// ── POST /api/auth/logout ──────────────────────────────────────────────────

func handleLogout(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		RevokeToken(strings.TrimPrefix(auth, "Bearer "))
	}
	writeJSON(w, map[string]any{"ok": true})
}

// ── PUT /api/user/name ─────────────────────────────────────────────────────

func handleRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := authUser(r)
	if !ok {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := RenameUser(userID, body.Name); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	u, _ := GetUserByID(userID)
	writeJSON(w, map[string]any{"user": publicUser(u)})
}

// ── POST /api/score ────────────────────────────────────────────────────────

func handleScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := authUser(r)
	if !ok {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Correct int `json:"correct"`
		Total   int `json:"total"`
		Level   int `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Correct < 0 || body.Total < 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	AddScore(userID, body.Correct, body.Total, body.Level)
	writeJSON(w, map[string]any{"ok": true})
}

// ── GET /api/ranking ───────────────────────────────────────────────────────

func handleRanking(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	switch period {
	case "day", "week", "month", "year":
	default:
		period = "day"
	}
	entries := GetRanking(period)
	writeJSON(w, entries)
}

// ── JSON helpers ───────────────────────────────────────────────────────────

type publicUserData struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IsGuest bool   `json:"is_guest"`
}

func publicUser(u *User) publicUserData {
	return publicUserData{ID: u.ID, Name: u.Name, IsGuest: u.IsGuest}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
