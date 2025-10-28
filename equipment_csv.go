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

/**
*  loadEquipmentFromCSV builds a name type index from the SRD equipment CSV
**/
func loadEquipmentFromCSV(path string) error {
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
		return errors.New("equipment CSV is empty")
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
	iName, iType := col("name"), col("type")
	if iName < 0 || iType < 0 {
		return errors.New("equipment CSV missing required headers: name, type")
	}

	tmp := map[string]string{}
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
	csvEquipmentTypeByName = tmp
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
