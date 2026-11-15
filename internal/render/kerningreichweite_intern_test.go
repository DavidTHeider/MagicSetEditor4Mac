//go:build fremdpakete

// **Welche Schriften des Packs überhaupt kernen** —
// #752.
//
// # Der Befund
//
// `vorsilbenbreite.go` und `breitenaht_intern_test.go` sagten, es sei „kein
// Kerning im Spiel" — begründet damit, dass unsere Summe einzelner Vorschübe
// bei achtzehn Messungen knapp **unter** wx' ganzer Zahl liegt.
//
// **Gemessen wurde `mplantin.ttf`, und die trägt gar keine Kerningtabelle.**
// Die Beobachtung war damit nicht falsch, sondern leer: sie kann über Kerning
// nichts sagen. Angestrichen vom Zweitblick durch Codex am 2026-08-21.
//
// # Was daraus folgt
//
// **`beleren-bold` führt ein `kern`-Merkmal**, und das ist die Schrift, in der
// jeder Kartenname steht (`magic-m15.mse-style`, `font: name: Beleren Bold`).
// Unsere Breite summiert einzelne `GlyphAdvance` und lässt es weg; MSE gibt die
// ganze Vorsilbe an `GetTextExtent` (`src/render/text/font.cpp`, Zeile 27,
// Fassung f9d9356 vom 2021-07-06).
//
// **Wie weit das auseinanderläuft, ist seit QA-03 gemessen**
// (`docs/SCHRIFTUNTERSUCHUNG.md`). Dieser Wächter hält weiter fest, **worüber**
// die Breitenmessung von #598 spricht — damit niemand sie noch einmal für eine
// Aussage über Kerning nimmt.
package render

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// TestWelcheSchriftenDesPacksKernen hält die Reichweite der Breitenmessung fest.
func TestWelcheSchriftenDesPacksKernen(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	ordner := filepath.Join(filepath.Dir(wurzel), "Magic - Fonts")

	faelle := []struct {
		datei, wozu string
		kernt       bool
	}{
		{"mplantin.ttf", "Regeltext und Typzeile", false},
		{"beleren-bold_P1.01.ttf", "jeder Kartenname", true},
		{"belerensmallcaps-bold.ttf", "Kapitälchen", true},
		{"matrixb.ttf", "Archenemy", true},
		{"ModMatrix.ttf", "Planechase", true},
		{"MATRIX.TTF", "ältere Vorlagen", false},
	}

	for _, fall := range faelle {
		roh, err := os.ReadFile(filepath.Join(ordner, fall.datei)) //nolint:gosec // Pfad aus der Umgebung
		if err != nil {
			t.Skipf("keine Schriftdatei %s: %v", fall.datei, err)
		}

		if kerntSchrift(t, roh) != fall.kernt {
			t.Errorf("%s (%s): kernt = %v, erwartet %v — die Reichweite der "+
				"Breitenmessung gehört neu geschrieben",
				fall.datei, fall.wozu, !fall.kernt, fall.kernt)
		}
	}
}

// kerntSchrift sagt, ob eine Schriftdatei Kerningpaare mitbringt.
//
// **Zwei Wege, und beide zählen**: die alte `kern`-Tabelle und ein
// `kern`-Merkmal in `GPOS`. Eine Schrift, die nur eines von beiden führt,
// kernt trotzdem.
func kerntSchrift(t *testing.T, roh []byte) bool {
	t.Helper()

	tabellen := tabellenverzeichnis(roh)
	if _, da := tabellen["kern"]; da {
		return true
	}

	gpos, da := tabellen["GPOS"]
	if !da || int(gpos)+8 > len(roh) {
		return false
	}

	// GPOS: ScriptList, FeatureList, LookupList — die Liste der Merkmale steht
	// als zweiter Versatz im Kopf.
	liste := gpos + uint32(binary.BigEndian.Uint16(roh[gpos+6:gpos+8]))
	if int(liste)+2 > len(roh) {
		return false
	}

	for i := range int(binary.BigEndian.Uint16(roh[liste : liste+2])) {
		eintrag := int(liste) + 2 + i*6
		if eintrag+4 > len(roh) {
			return false
		}

		if string(roh[eintrag:eintrag+4]) == "kern" {
			return true
		}
	}

	return false
}

// tabellenverzeichnis gibt je Kennung den Versatz der Tabelle.
func tabellenverzeichnis(roh []byte) map[string]uint32 {
	aus := map[string]uint32{}
	if len(roh) < 12 {
		return aus
	}

	for i := range int(binary.BigEndian.Uint16(roh[4:6])) {
		eintrag := 12 + i*16
		if eintrag+16 > len(roh) {
			return aus
		}

		aus[string(roh[eintrag:eintrag+4])] = binary.BigEndian.Uint32(roh[eintrag+8 : eintrag+12])
	}

	return aus
}
