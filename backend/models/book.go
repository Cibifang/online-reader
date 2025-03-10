package models

type Book struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

var Books []Book
