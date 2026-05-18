package app

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	PaperlessHost       string `json:"paperlessHost"`
	Token               string `json:"token"`
	AmountCustomFieldId int    `json:"amountCustomFieldId"`
	Titles              struct {
		Title1 map[string]string `json:"title1"`
		Title2 map[string]string `json:"title2"`
		Title3 map[string]string `json:"title3"`
		Title4 map[string]string `json:"title4"`
	} `json:"titles"`
	YearKeywords   []string `json:"yearKeywords"`
	AmountKeywords []string `json:"amountKeywords"`
}

// Global variables that will be loaded from config
var (
	PaperlessHost       string
	PaperlessToken      string
	AmountCustomFieldId int
	Title1              map[string]string
	Title2              map[string]string
	Title3              map[string]string
	Title4              map[string]string
	YearKeywords        []string
	AmountKeywords      []string
)

// LoadConfig loads configuration from a JSON file
func LoadConfig(filePath string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading config file %s: %w", filePath, err)
	}

	var config Config
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	// Assign loaded values to global variables
	PaperlessHost = config.PaperlessHost
	PaperlessToken = config.Token
	AmountCustomFieldId = config.AmountCustomFieldId
	Title1 = config.Titles.Title1
	Title2 = config.Titles.Title2
	Title3 = config.Titles.Title3
	Title4 = config.Titles.Title4
	YearKeywords = config.YearKeywords
	AmountKeywords = config.AmountKeywords

	fmt.Println("Configuration loaded successfully from", filePath)
	return nil
}
