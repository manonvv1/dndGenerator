package main

import (
	"encoding/json"
	"os"
)

var characters []Character
const dbFile = "characters.json"

/**
*   loadCharacters loads the character from the DB
**/
func loadCharacters() {
	data, err := os.ReadFile(dbFile)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &characters)
}

/**
*   saveCharacters persists the character DB
**/
func saveCharacters() {
	data, _ := json.MarshalIndent(characters, "", "  ")
	_ = os.WriteFile(dbFile, data, 0644)
}