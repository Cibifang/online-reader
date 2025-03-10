package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Cibifang/online-reader/backend/models"
	"github.com/Cibifang/online-reader/backend/utils"
)

func HandleTranslate(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Word string `json:"word"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cleanWord := strings.TrimFunc(requestData.Word, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	})

	if cleanWord == "" {
		http.Error(w, "Invalid word", http.StatusBadRequest)
		return
	}

	word, exists := models.Words[cleanWord]

	if !exists || word.Translation == "" {
		translation, err := utils.TranslateWithYoudao(cleanWord)
		if err != nil {
			log.Printf("Error translating word: %v", err)
			translation = "翻译: " + cleanWord
		}

		if !exists {
			word = models.Word{
				Text:   cleanWord,
				Status: models.StatusUnfamiliar,
			}
		}
		word.Translation = translation
		models.Words[cleanWord] = word
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"word":        cleanWord,
		"translation": word.Translation,
		"status":      word.Status,
	})
	log.Printf("Translated word: %s", cleanWord)
}

func HandleGetWords(w http.ResponseWriter, r *http.Request) {
	wordsList := make([]models.Word, 0, len(models.Words))
	for _, word := range models.Words {
		wordsList = append(wordsList, word)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wordsList)
	log.Printf("Fetched %d words", len(wordsList))
}

func HandleSaveWord(w http.ResponseWriter, r *http.Request) {
	var word models.Word
	if err := json.NewDecoder(r.Body).Decode(&word); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	existingWord, exists := models.Words[word.Text]
	if exists {
		word.Translation = existingWord.Translation
	}
	models.Words[word.Text] = word

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(word)
	log.Printf("Saved word: %s with status: %s", word.Text, word.Status)
}
