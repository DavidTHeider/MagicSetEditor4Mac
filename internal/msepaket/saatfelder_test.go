package msepaket_test

// **Das lange Textfeld der eigenen Saat ist mehrzeilig** — FP-02
// (#2156).
//
// Gegen das echte Paket unter `saat/`, nicht gegen eine Nachbildung: die
// Werkstattkarten tragen Absätze und einen Text von elf Zeilen, und das
// Feld `text` erklärte sich ohne `multi line:` — der Editor zeigte dafür
// eine einzelne schmale Zeile. MSE2 liest den Schlüssel als `multi_line`
// (`src/data/field/text.cpp`, `REFLECT(multi_line)`, Fassung f9d9356 vom
// 2021-07-06); fremde Pakete erklären ihn längst.

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDasLangeTextfeldDerSaatIstMehrzeilig(t *testing.T) {
	t.Parallel()

	paket, fehler := msepaket.Oeffne(filepath.Join("..", "..", "saat", "probe.mse-game"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = paket.Schliessen() }) //nolint:errcheck // nur gelesen

	spiel, befunde := paket.Spielmodell()
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	mehrzeilig := map[string]bool{}
	for _, feld := range spiel.Fields {
		mehrzeilig[feld.Name] = feld.Mehrzeilig
	}

	// **`text` ja, `name` nein** — ein Name ist eine Zeile, und ein Feld,
	// das alles mehrzeilig machte, wäre der globale Zwang, den FP-02
	// ausschließt.
	if !mehrzeilig["text"] || mehrzeilig["name"] {
		t.Errorf("mehrzeilig: %v", mehrzeilig)
	}
}
