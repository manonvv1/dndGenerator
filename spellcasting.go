package main

import (
	"sort"
	"strings"
)

/**
*  casterType returns "full", "half", "warlock", or "none" for a class name
**/
func casterType(class string) string {
	switch strings.ToLower(strings.TrimSpace(class)) {
	case "wizard", "cleric", "druid", "bard", "sorcerer":
		return "full"
	case "paladin", "ranger":
		return "half"
	case "warlock":
		return "warlock"
	default:
		return "none"
	}
}


var fullSlots = [21][10]int{
	{},
	{0, 2, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 3, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 4, 2, 0, 0, 0, 0, 0, 0, 0},
	{0, 4, 3, 0, 0, 0, 0, 0, 0, 0},
	{0, 4, 3, 2, 0, 0, 0, 0, 0, 0},
	{0, 4, 3, 3, 0, 0, 0, 0, 0, 0},
	{0, 4, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 4, 3, 3, 2, 0, 0, 0, 0, 0},
	{0, 4, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 4, 3, 3, 3, 2, 0, 0, 0, 0},
	{0, 4, 3, 3, 3, 2, 1, 0, 0, 0},
	{0, 4, 3, 3, 3, 2, 1, 0, 0, 0},
	{0, 4, 3, 3, 3, 2, 1, 1, 0, 0},
	{0, 4, 3, 3, 3, 2, 1, 1, 0, 0},
	{0, 4, 3, 3, 3, 2, 1, 1, 1, 0},
	{0, 4, 3, 3, 3, 2, 1, 1, 1, 0},
	{0, 4, 3, 3, 3, 2, 1, 1, 1, 1},
	{0, 4, 3, 3, 3, 3, 1, 1, 1, 1},
	{0, 4, 3, 3, 3, 3, 2, 1, 1, 1},
	{0, 4, 3, 3, 3, 3, 2, 2, 1, 1},
}

/**
*  halfCasterSlots gives the spell slot table for half-casters (like paladins and rangers)
* It uses half of the level (rounded down) to pick the right number of slots.
**/
func halfCasterSlots(level int) map[int]int {
	eff := level / 2 
	if eff <= 0 {
		return map[int]int{}
	}
	if eff > 20 {
		eff = 20
	}
	out := map[int]int{}
	row := fullSlots[eff]
	for sl := 1; sl <= 9; sl++ {
		if row[sl] > 0 {
			out[sl] = row[sl]
		}
	}
	return out
}

type warlockProgRow struct{ slotLevel, slots int }

var warlockTable = [21]warlockProgRow{
	{0, 0},
	{1, 1},
	{1, 2},
	{2, 2},
	{2, 2},
	{3, 2},
	{3, 2},
	{4, 2},
	{4, 2},
	{5, 2},
	{5, 2},
	{5, 3},
	{5, 3},
	{5, 3},
	{5, 3},
	{5, 3},
	{5, 3},
	{5, 4},
	{5, 4},
	{5, 4},
	{5, 4},
}

/**
*  spellSlotsFor returns a level→count map for a caster type at a given leve
**/
func spellSlotsFor(caster string, level int) map[int]int {
	if level < 1 {
		return map[int]int{}
	}
	switch caster {
	case "full":
		row := fullSlots[level]
		out := map[int]int{}
		for sl := 1; sl <= 9; sl++ {
			if row[sl] > 0 {
				out[sl] = row[sl]
			}
		}
		return out
	case "half":
		return halfCasterSlots(level)
	case "warlock":
		w := warlockTable[level]
		out := map[int]int{}
		if w.slotLevel > 0 && w.slots > 0 {
			out[w.slotLevel] = w.slots
		}
		return out
	default:
		return map[int]int{}
	}
}

/**
*  maxSpellLevel returns the highest spell level available for the caster type at level
**/
func maxSpellLevel(caster string, level int) int {
	switch caster {
	case "full":
		row := fullSlots[level]
		for sl := 9; sl >= 1; sl-- {
			if row[sl] > 0 {
				return sl
			}
		}
		return 0
	case "half":
		slots := halfCasterSlots(level)
		max := 0
		for sl := range slots {
			if sl > max {
				max = sl
			}
		}
		return max
	case "warlock":
		return warlockTable[level].slotLevel
	default:
		return 0
	}
}

/**
*  pickSpellsForClass chooses up to count spells ≤ maxLevel for a class
**/
func pickSpellsForClass(class string, maxLevel int, count int) []Spell {
	cl := strings.ToLower(strings.TrimSpace(class))
	pool := csvSpellsByClass[cl]

	filtered := make([]Spell, 0, len(pool))
	for _, s := range pool {
		if s.Level <= maxLevel {
			filtered = append(filtered, s)
		}
	}

	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].Name < filtered[j].Name })

	if count > len(filtered) {
		count = len(filtered)
	}
	out := append([]Spell(nil), filtered[:count]...)
	for i := range out {
		out[i].Prepared = true
	}
	return out
}

/**
*  spellLevelByName returns a spell's level by name
**/
func spellLevelByName(name string) (int, bool) {
	n := strings.ToLower(strings.TrimSpace(name))
	lvl, ok := csvSpellLevelIndex[n]
	return lvl, ok
}