package main

import (
	"sort"
	"strings"
)

var classSkillOptions = map[string][]string{
	"barbarian": {"Animal Handling", "Athletics", "Intimidation", "Nature", "Perception", "Survival"},
	"fighter":   {"Acrobatics", "Animal Handling", "Athletics", "History", "Insight", "Intimidation", "Perception", "Survival"},
	"cleric":    {"History", "Insight", "Medicine", "Persuasion", "Religion"},
	"wizard":    {"Arcana", "History", "Insight", "Investigation", "Medicine", "Religion"},
	"rogue":     {"Acrobatics", "Athletics", "Deception", "Insight", "Intimidation", "Investigation", "Perception", "Performance", "Persuasion", "Sleight of Hand", "Stealth"},
	"ranger":    {"Animal Handling", "Athletics", "Insight", "Investigation", "Nature", "Perception", "Stealth", "Survival"},
	"paladin":   {"Athletics", "Insight", "Intimidation", "Medicine", "Persuasion", "Religion"},
	"warlock":   {"Arcana", "Deception", "History", "Intimidation", "Investigation", "Nature", "Religion"},
	"monk":      {"Acrobatics", "Athletics", "History", "Insight", "Religion", "Stealth"},
}

var classSkillCount = map[string]int{
	"barbarian": 2, "fighter": 2, "cleric": 2, "wizard": 2,
	"rogue": 4, "ranger": 3, "paladin": 2, "warlock": 2, "monk": 2,
}

var backgroundSkills = map[string][]string{
    "acolyte": {"Insight", "Religion"},
}

/**
*  defaultClassSkills returns default skill proficiencies for a class
**/
func defaultClassSkills(className string) []string {
	low := strings.ToLower(className)
	opts := append([]string(nil), classSkillOptions[low]...) 
	sort.Strings(opts)
	n := classSkillCount[low]
	if n <= 0 { n = 2 }
	if len(opts) > n { opts = opts[:n] }
	return opts
}

/**
*  defaultBackgroundSkills returns skills granted by a background
**/
func defaultBackgroundSkills(bg string) []string {
	return append([]string{}, backgroundSkills[strings.ToLower(bg)]...)
}

/**
*  normalizeSkill lowercases, trims for a skill name and replaces underscores with spaces
**/
func normalizeSkill(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), "_", " "))), " ")
}

/**
*  finalSkills returns the final normalized, sorted skills from provided/class/bg
**/
func finalSkills(className, bg string, provided []string) []string {
	var out []string
	if len(provided) > 0 {
		for _, p := range provided {
			if t := strings.TrimSpace(p); t != "" {
				out = append(out, normalizeSkill(t))
			}
		}
	} else {
		for _, s := range defaultClassSkills(className) {
			out = append(out, normalizeSkill(s))
		}
	}
	if bg != "" {
		for _, s := range defaultBackgroundSkills(bg) {
			out = append(out, normalizeSkill(s))
		}
	}
	sort.Strings(out)
	return out
}