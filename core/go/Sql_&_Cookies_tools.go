package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/core/structs"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// Function to Create / initialize the dtabase
func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./databases/forum.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			role TEXT
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Function to insert user data in the database
func insertUser(db *sql.DB, email string, username string, password string, role string) error {
	// We hash the password for more security
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (email, username, password, role) VALUES(?, ?, ?, ?)", email, username, hashedPassword, role)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return errors.New("username ou email déjà connu")
		} else {
			fmt.Println(err)
		}
	}
	return err
}

// Function to verify user data when he login
func verifyLog(db *sql.DB, username string, email string, password string) (structs.User, error) {
	var hashedPassword string
	var userData structs.User
	err := db.QueryRow("SELECT password, username, email, role FROM users WHERE username = ? OR email = ?", username, email).Scan(&hashedPassword, &userData.Username, &userData.Email, &userData.Role)
	if err != nil {
		return userData, err
	}
	// We compare the hash to the real password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return userData, err
}

// Function to get the username from the database
func getUsername(db *sql.DB, username string) string {
	var userData structs.User
	err := db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&userData.Username)
	if err != nil {
		return ""
	}
	return userData.Username
}

// Function to get the email from the database
func getEmail(db *sql.DB, email string) string {
	var userData structs.User
	err := db.QueryRow("SELECT email FROM users WHERE email = ?", email).Scan(&userData.Email)
	if err != nil {
		return ""
	}
	return userData.Email
}

// Function to get the role from the database
func getRole(db *sql.DB, username string) string {
	var userData structs.User
	err := db.QueryRow("SELECT role FROM users WHERE username = ?", username).Scan(&userData.Role)
	if err != nil {
		return ""
	}
	return userData.Role
}

// Function to get all of the users data from the database
func getAllUsers(db *sql.DB) ([]structs.User, error) {
	rows, err := db.Query("SELECT username, role FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []structs.User
	for rows.Next() {
		var user structs.User
		if err := rows.Scan(&user.Username, &user.Role); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Function to modify the username from the database
func modifyUsername(db *sql.DB, newUsername string, oldUsername string) error {
	_, err := db.Exec("UPDATE users SET username = ? WHERE username = ?", newUsername, oldUsername)
	if err != nil {
		fmt.Println("Error updating user:", err)
		return err
	}

	err = updatePostsUsername(db, newUsername, oldUsername)
	if err != nil {
		fmt.Println("Error updating posts:", err)
		return err
	}

	err = updateCommentsUsername(db, newUsername, oldUsername)
	if err != nil {
		fmt.Println("Error updating comments:", err)
		return err
	}
	err = updateLikeUseranem(db, oldUsername, newUsername)
	if err != nil {
		fmt.Println("Error updating likes:", err)
		return err
	}

	return nil
}

func updatePostsUsername(db *sql.DB, newUsername string, oldUsername string) error {
	_, err := db.Exec("UPDATE post SET user = ? WHERE user = ?", newUsername, oldUsername)
	if err != nil {
		return fmt.Errorf("error updating posts: %w", err)
	}
	return nil
}

func updateCommentsUsername(db *sql.DB, newUsername string, oldUsername string) error {
	_, err := db.Exec("UPDATE comment SET user = ? WHERE user = ?", newUsername, oldUsername)
	if err != nil {
		return fmt.Errorf("error updating comments: %w", err)
	}
	return nil
}

// Function to modify the email from the database
func modifyEmail(db *sql.DB, newEmail string, oldEmail string) error {
	_, err := db.Exec("UPDATE users SET email = ? WHERE email = ?", newEmail, oldEmail)
	if err != nil {
		fmt.Println("Error updating user:", err)
	}
	return err
}

// Function to modify the role from the database
func modifyRole(db *sql.DB, username string) error {
	var currentRole string
	err := db.QueryRow("SELECT role FROM users WHERE username = ?", username).Scan(&currentRole)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var newRole string
	if currentRole == "admin" {
		newRole = "admin"
	} else if currentRole == "user" {
		newRole = "moderator"
	} else {
		newRole = "user"
	}

	_, err = db.Exec("UPDATE users SET role = ? WHERE username = ?", newRole, username)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Function to delete the email from the database
func deleteRole(db *sql.DB, username string, role string) error {
	_, err := db.Exec("UPDATE users SET role = ? WHERE username = ?", role, username)
	if err != nil {
		fmt.Println("Error updating user:", err)
	}
	return err
}

// Function to get the list of post compared to a username
func GetListPostByUsername(db *sql.DB, username string) (list_Post, error) {
	var listPost list_Post

	// Query to get the posts by category_id
	postsQuery := `SELECT id, user, text, title, imageURL, category_id FROM post WHERE user = ?`
	rows, err := db.Query(postsQuery, username)
	if err != nil {
		return listPost, err
	}
	defer rows.Close()

	for rows.Next() {
		var post structs.Post
		err := rows.Scan(&post.ID, &post.User, &post.Text, &post.Title, &post.ImageURL, &post.SelectedCategory)
		if err != nil {
			return listPost, err
		}
		listPost.Posts = append(listPost.Posts, post)
	}

	if err = rows.Err(); err != nil {
		return listPost, err
	}

	return listPost, nil
}

// Function to get the 3 most recent Posts
func GetListRecentPost(db *sql.DB) (list_Post, error) {
	var listPost list_Post

	// Query to get 3 most recent posts
	postsQuery := `SELECT id, user, text, title, imageURL, category_id FROM post ORDER BY date DESC LIMIT 3`
	rows, err := db.Query(postsQuery)
	if err != nil {
		return listPost, err
	}
	defer rows.Close()

	for rows.Next() {
		var post structs.Post
		err := rows.Scan(&post.ID, &post.User, &post.Text, &post.Title, &post.ImageURL, &post.SelectedCategory)
		if err != nil {
			return listPost, err
		}
		listPost.Posts = append(listPost.Posts, post)
	}

	if err = rows.Err(); err != nil {
		return listPost, err
	}

	return listPost, nil
}

// Function to create a cookie
func CreateCookie(w http.ResponseWriter, name string, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",   // For all of the pages
		MaxAge:   86400, // For 1 day (86400 sec = 1 day)
		HttpOnly: true,  // Prevents JS code from accessing cookies carrying this notion
		Secure:   true,  // Cookies bearing the Secure flag are only sent if the request is https;
	}
	http.SetCookie(w, cookie)
}

// Function to delete a cookie
func DeleteCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(w, cookie)
}

// Function to get a cookie
func getCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, err
}

// Function to verify if the user has a cookie
func verifyCookie(r *http.Request) bool {
	cookie, err := getCookie(r, "session_token")
	if len(cookie) == 0 || err != nil {
		return false
	} else {
		return true
	}
}
