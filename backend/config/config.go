package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Config represents the configuration structure
type Config struct {
	YoudaoAppKey    string `json:"YoudaoAppKey"`
	YoudaoAppSecret string `json:"YoudaoAppSecret"`
}

// AppConfig is the global configuration variable
var AppConfig Config

// LoadConfig loads the configuration from config.json
func LoadConfig() {
	log.Println("Loading configuration...")

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	configPath := filepath.Join(dir, "config.json")
	log.Printf("Looking for config file at: %s", configPath)

	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = json.Unmarshal(file, &AppConfig)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	if AppConfig.YoudaoAppKey == "your-app-key" || AppConfig.YoudaoAppSecret == "your-app-secret" {
		log.Println("Warning: Youdao API credentials are not set. Please update the config.json file.")
	} else {
		log.Println("Youdao API credentials loaded successfully.")
	}
}
