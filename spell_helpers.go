// Layer: Domain (helpers for spell logic; no IO)

package main

import "strings"

/**
*  learnsSpells reports whether the class learns spells
**/
func learnsSpells(class string) bool {
	switch strings.ToLower(strings.TrimSpace(class)) {
	case "bard", "sorcerer", "warlock", "ranger", "wizard":
		return true
	default:
		return false
	}
}

/**
*  preparesSpells reports whether the class prepares spells
**/
func preparesSpells(class string) bool {
	switch strings.ToLower(strings.TrimSpace(class)) {
	case "cleric", "druid", "paladin", "wizard":
		return true
	default:
		return false
	}
}

/**
*  spellcastingAbilityForClass returns the casting ability for the class
**/
func spellcastingAbilityForClass(class string) string {
	switch strings.ToLower(strings.TrimSpace(class)) {
	case "wizard":
		return "intelligence"
	case "cleric", "druid", "ranger":
		return "wisdom"
	case "bard", "sorcerer", "warlock", "paladin":
		return "charisma"
	default:
		return ""
	}
}

/**
*  cantripsKnown returns the number of cantrips known for class at level
**/
func cantripsKnown(class string, level int) int {
	l := strings.ToLower(strings.TrimSpace(class))
	if level < 1 {
		return 0
	}

	switch l {
	case "warlock", "bard":
		switch {
		case level >= 10:
			return 4
		case level >= 4:
			return 3
		default:
			return 2
		}
	case "wizard":
		switch {
		case level >= 10:
			return 5
		case level >= 4:
			return 4
		default:
			return 3
		}
	case "cleric", "druid":
		switch {
		case level >= 10:
			return 5
		case level >= 4:
			return 4
		default:
			return 3
		}
	case "sorcerer":
		switch {
		case level >= 10:
			return 6
		case level >= 4:
			return 5
		default:
			return 4
		}
	default:
		return 0
	}
}