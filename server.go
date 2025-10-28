// Layer: Infrastructure / UI (HTTP transport/controller layer)

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type createRequest struct {
	Name          string         `json:"name"`
	Race          string         `json:"race"`
	Class         string         `json:"class"`
	Level         int            `json:"level"`
	Background    string         `json:"background"`
	AbilityScores *AbilityScores `json:"ability_scores,omitempty"`
	Skills        []string       `json:"skills,omitempty"`
	Weapon        string         `json:"weapon,omitempty"`
	Armor         string         `json:"armor,omitempty"`
	Shield        string         `json:"shield,omitempty"`
	OffHand       string         `json:"offhand,omitempty"`
}

type apiError struct {
	Error string `json:"error"`
}

/**
*  writeJSON writes a JSON response with status code and content type
**/
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}


func handleCharactersGet(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		writeJSON(w, http.StatusOK, characters)
		return
	}
	if c := findCharLike(name); c != nil {
		writeJSON(w, http.StatusOK, c)
		return
	}
	writeJSON(w, http.StatusNotFound, apiError{Error: "character not found"})
}

func handleCharactersPost(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid json"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "name is required"})
		return
	}
	if req.Level < 1 {
		req.Level = 1
	}

	c := buildCharacterFromRequest(req)

	idx := -1
	for i := range characters {
		if strings.EqualFold(characters[i].Name, c.Name) {
			idx = i
			break
		}
	}
	if idx >= 0 {
		characters[idx] = c
	} else {
		characters = append(characters, c)
	}

	EnrichCharacter(&c)

	saveCharacters()
	writeJSON(w, http.StatusCreated, c)
}

/**
*  apiCharactersHandler handles GET and POST requests for /api/characters
**/
func apiCharactersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleCharactersGet(w, r)
		return
	case http.MethodPost:
		handleCharactersPost(w, r)
		return
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}


func baseScoresFromReq(req createRequest) (AbilityScores, bool) {
	if req.AbilityScores == nil {
		return assignStandardArray(), false
	}
	s := *req.AbilityScores
	providedAll := s.Strength > 0 && s.Dexterity > 0 && s.Constitution > 0 &&
		s.Intelligence > 0 && s.Wisdom > 0 && s.Charisma > 0
	providedAny := s.Strength != 0 || s.Dexterity != 0 || s.Constitution != 0 ||
		s.Intelligence != 0 || s.Wisdom != 0 || s.Charisma != 0

	if providedAll {
		return s, true
	}
	nz10 := func(x int) int { if x == 0 { return 10 }; return x }
	return AbilityScores{
		Strength:     nz10(s.Strength),
		Dexterity:    nz10(s.Dexterity),
		Constitution: nz10(s.Constitution),
		Intelligence: nz10(s.Intelligence),
		Wisdom:       nz10(s.Wisdom),
		Charisma:     nz10(s.Charisma),
	}, providedAny
}


func applyRaceBonusesTo(base AbilityScores, race string) AbilityScores {
	rStr, rDex, rCon, rInt, rWis, rCha := raceBonusDeltas(race)
	return AbilityScores{
		Strength:     base.Strength + rStr,
		Dexterity:    base.Dexterity + rDex,
		Constitution: base.Constitution + rCon,
		Intelligence: base.Intelligence + rInt,
		Wisdom:       base.Wisdom + rWis,
		Charisma:     base.Charisma + rCha,
	}
}

func deriveSkillsFor(req createRequest, bg string) []string {
	if len(req.Skills) == 0 {
		return finalSkills(req.Class, bg, nil)
	}
	var skills []string
	for _, s := range req.Skills {
		if t := strings.TrimSpace(s); t != "" {
			skills = append(skills, normalizeSkill(t))
		}
	}
	return skills
}

func buildSpellcastingFor(class string, level int) *Spellcasting {
	if ct := casterType(class); ct != "none" {
		slots := spellSlotsFor(ct, level)
		maxL := maxSpellLevel(ct, level)
		spells := pickSpellsForClass(class, maxL, 4)
		return &Spellcasting{SlotsByLevel: slots, Spells: spells}
	}
	return nil
}


/**
*  buildCharacterFromRequest constructs a Character struct from an API request
**/
func buildCharacterFromRequest(req createRequest) Character {
	base, providedAny := baseScoresFromReq(req)
	final := base
	if !providedAny {
		final = applyRaceBonusesTo(base, req.Race)
	}

	bg := "acolyte"
	skills := deriveSkillsFor(req, bg)
	sc := buildSpellcastingFor(req.Class, req.Level)

	return Character{
		Name:             req.Name,
		Race:             strings.ToLower(strings.TrimSpace(req.Race)),
		Class:            strings.ToLower(strings.TrimSpace(req.Class)),
		Level:            req.Level,
		Background:       bg,
		AbilityScores:    final,
		ProficiencyBonus: profByLevel(req.Level),
		Skills:           skills,
		Spellcasting:     sc,
		Equipment: Equipment{
			Armor:   strings.TrimSpace(req.Armor),
			Weapon:  strings.TrimSpace(req.Weapon),
			Shield:  strings.TrimSpace(req.Shield),
			OffHand: strings.TrimSpace(req.OffHand),
		},
	}
}

/**
*  startServer configures routes and starts the HTTP server
**/
func startServer(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/characters", apiCharactersHandler)

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/", fileServer)

	log.Printf("listening on %s (open http://localhost%[1]s/characters.html or /charactersheet.html)\n", addr)
	return http.ListenAndServe(addr, mux)
}

/**
*  serveCommand parses CLI args and starts the HTTP server
**/
func serveCommand(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "listen address")
	_ = fs.Parse(args)
	if err := startServer(*addr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
