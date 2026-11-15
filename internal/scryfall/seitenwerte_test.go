// **Beide Seiten einer Karte kommen an** —
// #1040.
package scryfall_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// zweiseitig baut eine Karte mit zwei Seiten, wie Scryfall sie schickt.
//
// **Die obersten Felder sind bei einer solchen Karte leer** — `mana_cost`,
// `type_line` und `oracle_text` stehen je Seite. Genau deshalb kam bisher fast
// nichts an.
func zweiseitig() *scryfall.Karte {
	return &scryfall.Karte{
		Name:       "Delver of Secrets // Insectile Aberration",
		Gestalt:    "transform",
		Seltenheit: "common",
		Seiten: []scryfall.Seite{
			{
				Name: "Delver of Secrets", Manakosten: "{U}",
				Typzeile: "Creature — Human Wizard", Regeltext: "Etwas.",
				Staerke: "1", Widerstand: "1", Kuenstler: "Nils Hamm",
			},
			{
				Name: "Insectile Aberration", Typzeile: "Creature — Human Insect",
				Regeltext: "Flying", Staerke: "3", Widerstand: "2",
				Kuenstler: "Nils Hamm",
			},
		},
	}
}

// **Die erste Seite füllt die gewöhnlichen Felder, die zweite die mit ` 2`.**
//
// # Woher die Übereinkunft stammt
//
// Aus `magic-set-generator` (MIT, §11), `src/mse.rs`, Zeilen 190–200:
//
//	macro_rules! push_alt {
//	    ($key:expr, $val:expr) => {
//	        if alt { result.push(format!("{} 2", $key), $val) }
//	        else   { result.push($key, $val) }
//	    };
//	}
//
// Es ist die einzige belegte Auskunft darüber, was in einem echten Set steht —
// und `magic.mse-game/card_fields` führt die Felder: 46 mit ` 2`, gemessen am
// 2026-08-24.
func TestBeideSeitenFuellenIhreFelder(t *testing.T) {
	t.Parallel()

	zuordnung, _, bekannt := scryfall.ZuordnungFuer("magic")
	if !bekannt {
		t.Fatal("für magic gibt es keine Zuordnung")
	}

	werte := zuordnung.Werte(zweiseitig())

	if werte["name"] != "Delver of Secrets" {
		t.Errorf("name = %q — erwartet der Name der **ersten Seite**, nicht "+
			"»A // B«", werte["name"])
	}

	if werte["name 2"] != "Insectile Aberration" {
		t.Errorf("name 2 = %q", werte["name 2"])
	}

	if werte["casting cost"] != "U" {
		t.Errorf("casting cost = %q — er steht an der Seite, nicht oben",
			werte["casting cost"])
	}

	if werte["rule text 2"] != "Flying" {
		t.Errorf("rule text 2 = %q", werte["rule text 2"])
	}

	if werte["power 2"] != "3" || werte["toughness 2"] != "2" {
		t.Errorf("power 2 = %q, toughness 2 = %q", werte["power 2"], werte["toughness 2"])
	}

	// **Was nur oben steht, gilt für beide.** Die Seltenheit ist eine
	// Eigenschaft des Drucks, nicht der Seite — und `push_alt!` schreibt sie
	// in der Referenz für beide Teile.
	if werte["rarity"] != "common" {
		t.Errorf("rarity = %q", werte["rarity"])
	}
}

// **Eine einseitige Karte bleibt, wie sie war.**
//
// Die Gegenprobe zur Zeile darüber: wer die Seitenbehandlung immer anwendete,
// schriebe für jede gewöhnliche Karte leere ` 2`-Felder.
func TestEineEinseitigeKarteBekommtKeineZweitenFelder(t *testing.T) {
	t.Parallel()

	zuordnung, _, _ := scryfall.ZuordnungFuer("magic")

	werte := zuordnung.Werte(&scryfall.Karte{
		Name: "Lightning Bolt", Manakosten: "{R}", Typzeile: "Instant",
		Regeltext: "Fügt 3 Schaden zu.", Seltenheit: "common",
	})

	for feld := range werte {
		if len(feld) > 2 && feld[len(feld)-2:] == " 2" {
			t.Errorf("eine einseitige Karte bekam %q", feld)
		}
	}
}

// **Eine Karte mit genau einer Seite ebenso.**
//
// Scryfall führt `card_faces` auch bei manchen einseitigen Karten — dann steht
// dort **eine** Seite, und eine zweite zu erfinden wäre eine Vermutung.
func TestEineKarteMitEinerSeiteBekommtKeineZweitenFelder(t *testing.T) {
	t.Parallel()

	zuordnung, _, _ := scryfall.ZuordnungFuer("magic")

	werte := zuordnung.Werte(&scryfall.Karte{
		Name: "Einseitig", Seltenheit: "common",
		Seiten: []scryfall.Seite{{Name: "Einseitig", Manakosten: "{R}"}},
	})

	if werte["name"] != "Einseitig" {
		t.Errorf("name = %q", werte["name"])
	}

	if _, da := werte["name 2"]; da {
		t.Errorf("es entstand ein Feld »name 2«")
	}
}

// **Eine Seite, die es nicht gibt, gibt die Karte zurück.**
//
// Der Zweig ist kein Zierat: `MitSeite` ist exportiert, und ein Aufrufer mit
// einer Stelle außerhalb bekäme sonst einen Absturz statt einer Karte.
func TestMitSeiteAusserhalbGibtDieKarte(t *testing.T) {
	t.Parallel()

	karte := zweiseitig()

	for name, stelle := range map[string]int{"vor der ersten": -1, "hinter der letzten": 2} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if karte.MitSeite(stelle) != karte {
				t.Errorf("MitSeite(%d) gab eine andere Karte", stelle)
			}
		})
	}
}
