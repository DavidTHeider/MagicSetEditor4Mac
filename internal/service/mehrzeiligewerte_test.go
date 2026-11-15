package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein mehrzeiliger Feldwert kommt vollständig an.**
//
// In einer MSE-Datei steht ein langer Text als **eingerückter Block** unter
// seinem Schlüssel:
//
//	rule_text:
//		Erste Zeile
//		Zweite Zeile
//
// Der Wert hinter dem Doppelpunkt ist dann leer, und die Zeilen darunter sind
// Fortsetzungen. Wer nur den Wert liest, bekommt ein leeres Feld — und weil ein
// Textfeld ohne Wert gar nicht gezeichnet wird, ohne Befund. Auf einer
// gewöhnlichen Karte trifft das Regeltext **und** Flavortext, also den halben
// Kartentext (#398).
func TestEinMehrzeiligerFeldwertKommtAn(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "mehrzeilig.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse_version: 2.1.2\ngame: probe\nstylesheet: probe-schlicht\n" +
		"card:\n\tname: Feuerball\n\trule_text:\n\t\tErste Zeile\n\t\tZweite Zeile\n" +
		"\tflavor_text:\n\t\tNur eine Zeile\n"

	if err := os.WriteFile(filepath.Join(ort, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	speicher := &speicherDouble{}
	if _, err := service.NeuerImporteur(speicher).Set(t.Context(), ort); err != nil {
		t.Fatalf("Import: %v", err)
	}

	karte := speicher.set.Cards[0]

	text, da := karte.Value("rule text")
	if !da {
		t.Fatal("die Karte führt kein Feld `rule text`")
	}

	if text != "Erste Zeile\nZweite Zeile" {
		t.Errorf("Regeltext %q, erwartet zwei Zeilen", text)
	}

	flavor, _ := karte.Value("flavor text")
	if flavor != "Nur eine Zeile" {
		t.Errorf("Flavortext %q", flavor)
	}

	// **Der einzeilige Wert bleibt, wie er war.**
	if name, _ := karte.Value("name"); name != "Feuerball" {
		t.Errorf("Name %q", name)
	}

	// **Und die Quelldatei ist unberührt** — gelesen wird, nicht geschrieben.
	gelesen, err := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !strings.Contains(string(gelesen), "\t\tErste Zeile\n") {
		t.Errorf("die Quelldatei hat sich verändert: %q", gelesen)
	}
}
