package main

import "strings"

/**
*  profByLevel returns SRD proficiency bonus for the given level
**/
func profByLevel(lvl int) int {
	switch {
	case lvl >= 17:
		return 6
	case lvl >= 13:
		return 5
	case lvl >= 9:
		return 4
	case lvl >= 5:
		return 3
	default:
		return 2
	}
}

/**
*  abilityMod returns the SRD ability modifier for a score
**/
func abilityMod(score int) int {
    d := score - 10
    if d >= 0 {
        return d / 2
    }
    return (d - 1) / 2
}

/**
*   assignStandardArray returns the SRD standard array assigned in order
**/
func assignStandardArray() AbilityScores {
	arr := StandardArray
	return AbilityScores{
		Strength:     arr[0],
		Dexterity:    arr[1],
		Constitution: arr[2],
		Intelligence: arr[3],
		Wisdom:       arr[4],
		Charisma:     arr[5],
	}
}

/**
*  raceBonusDeltas gives the stat bonuses for a race
**/
func raceBonusDeltas(race string) (str, dex, con, intl, wis, cha int) {
	r := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(race), "-", " "))

	switch {
	case strings.Contains(r, "lightfoot halfling"):
		return 0, 2, 0, 0, 0, 1
	case strings.Contains(r, "stout halfling"):
		return 0, 2, 1, 0, 0, 0
	case strings.Contains(r, "halfling"):
		return 0, 2, 0, 0, 0, 0
	}

	switch r {
	case "human":
		return 1, 1, 1, 1, 1, 1
	case "hill dwarf":
		return 0, 0, 2, 0, 1, 0
	case "dwarf":
		return 0, 0, 2, 0, 0, 0
	case "elf":
		return 0, 2, 0, 0, 0, 0
	case "dragonborn":
		return 2, 0, 0, 0, 0, 1
	case "gnome":
		return 0, 0, 0, 2, 0, 0
	case "half orc":
		return 2, 0, 1, 0, 0, 0
	case "tiefling":
		return 0, 0, 0, 1, 0, 2
	default:
		return 0, 0, 0, 0, 0, 0
	}
}

/**
*  slugify makes a name safe to use in the API link.
**/
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	repl := []string{"’", "", "'", "", ",", "", "(", "", ")", "", ".", "", "–", "-", "—", "-"}
	for i := 0; i < len(repl); i += 2 {
		s = strings.ReplaceAll(s, repl[i], repl[i+1])
	}
	s = strings.Join(strings.Fields(s), " ")
	s = strings.ReplaceAll(s, " ", "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}

/**
*  findCharLike finds a character by case-insensitive exact or substring match
**/
func findCharLike(name string) *Character {
	lc := strings.ToLower(strings.TrimSpace(name))
	for i := range characters {
		if strings.ToLower(characters[i].Name) == lc || strings.Contains(strings.ToLower(characters[i].Name), lc) {
			return &characters[i]
		}
	}
	return nil
}

/**
*  abilityScoreByName returns a character's score by short name (str/dex/...)
**/
func abilityScoreByName(c *Character, name string) int {
	switch strings.ToLower(name) {
	case "strength", "str":
		return c.AbilityScores.Strength
	case "dexterity", "dex":
		return c.AbilityScores.Dexterity
	case "constitution", "con":
		return c.AbilityScores.Constitution
	case "intelligence", "int":
		return c.AbilityScores.Intelligence
	case "wisdom", "wis":
		return c.AbilityScores.Wisdom
	case "charisma", "cha":
		return c.AbilityScores.Charisma
	default:
		return 10
	}
}