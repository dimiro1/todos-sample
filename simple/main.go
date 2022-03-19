package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID          int
	Title       string
	IsCompleted bool
	CreatedAt   time.Time
}

func ListTodos(db *sql.DB) ([]*Todo, error) {
	rows, err := db.Query("SELECT * FROM todos ORDER BY id desc")
	if err != nil {
		return []*Todo{}, err
	}
	defer func() {
		_ = rows.Close()
	}()

	todos := []*Todo{}

	for rows.Next() {
		var id int
		var title string
		var isCompleted int
		var createdAtStr string

		if err := rows.Scan(&id, &title, &isCompleted, &createdAtStr); err != nil {
			return []*Todo{}, err
		}

		createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return []*Todo{}, err
		}

		todos = append(todos, &Todo{
			ID:          id,
			Title:       title,
			IsCompleted: isCompleted == 1,
			CreatedAt:   createdAt,
		})
	}

	return todos, err
}

func CompleteTodo(db *sql.DB, id int) error {
	_, err := db.Exec("UPDATE todos SET is_completed=1 WHERE id = ?", id)
	return err
}

func AddNewTodo(db *sql.DB, title string) error {
	_, err := db.Exec("INSERT INTO todos (title, is_completed, created_at) VALUES (?, ?, datetime('now'))", title, 0)
	if err != nil {
		return err
	}
	return nil
}

func CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    is_completed INTEGER,
    created_at TEXT
)`)
	return err
}

func main() {
	// Connect to the sqlite database
	db, err := sql.Open("sqlite3", "file:database.sqlite3")
	if err != nil {
		panic(fmt.Sprintf("not able to connect to database: %s", err.Error()))
	}

	// Create the required tables
	if err := CreateTables(db); err != nil {
		panic(err)
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
