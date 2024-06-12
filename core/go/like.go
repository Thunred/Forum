package forum

import (
	"database/sql"
)

func initDBlike(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS like (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id TEXT NOT NULL,
		user_name TEXT NOT NULL,
		is_like BOOLEAN NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func insertLike(db *sql.DB, post_id string, user_name string, is_like bool) error {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM like WHERE post_id = ? AND user_name = ? AND is_like = ?)", post_id, user_name, is_like).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	_, err = db.Exec("INSERT INTO like (post_id, user_name, is_like) VALUES(?, ?, ?)", post_id, user_name, is_like)
	if err != nil {
		return err
	}

	return nil
}

func deleteLike(db *sql.DB, post_id string, user_name string) error {
	_, err := db.Exec("DELETE FROM like WHERE post_id = ? AND user_name = ?", post_id, user_name)
	if err != nil {
		return err
	}
	return nil
}

func countLike(db *sql.DB, post_id string) (int, error) {
	rows, err := db.Query("SELECT COUNT(*) FROM like WHERE post_id = ? AND is_like = 1", post_id)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

func countDislike(db *sql.DB, post_id string) (int, error) {
	rows, err := db.Query("SELECT COUNT(*) FROM like WHERE post_id = ? AND is_like = 0", post_id)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

func updateLikeUseranem(db *sql.DB, old_username string, new_username string) error {
	_, err := db.Exec("UPDATE like SET user_name = ? WHERE user_name = ?", new_username, old_username)
	if err != nil {
		return err
	}
	return nil
}