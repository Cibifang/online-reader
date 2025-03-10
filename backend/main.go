package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/Cibifang/online-reader/backend/config"
	"github.com/Cibifang/online-reader/backend/handlers"
)

func init() {
	config.LoadConfig()
	os.MkdirAll("uploads", os.ModePerm)
}

func main() {
	log.Println("Starting server...")
	log.Printf("Using Youdao AppKey: %s", config.AppConfig.YoudaoAppKey)

	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/upload", handlers.HandleUpload).Methods("POST")
	r.HandleFunc("/api/books", handlers.HandleGetBooks).Methods("GET")
	r.HandleFunc("/api/books/{id}", handlers.HandleGetBook).Methods("GET")
	r.HandleFunc("/api/translate", handlers.HandleTranslate).Methods("POST")
	r.HandleFunc("/api/words", handlers.HandleGetWords).Methods("GET")
	r.HandleFunc("/api/words", handlers.HandleSaveWord).Methods("POST")

	// Use CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	handler := c.Handler(r)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
