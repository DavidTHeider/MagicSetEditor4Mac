package msepaket_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Schriftfarbe darf ein eingerückter `script:`-Block sein.**
//
// Dieselbe Form wie bei `left:` und `alignment:` — und dieselbe Falle: wer nur
// `knoten.Wert` liest, bekommt die leere Zeichenkette und damit **Schwarz**,
// ohne dass irgendwo etwas fehlt.
//
// In `M15-Magic-Pack` steht sie sechsmal, alle in `magic-m15-future.mse-style`
// (#382).
func TestEineSchriftfarbeDarfEinSkriptblockSein(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "probe-farbe.mse-style")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nshort name: Farbe\n" +
		"card width: 375\ncard height: 523\n" +
		"card style:\n\tname:\n\t\tleft: 0\n\t\ttop: 0\n\t\twidth: 100\n\t\theight: 20\n" +
		"\t\tfont:\n\t\t\tname: Helvetica\n\t\t\tsize: 12\n\t\t\tcolor:\n\t\t\t\tscript:\n" +
		"\t\t\t\t\tif card.farbe == \"rot\" then rgb(255,0,0)\n\t\t\t\t\telse rgb(0,0,255)\n"

	if err := os.WriteFile(filepath.Join(ort, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	paket, err := msepaket.Oeffne(ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, _ := paket.Stilmodell()

	feld, da := stil.Feld("name")
	if !da {
		t.Fatal("der Stil führt kein Feld `name`")
	}

	if !strings.Contains(feld.Schrift.Farbe, "rgb(255,0,0)") {
		t.Errorf("die Schriftfarbe ist %q, erwartet den Skriptrumpf", feld.Schrift.Farbe)
	}

	// **Geklammert**, denn genau so kommt sie beim Auswerten an: `service`
	// streift die Klammer ab und rechnet, was dazwischen steht.
	if !strings.HasPrefix(strings.TrimSpace(feld.Schrift.Farbe), "{") {
		t.Errorf("die Schriftfarbe ist %q, erwartet eine geklammerte Angabe", feld.Schrift.Farbe)
	}
}

// **Ein Feldname in einer Datei ist kanonisch: `_` steht für ein Leerzeichen.**
//
//	void canonical_name_form_in_place(String& str) { … ' ' → '_' … }
//	void uncanonical_name_form_in_place(String& str) { … '_' → ' ' … }
//
// `src/util/string.cpp`, Zeilen 121–133, Fassung f9d9356 vom 2021-07-06.
//
// **Beide Richtungen stehen hier**, weil beide gebraucht werden: gelesen wird
// entkanonisch, geschrieben kanonisch
// (#396).
func TestKanonischUndZurueck(t *testing.T) {
	faelle := map[string]string{
		"rule text":    "rule_text",
		"casting cost": "casting_cost",
		"name":         "name",
		"":             "",
	}

	for mitLeerzeichen, kanonisch := range faelle {
		if gebaut := msepaket.Kanonisch(mitLeerzeichen); gebaut != kanonisch {
			t.Errorf("Kanonisch(%q) = %q, erwartet %q", mitLeerzeichen, gebaut, kanonisch)
		}

		if gebaut := msepaket.Entkanonisch(kanonisch); gebaut != mitLeerzeichen {
			t.Errorf("Entkanonisch(%q) = %q, erwartet %q", kanonisch, gebaut, mitLeerzeichen)
		}
	}
}
