package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Config represents the configuration structure
type Config struct {
	YoudaoAppKey    string `json:"YoudaoAppKey"`
	YoudaoAppSecret string `json:"YoudaoAppSecret"`
}

// Global configuration variable
var config Config

// Book represents a book in our system
type Book struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

// WordStatus represents the learning status of a word
type WordStatus string

const (
	StatusFamiliar   WordStatus = "familiar"
	StatusUnfamiliar WordStatus = "unfamiliar"
	StatusLearning   WordStatus = "learning"
)

// Word represents a word with its translation and status
type Word struct {
	Text        string     `json:"text"`
	Translation string     `json:"translation"`
	Status      WordStatus `json:"status"`
}

// YoudaoTranslateResponse represents the response from Youdao Translate API
type YoudaoTranslateResponse struct {
	ErrorCode   string   `json:"errorCode"`
	Query       string   `json:"query"`
	Translation []string `json:"translation"`
	Basic       struct {
		Explains []string `json:"explains"`
	} `json:"basic"`
	Web []struct {
		Key   string   `json:"key"`
		Value []string `json:"value"`
	} `json:"web"`
}

var books []Book
var words map[string]Word

func init() {
	// Load configuration
	loadConfig()

	// Create uploads directory if it doesn't exist
	os.MkdirAll("uploads", os.ModePerm)

	// Initialize words map
	words = make(map[string]Word)

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
}

func loadConfig() {
	log.Println("Loading configuration...")

	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	// Construct the full path to the config file
	configPath := filepath.Join(dir, "config.json")
	log.Printf("Looking for config file at: %s", configPath)

	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	if config.YoudaoAppKey == "your-app-key" || config.YoudaoAppSecret == "your-app-secret" {
		log.Println("Warning: Youdao API credentials are not set. Please update the config.json file.")
	} else {
		log.Println("Youdao API credentials loaded successfully.")
	}
}

func main() {
	log.Println("Starting server...")
	log.Printf("Using Youdao AppKey: %s", config.YoudaoAppKey)

	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/upload", handleUpload).Methods("POST")
	r.HandleFunc("/api/books", handleGetBooks).Methods("GET")
	r.HandleFunc("/api/books/{id}", handleGetBook).Methods("GET")
	r.HandleFunc("/api/translate", handleTranslate).Methods("POST")
	r.HandleFunc("/api/words", handleGetWords).Methods("GET")
	r.HandleFunc("/api/words", handleSaveWord).Methods("POST")

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

func handleUpload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form, 10 << 20 specifies a maximum upload of 10 MB
	r.ParseMultipartForm(10 << 20)

	// Get file from request
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create file path
	filename := handler.Filename
	path := filepath.Join("uploads", filename)

	// Create file
	dst, err := os.Create(path)
	if err != nil {
		http.Error(w, "Error creating the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error copying the file", http.StatusInternalServerError)
		return
	}

	// Create book entry
	book := Book{
		ID:    fmt.Sprintf("%d", len(books)+1),
		Title: strings.TrimSuffix(filename, filepath.Ext(filename)),
		Path:  path,
	}
	books = append(books, book)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "File uploaded successfully",
		"book":    book,
	})
	log.Printf("Book uploaded: %s", book.Title)
}

func handleGetBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
	log.Printf("Fetched %d books", len(books))
}

func handleGetBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	for _, book := range books {
		if book.ID == id {
			// Read book content
			content, err := os.ReadFile(book.Path)
			if err != nil {
				http.Error(w, "Error reading the book", http.StatusInternalServerError)
				return
			}

			// Return book content
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

func handleTranslate(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Word string `json:"word"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Clean the word (remove punctuation, etc.)
	cleanWord := strings.TrimFunc(requestData.Word, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	})

	if cleanWord == "" {
		http.Error(w, "Invalid word", http.StatusBadRequest)
		return
	}

	// Check if we already have this word
	word, exists := words[cleanWord]

	// If the word doesn't exist or we need to update the translation
	if !exists || word.Translation == "" {
		// Get translation from Youdao API
		translation, err := translateWithYoudao(cleanWord)
		if err != nil {
			log.Printf("Error translating word: %v", err)
			// Fallback to mock translation
			translation = "翻译: " + cleanWord
		}

		if !exists {
			word = Word{
				Text:   cleanWord,
				Status: StatusUnfamiliar,
			}
		}
		word.Translation = translation
		words[cleanWord] = word
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"word":        cleanWord,
		"translation": word.Translation,
		"status":      word.Status,
	})
	log.Printf("Translated word: %s", cleanWord)
}

func translateWithYoudao(word string) (string, error) {
	// If API credentials are not set, return mock translation
	if config.YoudaoAppKey == "your-app-key" || config.YoudaoAppSecret == "your-app-secret" {
		return "请先设置有道翻译API密钥", nil
	}

	// Prepare request parameters
	salt := fmt.Sprintf("%d", rand.Intn(10000))
	curtime := fmt.Sprintf("%d", time.Now().Unix())

	// Calculate sign
	signStr := config.YoudaoAppKey + truncate(word) + salt + curtime + config.YoudaoAppSecret
	sign := md5Sum(signStr)

	// Build request URL
	apiURL := "https://openapi.youdao.com/api"
	data := url.Values{}
	data.Set("q", word)
	data.Set("from", "en")
	data.Set("to", "zh-CHS")
	data.Set("appKey", config.YoudaoAppKey)
	data.Set("salt", salt)
	data.Set("sign", sign)
	data.Set("signType", "v3")
	data.Set("curtime", curtime)

	// Log request details for debugging
	log.Printf("Youdao API Request: URL=%s, Word=%s, AppKey=%s, Salt=%s, CurTime=%s, Sign=%s",
		apiURL, word, config.YoudaoAppKey, salt, curtime, sign)

	// Send request
	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return "", fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response body: %v", err)
	}

	// Log raw response for debugging
	log.Printf("Youdao API Raw Response: %s", string(body))

	// Parse response
	var result YoudaoTranslateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("JSON parsing error: %v", err)
	}

	// Check for errors
	if result.ErrorCode != "0" {
		return "", fmt.Errorf("API error: %s", result.ErrorCode)
	}

	// Build translation string
	var translation strings.Builder

	// Add basic translations
	if len(result.Translation) > 0 {
		translation.WriteString(strings.Join(result.Translation, ", "))
	}

	// Add explanations if available
	if result.Basic.Explains != nil && len(result.Basic.Explains) > 0 {
		translation.WriteString("\n解释: ")
		translation.WriteString(strings.Join(result.Basic.Explains, ", "))
	}

	// Add web translations if available
	if result.Web != nil && len(result.Web) > 0 {
		translation.WriteString("\n网络释义:\n")
		for _, item := range result.Web {
			translation.WriteString("- ")
			translation.WriteString(item.Key)
			translation.WriteString(": ")
			translation.WriteString(strings.Join(item.Value, ", "))
			translation.WriteString("\n")
		}
	}

	return translation.String(), nil
}

// Helper function to truncate string for Youdao API
func truncate(q string) string {
	if len(q) <= 20 {
		return q
	}
	return q[:10] + fmt.Sprintf("%d", len(q)) + q[len(q)-10:]
}

// Helper function to calculate MD5 hash
func md5Sum(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

func handleGetWords(w http.ResponseWriter, r *http.Request) {
	wordsList := make([]Word, 0, len(words))
	for _, word := range words {
		wordsList = append(wordsList, word)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wordsList)
	log.Printf("Fetched %d words", len(wordsList))
}

func handleSaveWord(w http.ResponseWriter, r *http.Request) {
	var word Word
	if err := json.NewDecoder(r.Body).Decode(&word); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update or add the word, preserving existing translation
	existingWord, exists := words[word.Text]
	if exists {
		word.Translation = existingWord.Translation
	}
	words[word.Text] = word

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(word)
	log.Printf("Saved word: %s with status: %s", word.Text, word.Status)
}
