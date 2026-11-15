package domain_test

// **Wie aus MSE-Feldern eine Karte wird** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Wogegen hier geprüft wird
//
// **Gegen die Feldnamen aus `docs/FORMATE.md`** (Zeile 1832 ff.), erhoben am
// `M15-Magic-Pack`: `casting cost` mit seinen vier Alternativnamen,
// `super type` und `sub type` als zwei Felder, `pt` als geteiltes drittes.
//
// **Und gegen die Grenze des Spiels**: die Feldnamen gehören dem Paket
// `magic`; ein fremdes Spiel bekommt keine erfundenen Eigenschaften (§15).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine MSE-Karte wird zu Eigenschaften** — die Feldnamen stehen in
// `docs/FORMATE.md` (Zeile 1832 ff.), erhoben am `M15-Magic-Pack`.
func TestEineMSEKarteWirdZuEigenschaften(t *testing.T) {
	t.Parallel()

	eigen := domain.EigenschaftenAus("magic", map[string]string{
		"name":         "Grizzlybär",
		"casting cost": "{1}{G}",
		"super type":   "Creature",
		"sub type":     "Bear",
		"power":        "2",
		"toughness":    "2",
	})

	switch {
	case eigen.Name != "Grizzlybär":
		t.Errorf("Name = %q", eigen.Name)
	case eigen.Manakosten.Manawert() != 2:
		t.Errorf("Manawert = %d", eigen.Manakosten.Manawert())
	case !eigen.Typzeile.Ist(domain.TypKreatur):
		t.Error("keine Kreatur")
	case eigen.Staerke.Zahl != 2 || !eigen.Widerstandskraft.Steht:
		t.Errorf("Werte = %+v / %+v", eigen.Staerke, eigen.Widerstandskraft)
	}
}

// **Die Alternativnamen gelten auch** — `docs/FORMATE.md` führt für die
// Manakosten vier davon, und gewachsene Pakete benutzen sie.
func TestDieAlternativnamenDerFelderGelten(t *testing.T) {
	t.Parallel()

	for _, feld := range []string{
		"casting cost", "mana cost", "manacost", "cost", "cc",
	} {
		eigen := domain.EigenschaftenAus("magic", map[string]string{feld: "{R}"})

		farben := eigen.Manakosten.Farben()
		if len(farben) != 1 || farben[0] != domain.FarbeRot {
			t.Errorf("%q: Farben = %v", feld, farben)
		}
	}
}

// **`pt` trägt beide Werte in einem Feld** — MSE führt es neben `power` und
// `toughness`, und manche Vorlagen füllen nur dieses.
func TestDasGeteilteFeldTraegtBeideWerte(t *testing.T) {
	t.Parallel()

	eigen := domain.EigenschaftenAus("magic", map[string]string{"pt": "3/4"})

	if eigen.Staerke.Zahl != 3 || eigen.Widerstandskraft.Zahl != 4 {
		t.Fatalf("Werte = %+v / %+v", eigen.Staerke, eigen.Widerstandskraft)
	}
}

// **Ein fremdes Spiel bekommt keine erfundenen Eigenschaften.**
//
// Die Feldnamen gehören dem Spielpaket `magic`; ein anderes Spiel nennt seine
// Felder, wie es will, und `type` heißt dort vielleicht etwas ganz anderes.
func TestEinFremdesSpielBekommtKeineEigenschaften(t *testing.T) {
	t.Parallel()

	eigen := domain.EigenschaftenAus("space", map[string]string{
		"name": "Sonde", "power": "9",
	})

	switch {
	case eigen.Name != "Sonde":
		t.Errorf("Name = %q — der Name gilt in jedem Spiel", eigen.Name)
	case eigen.Staerke.Steht:
		t.Error("ein fremdes Spiel hat Stärke bekommen")
	}
}

// **Ein Planeswalker trägt seine Loyalität** — CR 306.5b: die gedruckte Zahl
// sagt, mit wie vielen Marken er ins Spiel kommt.
func TestEinPlaneswalkerTraegtSeineLoyalitaet(t *testing.T) {
	t.Parallel()

	eigen := domain.EigenschaftenAus("magic", map[string]string{
		"super type": "Legendary Planeswalker",
		"sub type":   "Jace",
		"loyalty":    "3",
	})

	switch {
	case !eigen.Typzeile.Ist(domain.TypPlaneswalker):
		t.Error("kein Planeswalker")
	case !eigen.Loyalitaet.Steht || eigen.Loyalitaet.Zahl != 3:
		t.Errorf("Loyalität = %+v", eigen.Loyalitaet)
	}
}

// **Eine Kampfkarte trägt ihre Verteidigung** — CR 310.1 mit CR 210.1.
//
// **Der Feldname ist eine Annahme**, und der Grund steht am Feld: das
// erhobene Magic-Paket führt ihn nicht. Der Lauf sagt damit, was geschieht,
// **wenn** ein Paket ihn führt — und nichts darüber, ob eines es tut.
func TestEineKampfkarteTraegtIhreVerteidigung(t *testing.T) {
	t.Parallel()

	eigen := domain.EigenschaftenAus("magic", map[string]string{
		"super type": "Battle", "sub type": "Siege", "defense": "5",
	})

	if !eigen.Verteidigung.Steht || eigen.Verteidigung.Zahl != 5 {
		t.Fatalf("Verteidigung = %+v", eigen.Verteidigung)
	}
}
