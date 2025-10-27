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
	return filepath.Join("data", "5e-SRD-Spells.csv")
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

	if tryLoad(
		"5e-SRD-Spells.csv",
		"./5e-SRD-Spells.csv",
		filepath.Join("data", "5e-SRD-Spells.csv"),
		filepath.Join("Data", "5e-SRD-Spells.csv"),
		filepath.Join("DATA", "5e-SRD-Spells.csv"),
	) {
		return
	}

	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		_ = tryLoad(
			filepath.Join(dir, "5e-SRD-Spells.csv"),
			filepath.Join(dir, "data", "5e-SRD-Spells.csv"),
		)
	}
}

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

	hdr := rows[0]
	col := func(name string) int {
		name = strings.ToLower(strings.TrimSpace(name))
		for i, h := range hdr {
			if strings.ToLower(strings.TrimSpace(h)) == name {
				return i
			}
		}
		return -1
	}
	iName, iLevel, iClass := col("name"), col("level"), col("class")
	if iName < 0 || iLevel < 0 || iClass < 0 {
		return errors.New("spells CSV missing required headers: name, level, class")
	}

	tmpByClass := map[string][]Spell{}
	tmpLvlIdx := map[string]int{}

	for _, row := range rows[1:] {
		if len(row) <= iClass {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(row[iName]))
		if name == "" {
			continue
		}
		lvlStr := strings.TrimSpace(row[iLevel])
		if lvlStr == "" {
			continue
		}
		lvl, err := strconv.Atoi(lvlStr)
		if err != nil {
			continue
		}
		tmpLvlIdx[name] = lvl

		classes := strings.Split(row[iClass], ",")
		for _, c := range classes {
			cl := strings.ToLower(strings.TrimSpace(c))
			if cl == "" {
				continue
			}
			tmpByClass[cl] = append(tmpByClass[cl], Spell{
				Name:     name,
				Level:    lvl,
				Prepared: false,
			})
		}
	}

	for cl, list := range tmpByClass {
		sort.SliceStable(list, func(i, j int) bool { return list[i].Name < list[j].Name })
		uniq := list[:0]
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

	csvSpellsByClass = tmpByClass
	csvSpellLevelIndex = tmpLvlIdx
	return nil
}