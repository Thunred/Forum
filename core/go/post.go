package forum

import (
	"database/sql"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"
)

type Post struct {
	User     string
	Text     string
	Title    string
	ImageURL string
}

func initDBPost() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./databases/forum.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS post (
			id TEXT PRIMARY KEY,
			user TEXT NOT NULL,
			text TEXT NOT NULL,
			title TEXT NOT NULL,
			imageURL TEXT, 
			UNIQUE(id)
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func insertPost(db *sql.DB, user string, text string, title string, imageURL string) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO post (id, user, text, title,imageURL) VALUES(?, ?, ?, ?, ?)", id.String(), user, text, title, imageURL)
	if err != nil {
		return err
	}

	return nil
}

const uploadPath = "databases/upload_image"

func (p Post) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := initDBPost()
	if err != nil {
		return
	}
	defer db.Close()

	var t *template.Template

	switch r.URL.Path {
	case "/post":
		if r.Method == "POST" {
			err := r.ParseMultipartForm(20 << 20)
			if err != nil {
				http.Error(w, "Erreur lors de l'analyse du formulaire", http.StatusInternalServerError)
				return
			}
			p.Title = r.FormValue("title")
			p.Text = r.FormValue("content")

			file, fileHeader, err := r.FormFile("image")
			if err != nil && err != http.ErrMissingFile {
				http.Error(w, "Erreur lors de la récupération du fichier", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			if err == nil {
				ext := filepath.Ext(fileHeader.Filename)
				allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
				if !allowedExts[ext] {
					http.Error(w, "Extension de fichier non autorisée", http.StatusBadRequest)
					return
				}

				fileSize := fileHeader.Size
				var maxFileSize int64 = 20 * 1024 * 1024
				if fileSize > maxFileSize {
					http.Error(w, "Image trop grande (max 20 Mo)", http.StatusBadRequest)
					return
				}
				fileID, err := generateUniqueFilename(uploadPath, ext)
				if err != nil {
					http.Error(w, "Erreur lors de la génération de l'ID de fichier unique", http.StatusInternalServerError)
					return
				}
				filePath := filepath.Join(uploadPath, fileID+ext)
				outFile, err := os.Create(filePath)
				if err != nil {
					http.Error(w, "Erreur lors de la création du fichier", http.StatusInternalServerError)
					return
				}
				defer outFile.Close()

				_, err = io.Copy(outFile, file)
				if err != nil {
					http.Error(w, "Erreur lors de la copie des données du fichier", http.StatusInternalServerError)
					return
				}

				p.ImageURL = "databases/upload_image" + fileID + ext
			}

			err = insertPost(db, p.User, p.Text, p.Title, p.ImageURL)
			if err != nil {
				http.Error(w, "Erreur lors de l'insertion du post dans la base de données", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/", http.StatusFound)
		}
		t, _ = template.ParseFiles("src/html/post.html")
	default:
		http.NotFound(w, r)
		return
	}

	t.Execute(w, p)
}

func generateUniqueFilename(uploadPath string, ext string) (string, error) {
	for {
		fileID := uuid.Must(uuid.NewV4()).String()
		filePath := filepath.Join(uploadPath, fileID+ext)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fileID, nil
		}
	}
}