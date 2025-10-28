// Layer: Infrastructure (data source adapter: equipment CSV helpers)

package main

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultEquipmentCSV = "5e-SRD-Equipment.csv"
	armorSuffix         = " armor"
)

var (
	csvEquipmentTypeByName = map[string]string{}
	equipmentCSVLoaded     = false
)

/**
*  equipmentCSVPath returns the equipment CSV path
**/
func equipmentCSVPath() string {
	if p := strings.TrimSpace(os.Getenv("EQUIPMENT_CSV")); p != "" {
		return p
	}
	return filepath.Join("data", defaultEquipmentCSV)
}

/**
*  tryLoadEquipment attempts to load the CSV
**/
func tryLoadEquipment(paths ...string) bool {
	for _, p := range paths {
		if err := loadEquipmentFromCSV(p); err == nil {
			equipmentCSVLoaded = true
			return true
		}
	}
	return false
}

func init() {
	if p := strings.TrimSpace(os.Getenv("EQUIPMENT_CSV")); p != "" && tryLoadEquipment(p) {
		return
	}
	if tryLoadEquipment(
		defaultEquipmentCSV,
		"./"+defaultEquipmentCSV,
		filepath.Join("data", defaultEquipmentCSV),
		filepath.Join("Data", defaultEquipmentCSV),
		filepath.Join("DATA", defaultEquipmentCSV),
	) {
		return
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		_ = tryLoadEquipment(
			filepath.Join(dir, defaultEquipmentCSV),
			filepath.Join(dir, "data", defaultEquipmentCSV),
		)
	}
}

func readEquipmentCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true
	return r.ReadAll()
}

func findColumnIndex(header []string, name string) int {
	want := strings.ToLower(strings.TrimSpace(name))
	for i, h := range header {
		if strings.ToLower(strings.TrimSpace(h)) == want {
			return i
		}
	}
	return -1
}

func buildNameTypeIndex(rows [][]string, iName, iType int) map[string]string {
	tmp := make(map[string]string, len(rows))
	for _, row := range rows[1:] {
		if len(row) <= iType {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(row[iName]))
		typ := strings.ToLower(strings.TrimSpace(row[iType]))
		if name == "" || typ == "" {
			continue
		}
		tmp[name] = typ
	}
	return tmp
}

/**
*  loadEquipmentFromCSV builds a name type index from the SRD equipment CSV
**/
func loadEquipmentFromCSV(path string) error {
	rows, err := readEquipmentCSV(path)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return errors.New("equipment CSV is empty")
	}

	hdr := rows[0]
	iName := findColumnIndex(hdr, "name")
	iType := findColumnIndex(hdr, "type")
	if iName < 0 || iType < 0 {
		return errors.New("equipment CSV missing required headers: name, type")
	}

	csvEquipmentTypeByName = buildNameTypeIndex(rows, iName, iType)
	return nil
}

/**
*  equipmentType returns the equipment type for a name
**/
func equipmentType(name string) (string, bool) {
	t, ok := csvEquipmentTypeByName[strings.ToLower(strings.TrimSpace(name))]
	return t, ok
}

/**
*  normalizeEquipment makes the equipment same lowercase and spacing
**/
func normalizeEquipment(name string) string {
	n := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))

	if _, ok := csvEquipmentTypeByName[n]; ok {
		return n
	}
	if !strings.Contains(n, "armor") {
		if _, ok := csvEquipmentTypeByName[n+armorSuffix]; ok {
			return n + armorSuffix
		}
	}
	if strings.Contains(n, armorSuffix) {
		n2 := strings.ReplaceAll(n, armorSuffix, "")
		if _, ok := csvEquipmentTypeByName[n2]; ok {
			return n2
		}
	}
	return n
}

/**
*  isKnownEquipment reports whether the name exists in the CSV index
**/
func isKnownEquipment(name string) bool {
	_, ok := equipmentType(name)
	return ok
}
