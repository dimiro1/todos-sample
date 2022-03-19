package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Todo struct {
	ID          int `gorm:"primarykey"`
	Title       string
	IsCompleted bool
	CreatedAt   time.Time
}

func ListTodos(db *gorm.DB) ([]*Todo, error) {
	todos := []*Todo{}
	result := db.Find(&todos)

	return todos, result.Error
}

func CompleteTodo(db *gorm.DB, id int) error {
	result := db.Model(&Todo{}).Where("id = ?", id).Update("is_completed", true)
	return result.Error
}

func AddNewTodo(db *gorm.DB, title string) error {
	todo := &Todo{
		Title:       title,
		IsCompleted: false,
		CreatedAt:   time.Now(),
	}

	result := db.Create(todo)

	return result.Error
}

func main() {
	// Connect to the sqlite database
	db, err := gorm.Open(sqlite.Open("database_gorm.sqlite3"), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("not able to connect to database: %s", err.Error()))
	}

	// Create the required tables
	if err := db.AutoMigrate(&Todo{}); err != nil {
		panic(fmt.Sprintf("not able to create table: %s", err.Error()))
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		todos, err := ListTodos(db)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		indexTmpl := template.Must(template.ParseFiles("templates/index.gohtml"))

		w.Header().Set("Content-Type", "text/html")

		_ = indexTmpl.Execute(w, todos)
	})

	http.HandleFunc("/done-todo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusPermanentRedirect)
			return
		}

		idStr := r.FormValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "id must be a number", http.StatusBadRequest)
			return
		}

		if err := CompleteTodo(db, id); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})

	http.HandleFunc("/new-todo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusPermanentRedirect)
			return
		}

		title := r.FormValue("title")

		if err := AddNewTodo(db, title); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})

	if err := http.ListenAndServe(":9090", nil); err != nil {
		panic(err.Error())
	}
}
