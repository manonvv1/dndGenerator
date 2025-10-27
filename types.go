// Layer: Domain (pure business models; no IO or framework deps)

package main

var StandardArray = []int{15, 14, 13, 12, 10, 8}

type AbilityScores struct {
	Strength     int
	Dexterity    int
	Constitution int
	Intelligence int
	Wisdom       int
	Charisma     int
}

type Character struct {
	Name             string
	Race             string
	Class            string
	Level            int
	AbilityScores    AbilityScores
	Background       string
	ProficiencyBonus int
	Equipment        Equipment
	Skills           []string
	Spellcasting     *Spellcasting
}

type Equipment struct {
	Armor   string
	Weapon  string
	Shield  string
	OffHand string

	WeaponInfo WeaponMeta
	ArmorInfo  ArmorMeta
}

type Spell struct {
	Name     string
	Level    int
	School   string
	Range    string
	Prepared bool
}

type WeaponMeta struct {
	Category    string
	RangeNormal int
	TwoHanded   bool
	DamageDice  string
	Finesse     bool
	WeaponRange string
}

type ArmorMeta struct {
	ArmorClass  int
	DexBonus    bool
	MaxDexBonus *int
}

type Spellcasting struct {
	SlotsByLevel map[int]int
	Spells       []Spell
}