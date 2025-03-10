package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Cibifang/online-reader/backend/models"
	"github.com/gorilla/mux"
)

// HandleUpload handles file upload requests
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := handler.Filename
	path := filepath.Join("uploads", filename)

	dst, err := os.Create(path)
	if err != nil {
		http.Error(w, "Error creating the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error copying the file", http.StatusInternalServerError)
		return
	}

	book := models.Book{
		ID:    fmt.Sprintf("%d", len(models.Books)+1),
		Title: strings.TrimSuffix(filename, filepath.Ext(filename)),
		Path:  path,
	}
	models.Books = append(models.Books, book)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "File uploaded successfully",
		"book":    book,
	})
	log.Printf("Book uploaded: %s", book.Title)
}

// HandleGetBooks handles requests to get all books
func HandleGetBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Books)
	log.Printf("Fetched %d books", len(models.Books))
}

// HandleGetBook handles requests to get a specific book
func HandleGetBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	for _, book := range models.Books {
		if book.ID == id {
			content, err := os.ReadFile(book.Path)
			if err != nil {
				http.Error(w, "Error reading the book", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"book":    book,
				"content": string(content),
			})
			log.Printf("Fetched book: %s", book.Title)
			return
		}
	}

	http.Error(w, "Book not found", http.StatusNotFound)
}
