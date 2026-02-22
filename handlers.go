package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(lesson)
}
