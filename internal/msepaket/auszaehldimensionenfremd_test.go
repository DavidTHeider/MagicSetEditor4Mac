//go:build fremdpakete

// **Die Auszählungsdimensionen an einem echten Spielpaket** — Werkstück C von
// Phase 4.11 (#1338).
//
// # Was dieser Lauf hält, das kein Doppel halten kann
//
// **Dass der Weg über `include localized file:` wirklich trägt.** Die Blöcke
// stehen nicht in `magic.mse-game/game`, sondern in `statistics_en` — und
// dorthin führt eine Kette aus zwei Einbindungen. Ein Doppel mit einem Block
// in der Hauptdatei prüft den Leser und nicht den Weg.
//
//	MSE_FREMDPAKETE=~/full-magic-pack/data \
//	  go test -tags fremdpakete ./internal/msepaket/ -run Auszaehl -v
package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDieAuszaehldimensionenVonMagic(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	paket, fehler := msepaket.Oeffne(filepath.Join(wurzel, "magic.mse-game"))
	if fehler != nil {
		t.Fatalf("magic.mse-game öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)); fehler != nil {
		t.Fatalf("Module einbinden: %v", fehler)
	}

	dimensionen := paket.Auszaehldimensionen()

	// **Fünfzig, als Gleichheit und nicht als Untergrenze** — gezählt am
	// 2026-08-31 in `statistics_en`. Eine Untergrenze bliebe grün, wenn der
	// Rückfall auf `_en` ausfiele und stattdessen eine kürzere Fassung käme.
	if len(dimensionen) != 50 {
		t.Fatalf("%d Dimensionen — gezählt waren 50 in statistics_en", len(dimensionen))
	}

	nachName := map[string]msepaket.Auszaehldimension{}
	for _, eine := range dimensionen {
		nachName[eine.Name] = eine
	}

	farbgruppe, gibtEs := nachName["color category"]
	if !gibtEs {
		t.Fatalf("die Dimension color category fehlt — gelesen wurden %d Namen", len(nachName))
	}

	// **Die erklärte Reihenfolge ist der Ertrag dieser Zeile.** Land steht
	// vorn, nicht alphabetisch hinten; wer sie fallen ließe, bekäme eine
	// Auszählung, die anders sortiert ist als jede MSE2-Ansicht.
	erwartet := []string{
		"Land", "Colorless", "Artifact", "White", "Blue",
		"Black", "Red", "Green", "Multicolor", "Hybrid",
	}

	if len(farbgruppe.Gruppen) != len(erwartet) {
		t.Fatalf("%d Gruppen, erwartet %d: %+v",
			len(farbgruppe.Gruppen), len(erwartet), farbgruppe.Gruppen)
	}

	for i, name := range erwartet {
		if farbgruppe.Gruppen[i] != name {
			t.Errorf("Gruppe %d ist %q, erwartet %q", i, farbgruppe.Gruppen[i], name)
		}
	}

	if !farbgruppe.Zerlegen {
		t.Error("color category trägt `split list: true` und kam ohne an")
	}

	if farbgruppe.Farben["Land"] != "rgb(103,64,35)" {
		t.Errorf("die Farbe von Land ist %q", farbgruppe.Farben["Land"])
	}

	if farbgruppe.Skript == "" {
		t.Error("das Skript ist leer — ohne es rechnet die Dimension nichts")
	}

	// **Und die Gegenprobe auf die Zahl daneben**: `statistics category:`
	// kommt im ganzen Pack null-mal vor. Wenn ein späteres Pack eine
	// mitbrächte, soll dieser Lauf es sagen — die Aussage steht als Zahl in
	// `auszaehldimensionen.go`.
	if kategorien := paket.Kopf.Alle("statistics category"); len(kategorien) != 0 {
		t.Errorf("%d `statistics category:` — der Kommentar in "+
			"auszaehldimensionen.go behauptet null", len(kategorien))
	}
}
