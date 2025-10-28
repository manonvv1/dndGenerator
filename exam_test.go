package main

import "testing"

func mkChar(str, dex int, dice, wrange string, finesse bool) *Character {
	return &Character{
		AbilityScores: AbilityScores{Strength: str, Dexterity: dex},
		Equipment: Equipment{
			Weapon: "x",
			WeaponInfo: WeaponMeta{
				DamageDice:  dice,
				WeaponRange: wrange, 
				Finesse:     finesse,
			},
		},
	}
}

func TestAbilityMod(t *testing.T) {
	cases := []struct {
		score int
		want  int
	}{
		{10, 0}, {11, 0}, {9, -1}, {15, 2}, {17, 3},
	}
	for _, tc := range cases {
		if got := abilityMod(tc.score); got != tc.want {
			t.Fatalf("abilityMod(%d)=%d; want %d", tc.score, got, tc.want)
		}
	}
}

func TestDamage_Gor(t *testing.T) {
	c := mkChar(17, 14, "1d12", "Melee", false) 
	if got, want := computeWeaponDamageString(c), "1d12 + 3"; got != want {
		t.Fatalf("Gor dmg = %q; want %q", got, want)
	}
}

func TestDamage_Nyx(t *testing.T) {
	c := mkChar(13, 15, "1d6", "Melee", true) 
	if got, want := computeWeaponDamageString(c), "1d6 + 2"; got != want {
		t.Fatalf("Nyx dmg = %q; want %q", got, want)
	}
}

func TestDamage_Bruni(t *testing.T) {
	c := mkChar(8, 15, "1d8", "Melee", true) 
	if got, want := computeWeaponDamageString(c), "1d8 + 2"; got != want {
		t.Fatalf("Bruni dmg = %q; want %q", got, want)
	}
}
