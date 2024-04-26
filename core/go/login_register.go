package forum

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Email    string
	Username string
	Password string
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./databases/forumRegister.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func insertUser(db *sql.DB, email string, username string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (email, username, password) VALUES(?, ?, ?)", email, username, hashedPassword)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func verifyLog(db *sql.DB, username string, password string) bool {
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ServeHTTP for register page
func (u user) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDB()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template

	if r.URL.Path == "/register" {
		if r.Method == "POST" {
			u.Username = strings.TrimSpace(r.FormValue("username"))
			u.Email = strings.TrimSpace(r.FormValue("email"))
			u.Password = r.FormValue("password")

			insertUser(db, u.Email, u.Username, u.Password)
			fmt.Println(u.Email, u.Username, u.Password)
		}
		t, _ = template.ParseFiles("src/html/register.html")
	}

	if r.URL.Path == "/login" {
		if r.Method == "POST" {
			u.Username = strings.TrimSpace(r.FormValue("username"))
			u.Password = r.FormValue("password")

			if verifyLog(db, u.Username, u.Password) {
				http.Redirect(w, r, "/profile", http.StatusFound)
				return
			}
		}
		t, _ = template.ParseFiles("src/html/login.html")
	}

	if r.URL.Path == "/profile" {
		t, _ = template.ParseFiles("src/html/profile.html")
	}

	t.Execute(w, u)

}
