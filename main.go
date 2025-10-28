// Layer: Infrastructure / UI (CLI commands; invokes application/domain)
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

/**
* usage prints command-line help
**/
func usage() {
	app := os.Args[0]
	fmt.Printf(`Usage:
  %s create -name NAME [-race RACE] [-class CLASS] [-level N] [-str N -dex N -con N -int N -wis N -cha N] [-background BG | -bg BG] [-skills "skill1, skill2"]
  %s view -name NAME_OR_SUBSTRING
  %s list
  %s delete -name NAME
  %s equip -name NAME [-weapon WEAPON] [-armor ARMOR] [-shield SHIELD] [-slot SLOT]
  %s prepare -name NAME -spell "SPELL NAME"
  %s learn -name NAME -spell "SPELL NAME"
  %s enrich [-limit N] [-dryrun] [-rps N] [-workers N] 
  %s inspect [-name NAME_OR_SUBSTRING]
  %s serve [-addr :8080]
`, app, app, app, app, app, app, app, app, app, app)
}

func upsertCharacter(c Character) {
    idx := -1
    for i := range characters {
        if strings.EqualFold(characters[i].Name, c.Name) {
            idx = i
            break
        }
    }
    if idx >= 0 {
        characters[idx] = c
    } else {
        characters = append(characters, c)
    }
}

func printAbilityScores(c *Character) {
	fmt.Println("Ability scores:")
	fmt.Printf("  STR: %d (%+d)\n", c.AbilityScores.Strength, abilityMod(c.AbilityScores.Strength))
	fmt.Printf("  DEX: %d (%+d)\n", c.AbilityScores.Dexterity, abilityMod(c.AbilityScores.Dexterity))
	fmt.Printf("  CON: %d (%+d)\n", c.AbilityScores.Constitution, abilityMod(c.AbilityScores.Constitution))
	fmt.Printf("  INT: %d (%+d)\n", c.AbilityScores.Intelligence, abilityMod(c.AbilityScores.Intelligence))
	fmt.Printf("  WIS: %d (%+d)\n", c.AbilityScores.Wisdom, abilityMod(c.AbilityScores.Wisdom))
	fmt.Printf("  CHA: %d (%+d)\n", c.AbilityScores.Charisma, abilityMod(c.AbilityScores.Charisma))
}

func normalizeSkillList(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		s = strings.ReplaceAll(s, "_", " ")
		s = strings.ReplaceAll(s, "-", " ")
		out[i] = strings.ToLower(strings.Join(strings.Fields(s), " "))
	}
	sort.Strings(out)
	return out
}

func printSpellSlotsBlock(c *Character, keys []int, includeHeader bool) {
	if includeHeader {
		fmt.Println("Spell slots:")
	}
	for _, lvl := range keys {
		fmt.Printf("  Level %d: %d\n", lvl, c.Spellcasting.SlotsByLevel[lvl])
	}
}

func slotKeys(c *Character, min int) []int {
	keys := make([]int, 0, len(c.Spellcasting.SlotsByLevel))
	for lvl := range c.Spellcasting.SlotsByLevel {
		if lvl >= min {
			keys = append(keys, lvl)
		}
	}
	sort.Ints(keys)
	return keys
}

func cmdCreate(args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	name := fs.String("name", "", "required")
	race := fs.String("race", "", "")
	class := fs.String("class", "", "")
	level := fs.Int("level", 1, "")
	bgLong := fs.String("background", "", "optional")
	bgShort := fs.String("bg", "", "optional")
	_ = bgLong
	_ = bgShort
	bg := "acolyte"
	str := fs.Int("str", 0, "")
	dex := fs.Int("dex", 0, "")
	con := fs.Int("con", 0, "")
	intl := fs.Int("int", 0, "")
	wis := fs.Int("wis", 0, "")
	cha := fs.Int("cha", 0, "")
	skillsFlag := fs.String("skills", "", "comma separated")
	_ = fs.Parse(args)

	if *name == "" {
		fmt.Println("name is required")
		os.Exit(2)
	}

	raw := []int{*str, *dex, *con, *intl, *wis, *cha}
	providedAll, providedAny := true, false
	for _, v := range raw {
		if v != 0 {
			providedAny = true
		} else {
			providedAll = false
		}
	}
	def10 := func(x int) int {
		if x == 0 {
			return 10
		}
		return x
	}
	var base AbilityScores
	switch {
	case providedAll:
		base = AbilityScores{*str, *dex, *con, *intl, *wis, *cha}
	case providedAny:
		base = AbilityScores{def10(*str), def10(*dex), def10(*con), def10(*intl), def10(*wis), def10(*cha)}
	default:
		base = assignStandardArray()
	}

	rStr, rDex, rCon, rInt, rWis, rCha := raceBonusDeltas(*race)
	final := AbilityScores{
		Strength: base.Strength + rStr, Dexterity: base.Dexterity + rDex, Constitution: base.Constitution + rCon,
		Intelligence: base.Intelligence + rInt, Wisdom: base.Wisdom + rWis, Charisma: base.Charisma + rCha,
	}

	if bg == "" {
		bg = strings.ToLower(strings.TrimSpace(*bgShort))
	}
	if bg == "" {
		bg = "acolyte"
	}
	var provided []string
	if *skillsFlag != "" {
		for _, p := range strings.Split(*skillsFlag, ",") {
			if t := strings.TrimSpace(p); t != "" {
				provided = append(provided, t)
			}
		}
	}

	ct := casterType(*class)
	var sc *Spellcasting
	if ct != "none" {
		slots := spellSlotsFor(ct, *level)
		maxL := maxSpellLevel(ct, *level)
		chosen := pickSpellsForClass(*class, maxL, 4)
		sc = &Spellcasting{SlotsByLevel: slots, Spells: chosen}
	}

	c := Character{
		Name: *name, Race: strings.ToLower(*race), Class: strings.ToLower(*class), Level: *level,
		Background: bg, AbilityScores: final, ProficiencyBonus: profByLevel(*level),
		Skills: finalSkills(*class, bg, provided), Spellcasting: sc,
	}
	upsertCharacter(c)
	saveCharacters()
	fmt.Printf("saved character %s\n", c.Name)
}

func cmdView(args []string) {
	fs := flag.NewFlagSet("view", flag.ExitOnError)
	name := fs.String("name", "", "required")
	noSlots := fs.Bool("no-slots", false, "hide spell slot lines")
	_ = fs.Parse(args)

	c := findCharLike(*name)
	if c == nil {
		fmt.Printf("character \"%s\" not found\n", *name)
		return
	}

	fmt.Printf("Name: %s\n", c.Name)
	fmt.Printf("Class: %s\n", strings.ToLower(c.Class))
	fmt.Printf("Race: %s\n", strings.ToLower(c.Race))
	fmt.Printf("Background: %s\n", strings.ToLower(strings.TrimSpace(c.Background)))
	fmt.Printf("Level: %d\n", c.Level)

	printAbilityScores(c)
	fmt.Printf("Proficiency bonus: %+d\n", c.ProficiencyBonus)

	skillsOut := normalizeSkillList(c.Skills)
	fmt.Printf("Skill proficiencies: %s\n", strings.Join(skillsOut, ", "))

	if strings.TrimSpace(c.Equipment.Weapon) != "" {
		fmt.Printf("Main hand: %s\n", c.Equipment.Weapon)
		if dmg := computeWeaponDamageString(c); dmg != "" {
			fmt.Printf("Weapon damage: %s\n", dmg)
		}
	}
	if strings.TrimSpace(c.Equipment.OffHand) != "" {
		fmt.Printf("Off hand: %s\n", c.Equipment.OffHand)
	}
	if strings.TrimSpace(c.Equipment.Armor) != "" {
		fmt.Printf("Armor: %s\n", c.Equipment.Armor)
	}
	if strings.TrimSpace(c.Equipment.Shield) != "" {
		fmt.Printf("Shield: %s\n", c.Equipment.Shield)
	}

	if c.Spellcasting != nil {
		switch casterType(c.Class) {
		case "half":
			if ck := cantripsKnown(c.Class, c.Level); ck > 0 {
				fmt.Println("Spell slots:")
				fmt.Printf("  Level 0: %d\n", ck)
			}
			keys := slotKeys(c, 1)
			if len(keys) > 0 {
				if cantripsKnown(c.Class, c.Level) == 0 {
					fmt.Println("Spell slots:")
				}
				printSpellSlotsBlock(c, keys, false)
			}
		case "warlock":
			if ck := cantripsKnown(c.Class, c.Level); ck > 0 {
				fmt.Println("Spell slots:")
				fmt.Printf("  Level 0: %d\n", ck)
			}
			keys := slotKeys(c, 0)
			if len(keys) > 0 {
				if cantripsKnown(c.Class, c.Level) == 0 {
					fmt.Println("Spell slots:")
				}
				printSpellSlotsBlock(c, keys, false)
			}
		case "full":
			if !*noSlots {
				fmt.Println("Spell slots:")
				if ck := cantripsKnown(c.Class, c.Level); ck > 0 {
					fmt.Printf("  Level 0: %d\n", ck)
				}
				keys := slotKeys(c, 1)
				printSpellSlotsBlock(c, keys, false)
			}
			sca := spellcastingAbilityForClass(c.Class)
			if sca != "" {
				abMod := abilityMod(abilityScoreByName(c, sca))
				saveDC := 8 + c.ProficiencyBonus + abMod
				attack := c.ProficiencyBonus + abMod
				fmt.Printf("Spellcasting ability: %s\n", sca)
				fmt.Printf("Spell save DC: %d\n", saveDC)
				fmt.Printf("Spell attack bonus: %+d\n", attack)
			}
		}
	}

	fmt.Printf("Armor class: %d\n", computeArmorClass(c))
	fmt.Printf("Initiative bonus: %d\n", computeInitiativeBonus(c))
	fmt.Printf("Passive perception: %d\n", computePassivePerception(c))
}

func cmdList() {
	for _, c := range characters {
		fmt.Printf("- %s (%s, level %d, %s)\n", c.Name, c.Class, c.Level, c.Race)
		fmt.Printf("Background: %s  ProficiencyBonus: %d\n", c.Background, c.ProficiencyBonus)
	}
}

func cmdDelete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	name := fs.String("name", "", "")
	_ = fs.Parse(args)
	for i := range characters {
		if characters[i].Name == *name {
			characters = append(characters[:i], characters[i+1:]...)
			saveCharacters()
			fmt.Printf("deleted %s\n", *name)
			return
		}
	}
	fmt.Printf("character \"%s\" not found\n", *name)
}

func cmdEquip(args []string) {
	fs := flag.NewFlagSet("equip", flag.ExitOnError)
	name := fs.String("name", "", "")
	weapon := fs.String("weapon", "", "")
	armor := fs.String("armor", "", "")
	shield := fs.String("shield", "", "")
	slot := fs.String("slot", "", "")
	_ = fs.Parse(args)

	c := findCharLike(*name)
	if c == nil {
		fmt.Printf("character \"%s\" not found\n", *name)
		return
	}

	s := strings.ToLower(strings.TrimSpace(*slot))
	if s == "offhand" {
		s = "off hand"
	}
	if s == "mainhand" {
		s = "main hand"
	}

	normalize := func(x string) string { return normalizeEquipment(x) }
	warnUnknown := func(x string) {
		if x == "" {
			return
		}
		if !isKnownEquipment(x) && equipmentCSVLoaded {
			fmt.Printf("(warning) \"%s\" not found in equipment CSV; continuing\n", x)
		}
	}

	if *weapon != "" {
		w := normalize(*weapon)
		warnUnknown(w)
		if s == "off hand" {
			if c.Equipment.OffHand != "" {
				fmt.Println("off hand already occupied")
				return
			}
			c.Equipment.OffHand = w
			fmt.Printf("Equipped weapon %s to off hand\n", w)
		} else {
			if c.Equipment.Weapon != "" {
				fmt.Println("main hand already occupied")
				return
			}
			c.Equipment.Weapon = w
			hand := "main hand"
			if s != "" && s != "main hand" {
				hand = s
			}
			fmt.Printf("Equipped weapon %s to %s\n", w, hand)
		}
	}
	if *armor != "" {
		a := normalize(*armor)
		warnUnknown(a)
		c.Equipment.Armor = a
		fmt.Printf("Equipped armor %s\n", a)
	}
	if *shield != "" {
		sh := normalize(*shield)
		warnUnknown(sh)
		c.Equipment.Shield = sh
		fmt.Printf("Equipped shield %s\n", sh)
	}
	saveCharacters()
}

func maxSlotLevel(slots map[int]int) int {
	max := 0
	for lvl := range slots {
		if lvl > max {
			max = lvl
		}
	}
	return max
}

func cmdPrepare(args []string) {
	fs := flag.NewFlagSet("prepare", flag.ExitOnError)
	name := fs.String("name", "", "required")
	spell := fs.String("spell", "", "required")
	_ = fs.Parse(args)
	if *spell != "" && len(fs.Args()) > 0 {
		*spell = *spell + " " + strings.Join(fs.Args(), " ")
	}
	if *name == "" || *spell == "" {
		usage()
		return
	}
	c := findCharLike(*name)
	if c == nil {
		fmt.Printf("character \"%s\" not found\n", *name)
		return
	}
	if casterType(c.Class) == "none" {
		fmt.Println("this class can't cast spells")
		return
	}
	if learnsSpells(c.Class) && !preparesSpells(c.Class) {
		fmt.Println("this class learns spells and can't prepare them")
		return
	}

	sp := strings.ToLower(strings.TrimSpace(*spell))
	slots := spellSlotsFor(casterType(c.Class), c.Level)
	if spellLvl, ok := spellLevelByName(sp); ok {
		if max := maxSlotLevel(slots); max == 0 || spellLvl > max {
			fmt.Println("the spell has higher level than the available spell slots")
			return
		}
	} else {
		fmt.Println("the spell has higher level than the available spell slots")
		return
	}

	if c.Spellcasting == nil {
		c.Spellcasting = &Spellcasting{SlotsByLevel: map[int]int{}, Spells: []Spell{}}
	}
	for i := range c.Spellcasting.Spells {
		if strings.ToLower(c.Spellcasting.Spells[i].Name) == sp {
			c.Spellcasting.Spells[i].Prepared = true
			saveCharacters()
			fmt.Printf("Prepared spell %s\n", sp)
			return
		}
	}
	lvl, _ := spellLevelByName(sp)
	c.Spellcasting.Spells = append(c.Spellcasting.Spells, Spell{Name: sp, Level: lvl, Prepared: true})
	saveCharacters()
	fmt.Printf("Prepared spell %s\n", sp)
}

func cmdLearn(args []string) {
	fs := flag.NewFlagSet("learn", flag.ExitOnError)
	name := fs.String("name", "", "required")
	spell := fs.String("spell", "", "required")
	_ = fs.Parse(args)

	if *spell != "" && len(fs.Args()) > 0 {
		*spell = *spell + " " + strings.Join(fs.Args(), " ")
	}
	if *name == "" || *spell == "" {
		usage()
		return
	}

	ch := findCharLike(*name)
	if ch == nil {
		fmt.Printf("character \"%s\" not found\n", *name)
		return
	}
	if casterType(ch.Class) == "none" {
		fmt.Println("this class can't cast spells")
		return
	}
	if preparesSpells(ch.Class) {
		fmt.Println("this class prepares spells and can't learn them")
		return
	}

	target := strings.ToLower(strings.TrimSpace(*spell))
	slots := spellSlotsFor(casterType(ch.Class), ch.Level)
	if spellLvl, ok := spellLevelByName(target); ok {
		if max := maxSlotLevel(slots); max == 0 || spellLvl > max {
			fmt.Println("the spell has higher level than the available spell slots")
			return
		}
	} else {
		fmt.Println("the spell has higher level than the available spell slots")
		return
	}

	if ch.Spellcasting == nil {
		ch.Spellcasting = &Spellcasting{SlotsByLevel: map[int]int{}, Spells: []Spell{}}
	}
	for _, s := range ch.Spellcasting.Spells {
		if strings.ToLower(s.Name) == target {
			saveCharacters()
			fmt.Printf("Learned spell %s\n", target)
			return
		}
	}
	lvl, _ := spellLevelByName(target)
	ch.Spellcasting.Spells = append(ch.Spellcasting.Spells, Spell{
		Name:     target,
		Level:    lvl,
		Prepared: false,
	})
	saveCharacters()
	fmt.Printf("Learned spell %s\n", target)
}

func cmdEnrich(args []string) {
	fs := flag.NewFlagSet("enrich", flag.ExitOnError)
	limit := fs.Int("limit", 0, "")
	_ = fs.Parse(args)

	processed := 0
	for i := range characters {
		if *limit > 0 && processed >= *limit {
			break
		}
		EnrichCharacter(&characters[i])
		processed++
	}
	saveCharacters()
	fmt.Println("enrichment done")
}

func cmdInspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	name := fs.String("name", "", "optional: name or substring")
	_ = fs.Parse(args)

	showChar := func(c *Character) {
		fmt.Printf("== %s ==\n", c.Name)
		if strings.TrimSpace(c.Equipment.Weapon) != "" {
			wi := c.Equipment.WeaponInfo
			fmt.Printf("Weapon: %s\n", c.Equipment.Weapon)
			if wi.Category != "" || wi.RangeNormal != 0 || wi.TwoHanded {
				fmt.Printf("  category=%s, range.normal=%d, two-handed=%t\n", wi.Category, wi.RangeNormal, wi.TwoHanded)
			} else {
				fmt.Println("  (no enriched weapon data)")
			}
		}
		if strings.TrimSpace(c.Equipment.Armor) != "" {
			ai := c.Equipment.ArmorInfo
			fmt.Printf("Armor: %s\n", c.Equipment.Armor)
			if ai.ArmorClass != 0 || ai.DexBonus || ai.MaxDexBonus != nil {
				if ai.MaxDexBonus != nil {
					fmt.Printf("  armor_class=%d, dex_bonus=%t, max_dex_bonus=%d\n", ai.ArmorClass, ai.DexBonus, *ai.MaxDexBonus)
				} else {
					fmt.Printf("  armor_class=%d, dex_bonus=%t\n", ai.ArmorClass, ai.DexBonus)
				}
			} else {
				fmt.Println("  (no enriched armor data)")
			}
		}
		if c.Spellcasting != nil && len(c.Spellcasting.Spells) > 0 {
			any := false
			for _, s := range c.Spellcasting.Spells {
				if s.School != "" || s.Range != "" {
					if !any {
						fmt.Println("Spells (enriched):")
						any = true
					}
					fmt.Printf("  - %s: school=%s, range=%s\n", s.Name, s.School, s.Range)
				}
			}
			if !any {
				fmt.Println("Spells: (no enriched spell data)")
			}
		}
		fmt.Println()
	}

	if strings.TrimSpace(*name) == "" {
		for i := range characters {
			showChar(&characters[i])
		}
		return
	}
	c := findCharLike(*name)
	if c == nil {
		fmt.Printf("character \"%s\" not found\n", *name)
		return
	}
	showChar(c)
}

/**
* main is the CLI entrypoint for creating, viewing, managing, and enriching characters
**/
func main() {
	loadCharacters()
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "serve":
		serveCommand(os.Args[2:])
	case "create":
		cmdCreate(os.Args[2:])
	case "view":
		cmdView(os.Args[2:])
	case "list":
		cmdList()
	case "delete":
		cmdDelete(os.Args[2:])
	case "equip":
		cmdEquip(os.Args[2:])
	case "prepare", "prepare-spell":
		cmdPrepare(os.Args[2:])
	case "learn", "learn-spell":
		cmdLearn(os.Args[2:])
	case "enrich":
		cmdEnrich(os.Args[2:])
	case "inspect":
		cmdInspect(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}
