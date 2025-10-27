package main

import "strings"

/**
* DexModOf returns the Dexterity ability modifier
**/
func dexModOf(c *Character) int { return abilityMod(c.AbilityScores.Dexterity) }

/**
* wisModOf returns the Wisdom ability modifier
**/
func wisModOf(c *Character) int { return abilityMod(c.AbilityScores.Wisdom) }


/**
* hasSkill reports whether the character is proficient in the given skill name
**/
func hasSkill(c *Character, name string) bool {
	n := normalizeSkill(name)
	for _, s := range c.Skills {
		if s == n {
			return true
		}
	}
	return false
}

/**
* armorBaseAndDexCap gives the base AC and Dex limit for an armor name
**/
func armorBaseAndDexCap(armorName string) (baseAC int, maxDex int, ok bool) {
	a := strings.ToLower(strings.TrimSpace(armorName))
	if strings.HasSuffix(a, " armor") {
		a = strings.TrimSuffix(a, " armor")
	}
	switch a {
	case "padded", "padded armor":
		return 11, -1, true
	case "leather", "leather armor":
		return 11, -1, true
	case "studded leather", "studded leather armor":
		return 12, -1, true

	case "hide", "hide armor":
		return 12, 2, true
	case "chain shirt":
		return 13, 2, true
	case "scale mail", "scale mail armor":
		return 14, 2, true
	case "breastplate":
		return 14, 2, true
	case "half plate":
		return 15, 2, true

	case "ring mail", "ring mail armor":
		return 14, 0, true
	case "chain mail", "chainmail":
		return 16, 0, true
	case "splint":
		return 17, 0, true
	case "plate", "plate armor":
		return 18, 0, true
	}
	return 0, 0, false
}

/**
* computeArmorClass works out Armor Class using the SRD rules and shield bonus
**/
func computeArmorClass(c *Character) int {
    armorName := normalizeEquipment(c.Equipment.Armor)
    shieldName := normalizeEquipment(c.Equipment.Shield)

    hasArmor := strings.TrimSpace(armorName) != ""
    hasShield := strings.TrimSpace(shieldName) != ""

    shBonus := 0
    if hasShield {
        shBonus = 2
    }

    cls := strings.ToLower(strings.TrimSpace(c.Class))

    if !hasArmor && !hasShield && cls == "monk" {
        return 10 + abilityMod(c.AbilityScores.Dexterity) + abilityMod(c.AbilityScores.Wisdom)
    }
    if !hasArmor && cls == "barbarian" {
        ac := 10 + abilityMod(c.AbilityScores.Dexterity) + abilityMod(c.AbilityScores.Constitution) + shBonus
        return ac
    }

    if hasArmor {
        if base, cap, ok := armorBaseAndDexCap(armorName); ok {
            dex := abilityMod(c.AbilityScores.Dexterity)
            if cap >= 0 && dex > cap { dex = cap }
            if cap == 0 { dex = 0 }
            return base + dex + shBonus
        }
        ac := 10 + abilityMod(c.AbilityScores.Dexterity) + shBonus
        return ac
    }

    ac := 10 + abilityMod(c.AbilityScores.Dexterity) + shBonus
    return ac
}


/**
* computeInitiativeBonus returns the initiative bonus
**/
func computeInitiativeBonus(c *Character) int { return dexModOf(c) }


/**
* computePassivePerception returns 10 + Wis mod
**/
func computePassivePerception(c *Character) int {
	pp := 10 + wisModOf(c)
	if hasSkill(c, "perception") {
		pp += c.ProficiencyBonus
	}
	return pp
}