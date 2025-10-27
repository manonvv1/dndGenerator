package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strings"
)

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


type apiEquipment struct {
	EquipmentCategory struct{ Name string `json:"name"` } `json:"equipment_category"`
	WeaponRange string `json:"weapon_range"`
	Range       struct{ Normal int `json:"normal"` } `json:"range"`
	Properties  []struct{ Index, Name string } `json:"properties"`
	ArmorClass  struct {
		Base int `json:"base"`
		DexBonus bool `json:"dex_bonus"`
		MaxBonus *int `json:"max_bonus"`
	} `json:"armor_class"`
}

type apiSpell struct {
	School struct{ Name string } `json:"school"`
	Range  string                `json:"range"`
}

/**
*  EnrichCharacter enriches a Character with weapon, armor, and spell data from the D&D  API
**/
func EnrichCharacter(c *Character) {
	if w := strings.TrimSpace(c.Equipment.Weapon); w != "" {
		var eq apiEquipment
		if err := httpGetJSON("https://www.dnd5eapi.co/api/equipment/"+slugify(w), &eq, nil); err == nil {
			c.Equipment.WeaponInfo.Category = eq.EquipmentCategory.Name
			c.Equipment.WeaponInfo.RangeNormal = eq.Range.Normal
			two := false
			for _, p := range eq.Properties {
				pi := strings.ToLower(p.Index + " " + p.Name)
				if strings.Contains(pi, "two-handed") { two = true; break }
			}
			c.Equipment.WeaponInfo.TwoHanded = two
		}
	}

	if a := strings.TrimSpace(c.Equipment.Armor); a != "" {
		var eq apiEquipment
		err := httpGetJSON("https://www.dnd5eapi.co/api/equipment/"+slugify(a), &eq, nil)
		if err != nil && !strings.Contains(strings.ToLower(a), "armor") {
			_ = httpGetJSON("https://www.dnd5eapi.co/api/equipment/"+slugify(a+" armor"), &eq, nil)
		}
		if eq.ArmorClass.Base != 0 || eq.ArmorClass.DexBonus || eq.ArmorClass.MaxBonus != nil {
			c.Equipment.ArmorInfo.ArmorClass  = eq.ArmorClass.Base
			c.Equipment.ArmorInfo.DexBonus    = eq.ArmorClass.DexBonus
			c.Equipment.ArmorInfo.MaxDexBonus = eq.ArmorClass.MaxBonus
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