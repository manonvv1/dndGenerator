// Layer: Infrastructure (persistence adapter: read/write characters.json)

package main

import (
	"encoding/json"
	"os"
	"strings"

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

func findCharLike(name string) *Character {
	q := strings.ToLower(strings.TrimSpace(name))
	if q == "" {
		return nil
	}
	for i := range characters {
		if strings.EqualFold(characters[i].Name, name) {
			return &characters[i]
		}
	}
	for i := range characters {
		if strings.Contains(strings.ToLower(characters[i].Name), q) {
			return &characters[i]
		}
	}
	return nil
}
