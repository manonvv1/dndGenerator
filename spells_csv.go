// Layer: Infrastructure (data source adapter: load spells from CSV)

package main

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const defaultSpellsFile = "5e-SRD-Spells.csv"

var (
	csvSpellsByClass   = map[string][]Spell{}
	csvSpellLevelIndex = map[string]int{}
	csvLoaded          = false
)

/**
*  spellsCSVPath returns the spells CSV path
**/
func spellsCSVPath() string {
	if p := strings.TrimSpace(os.Getenv("SPELLS_CSV")); p != "" {
		return p
	}
	return filepath.Join("data", defaultSpellsFile)
}

/**
*  tryLoadSpells attempts to load the CSV
**/
func tryLoad(paths ...string) bool {
	for _, p := range paths {
		if err := loadSpellsFromCSV(p); err == nil {
			csvLoaded = true
			return true
		}
	}
	return false
}

func init() {
	if p := strings.TrimSpace(os.Getenv("SPELLS_CSV")); p != "" && tryLoad(p) {
		return
	}

	commonPaths := []string{
		defaultSpellsFile,
		"./" + defaultSpellsFile,
		filepath.Join("data", defaultSpellsFile),
		filepath.Join("Data", defaultSpellsFile),
		filepath.Join("DATA", defaultSpellsFile),
	}
	if tryLoad(commonPaths...) {
		return
	}

	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		_ = tryLoad(
			filepath.Join(dir, defaultSpellsFile),
			filepath.Join(dir, "data", defaultSpellsFile),
		)
	}
}

// --- Helpers to reduce cognitive complexity (behavior unchanged) ---

func trimLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

type spellHeaderIdx struct {
	iName  int
	iLevel int
	iClass int
}

func resolveHeaderIndexes(hdr []string) (spellHeaderIdx, error) {
	colIdx := func(name string) int {
		n := trimLower(name)
		for i, h := range hdr {
			if trimLower(h) == n {
				return i
			}
		}
		return -1
	}
	idx := spellHeaderIdx{
		iName:  colIdx("name"),
		iLevel: colIdx("level"),
		iClass: colIdx("class"),
	}
	if idx.iName < 0 || idx.iLevel < 0 || idx.iClass < 0 {
		return spellHeaderIdx{}, errors.New("spells CSV missing required headers: name, level, class")
	}
	return idx, nil
}

func parseSpellRow(row []string, idx spellHeaderIdx) (name string, level int, classes []string, ok bool) {
	if len(row) <= idx.iClass {
		return "", 0, nil, false
	}
	name = trimLower(row[idx.iName])
	if name == "" {
		return "", 0, nil, false
	}
	lvlStr := strings.TrimSpace(row[idx.iLevel])
	lvl, err := strconv.Atoi(lvlStr)
	if err != nil {
		return "", 0, nil, false
	}
	rawClasses := strings.Split(row[idx.iClass], ",")
	classes = make([]string, 0, len(rawClasses))
	for _, c := range rawClasses {
		if cl := trimLower(c); cl != "" {
			classes = append(classes, cl)
		}
	}
	if len(classes) == 0 {
		return "", 0, nil, false
	}
	return name, lvl, classes, true
}

func addToTempIndexes(name string, lvl int, classes []string, tmpByClass map[string][]Spell, tmpLvlIdx map[string]int) {
	tmpLvlIdx[name] = lvl
	for _, cl := range classes {
		tmpByClass[cl] = append(tmpByClass[cl], Spell{
			Name:     name,
			Level:    lvl,
			Prepared: false,
		})
	}
}

func sortAndDedupePerClass(tmpByClass map[string][]Spell) {
	for cl, list := range tmpByClass {
		sort.SliceStable(list, func(i, j int) bool { return list[i].Name < list[j].Name })
		uniq := make([]Spell, 0, len(list))
		var last string
		for _, s := range list {
			if s.Name == last {
				continue
			}
			uniq = append(uniq, s)
			last = s.Name
		}
		tmpByClass[cl] = uniq
	}
}

// --- Refactored function (same behavior, lower complexity) ---

/**
*  loadSpellsFromCSV parses the SRD spells CSV into per-class lists and a name level index
**/
func loadSpellsFromCSV(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true

	rows, err := r.ReadAll()
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return errors.New("spells CSV is empty")
	}

	idx, err := resolveHeaderIndexes(rows[0])
	if err != nil {
		return err
	}

	tmpByClass := map[string][]Spell{}
	tmpLvlIdx := map[string]int{}

	for _, row := range rows[1:] {
		name, lvl, classes, ok := parseSpellRow(row, idx)
		if !ok {
			continue
		}
		addToTempIndexes(name, lvl, classes, tmpByClass, tmpLvlIdx)
	}

	sortAndDedupePerClass(tmpByClass)

	csvSpellsByClass = tmpByClass
	csvSpellLevelIndex = tmpLvlIdx
	return nil
}

