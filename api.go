// ===== Infrastructure Adapter (HTTP client for DnD 2014 API) =====

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strings"
)

const baseEquipmentURL = "https://www.dnd5eapi.co/api/equipment/"

type EquipmentProvider interface {
	FetchWeaponMeta(name string) (WeaponMeta, bool)
	FetchArmorMeta(name string) (ArmorMeta, bool)
}

var EquipProv EquipmentProvider = &HttpEquipmentAdapter{}

/**
*  httpGetJSON performs an HTTP GET and decodes JSON response into v
**/
func httpGetJSON(url string, v any, tick <-chan time.Time) error {
	if tick != nil {
		<-tick
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

type HttpEquipmentAdapter struct{}

func (a *HttpEquipmentAdapter) FetchWeaponMeta(name string) (WeaponMeta, bool) {
	var eq apiEquipment
	if err := httpGetJSON(baseEquipmentURL+slugify(name), &eq, nil); err != nil {
		return WeaponMeta{}, false
	}
	wm := WeaponMeta{
		Category:    eq.EquipmentCategory.Name,
		RangeNormal: eq.Range.Normal,
		DamageDice:  eq.Damage.DamageDice,
		WeaponRange: eq.WeaponRange,
		TwoHanded:   false,
		Finesse:     false,
	}
	for _, p := range eq.Properties {
		switch strings.ToLower(p.Index) {
		case "two-handed":
			wm.TwoHanded = true
		case "finesse":
			wm.Finesse = true
		}
	}
	if wm.RangeNormal == 0 && strings.EqualFold(wm.WeaponRange, "melee") {
		wm.RangeNormal = 5
	}
	return wm, true
}

func (a *HttpEquipmentAdapter) FetchArmorMeta(name string) (ArmorMeta, bool) {
	var eq apiEquipment
	if err := httpGetJSON(baseEquipmentURL+slugify(name), &eq, nil); err != nil {
		_ = httpGetJSON(baseEquipmentURL+slugify(name+" armor"), &eq, nil)
	}
	if eq.ArmorClass.Base == 0 && !eq.ArmorClass.DexBonus && eq.ArmorClass.MaxBonus == nil {
		return ArmorMeta{}, false
	}
	return ArmorMeta{
		ArmorClass:  eq.ArmorClass.Base,
		DexBonus:    eq.ArmorClass.DexBonus,
		MaxDexBonus: eq.ArmorClass.MaxBonus,
	}, true
}


type apiEquipment struct {
	EquipmentCategory struct{ Name string `json:"name"` } `json:"equipment_category"`
	WeaponRange string `json:"weapon_range"`
	Range       struct{ Normal int `json:"normal"` } `json:"range"`
	Properties  []struct{
		Index string `json:"index"`
		Name  string `json:"name"`
	} `json:"properties"`
	Damage struct {
		DamageDice string `json:"damage_dice"`
	} `json:"damage"`
	ArmorClass  struct {
		Base     int  `json:"base"`
		DexBonus bool `json:"dex_bonus"`
		MaxBonus *int `json:"max_bonus"`
	} `json:"armor_class"`
}

type apiSpell struct {
	School struct{ Name string } `json:"school"`
	Range  string                `json:"range"`
}

/**
*  EnrichCharacter enriches a Character with weapon, armor, and spell data from the D&D API
**/
func EnrichCharacter(c *Character) {
	if w := strings.TrimSpace(c.Equipment.Weapon); w != "" {
		if wm, ok := EquipProv.FetchWeaponMeta(w); ok {
			c.Equipment.WeaponInfo = wm
		}
	}

	if a := strings.TrimSpace(c.Equipment.Armor); a != "" {
		if am, ok := EquipProv.FetchArmorMeta(a); ok {
			c.Equipment.ArmorInfo = am
		}
	}

	if c.Spellcasting != nil {
		for i := range c.Spellcasting.Spells {
			name := c.Spellcasting.Spells[i].Name
			var sp apiSpell
			if err := httpGetJSON("https://www.dnd5eapi.co/api/spells/"+slugify(name), &sp, nil); err == nil {
				c.Spellcasting.Spells[i].School = sp.School.Name
				c.Spellcasting.Spells[i].Range  = sp.Range
			}
		}
	}
}
