//go:build fremdpakete

package service_test

// **Die zwölf Kartenskripte von `magic.mse-game`, wirklich ausgeführt** —
// #1337.
//
// **Hinter `fremdpakete`** (§3, §11): das Spielpaket liegt nicht im
// Repositorium, und dieser Lauf zählt nicht zur Abdeckung.
//
// # Warum er trotzdem gefahren wird
//
// Weil eine selbst geschriebene Probe sagt, dass `new_card` **eine** Karte
// baut — nicht, dass die zwölf Skripte durchlaufen, die es wirklich gibt. Sie
// benutzen `for`-Schleifen, Bedingungen und Feldnamen, die niemand hier
// erfunden hat.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieEchtenKartenskripteLaufen(t *testing.T) {
	paket, fehler := msepaket.Oeffne(
		"/Users/david/full-magic-pack/data/magic.mse-game")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	paket.ModuleEinbinden(msepaket.NeuerPaketleser(
		"/Users/david/full-magic-pack/data"))

	spiel, _ := paket.Spielmodell()

	// **Das Spielskript gehört dazu.** `magic.mse-game` erklärt darin
	// `skeleton_runner` und `skeleton_script`; ohne es scheiterten sieben der
	// zwölf Skripte mit „der Name ist hier nicht bekannt".
	spielskript := spiel.Skript

	skripte := paket.Kartenskripte()
	if len(skripte) != 12 {
		t.Fatalf("%d Skripte, erwartet waren zwölf", len(skripte))
	}

	geglueckt := 0

	for _, skript := range skripte {
		karten, _, fehler := service.KartenAusSkript(
			t.Context(), skript.Skript, spiel.Fields, spielskript)
		if fehler != nil {
			t.Logf("%q: %v", skript.Name, fehler)

			continue
		}

		if len(karten) == 0 {
			t.Errorf("%q gab keine Karte", skript.Name)

			continue
		}

		geglueckt++

		// **Jede Karte trägt mindestens einen Wert** — und keinen mit einem
		// Feldnamen in Skriptschreibweise.
		//
		// **Nicht „jede Karte hat einen Namen".** Der erste Anlauf verlangte
		// das und lag falsch: `&Common Cycle` legt fünf **leere** Karten an,
		// je eine Farbe, mit nichts als `casting_cost`. Nachgelesen in
		// `add_cards_scripts` — die Zusicherung war strenger als die Datei.
		for _, karte := range karten {
			if len(karte.Values) == 0 {
				t.Errorf("%q gab eine Karte ohne jeden Wert", skript.Name)
			}

			for _, wert := range karte.Values {
				if strings.Contains(wert.Field, "_") {
					t.Errorf("%q gab das Feld %q in Skriptschreibweise",
						skript.Name, wert.Field)
				}
			}
		}
	}

	// **Zwölf von zwölf, und die Zahl steht hier als Zahl.**
	//
	// Der Weg dahin ging über drei Befunde, die **keine Probe** gezeigt hätte
	// und die alle erst am echten Paket auffielen:
	//
	//  1. **Null von zwölf**: ein Feld heißt in der Datei `casting cost`, im
	//     Skript `casting_cost`.
	//  2. **Fünf von zwölf**: die „Design Skeleton"-Einträge rufen
	//     `skeleton_runner` aus dem Spielskript — und eine nackte Umgebung
	//     kennt nicht einmal `length`.
	//  3. **Sieben von zwölf**: `card_color` ist ein `multiple choice`, und
	//     die Behauptung, das lasse sich nicht setzen, war **nicht
	//     nachgelesen**. `MultipleChoiceValue` leitet sehr wohl von
	//     `ChoiceValue` ab.
	//
	// **Wer diese Zahl senkt, hat etwas kaputtgemacht.**
	if geglueckt != len(skripte) {
		t.Errorf("%d von %d Skripten liefen durch", geglueckt, len(skripte))
	}
}
