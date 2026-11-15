package app

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEinUmgelenkterBestandHaeltDenUmzugOffen hält Zweitblick über Phase 6.4,
// Befund 13: ein Lauf mit eigenem Bestand legte den Zwischenspeicher am
// Ort des Rechners an — und weil der Umzug nur in einen **fehlenden** Ort
// ging, fand er danach nie mehr statt.
//
// **Der Speicher bleibt beim Rechner** (`bestandsnachbarn_test.go`); was sich
// ändert, ist der Umzug: er führt zusammen, statt vor einem vorhandenen Ort
// aufzugeben.
func TestEinUmgelenkterBestandHaeltDenUmzugOffen(t *testing.T) {
	heim := t.TempDir()
	t.Setenv("HOME", heim)
	t.Setenv(Bestandsumgebung, filepath.Join(heim, "probe", "bestand.db"))

	ort, fehler := scryfallverzeichnis(true)
	if fehler != nil {
		t.Fatal(fehler)
	}

	// **Was der Lauf mit eigenem Bestand hinterließ**: ein Bild und eine
	// Antwort, die es auch im alten Speicher gibt.
	for name, inhalt := range map[string]string{
		"bilder/neu.jpg": "neu", "doppelt.json": "neu",
	} {
		if anlegefehler := os.MkdirAll(filepath.Dir(filepath.Join(ort, name)), 0o700); anlegefehler != nil {
			t.Fatal(anlegefehler)
		}

		if schreibfehler := os.WriteFile(filepath.Join(ort, name), []byte(inhalt), 0o600); schreibfehler != nil {
			t.Fatal(schreibfehler)
		}
	}

	// **Und der nächste gewöhnliche Start zieht trotzdem um** — ein Start
	// ohne alten Speicher lässt alles, wie es ist.
	t.Setenv(Bestandsumgebung, "")

	if _, leerfehler := scryfallverzeichnis(true); leerfehler != nil {
		t.Fatal(leerfehler)
	}

	bestand, fehler := Datenbankpfad()
	if fehler != nil {
		t.Fatal(fehler)
	}

	alt := filepath.Join(filepath.Dir(bestand), "scryfall")
	for name := range map[string]bool{"bilder/alt.jpg": true, "alt.json": true, "doppelt.json": true} {
		if fehler := os.MkdirAll(filepath.Dir(filepath.Join(alt, name)), 0o700); fehler != nil {
			t.Fatal(fehler)
		}

		if fehler := os.WriteFile(filepath.Join(alt, name), []byte("alt"), 0o600); fehler != nil {
			t.Fatal(fehler)
		}
	}

	if _, fehler := scryfallverzeichnis(true); fehler != nil {
		t.Fatal(fehler)
	}

	for name, erwartet := range map[string]string{
		"bilder/alt.jpg": "alt", "bilder/neu.jpg": "neu", "alt.json": "alt", "doppelt.json": "neu",
	} {
		inhalt, lesefehler := os.ReadFile(filepath.Join(ort, name)) // #nosec G304 -- aus t.TempDir()
		if lesefehler != nil || string(inhalt) != erwartet {
			t.Errorf("%s: %q, %v — erwartet %q", name, inhalt, lesefehler, erwartet)
		}
	}

	if _, fehler := os.Stat(alt); fehler == nil {
		t.Error("der alte Ordner blieb liegen")
	}
}

// TestHalbGeschriebenesZaehltNicht hält Zweitblick über Phase 6.4, Befund 14:
// Bilder und Sammeldaten zählten jede Datei im Ordner — auch die
// Nachbardatei eines Bildes, das gerade geschrieben wird, und die `.teil` einer
// laufenden Sammeldatei. Ein Leeren löschte sie **unter dem Lauf weg**.
func TestHalbGeschriebenesZaehltNicht(t *testing.T) {
	ort := t.TempDir()

	for name, pfad := range map[string]string{
		SpeicherBilder:      filepath.Join(ort, "bilder"),
		SpeicherSammeldaten: filepath.Join(ort, "sammeldaten"),
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := os.MkdirAll(pfad, 0o700); fehler != nil {
				t.Fatal(fehler)
			}

			fertig := filepath.Join(pfad, "fertig.jpg")
			halbe := []string{
				filepath.Join(pfad, ".fertig.jpg.12345"),
				filepath.Join(pfad, "default-cards.jsonl.gz.6789.teil"),
			}

			for _, datei := range append([]string{fertig}, halbe...) {
				if fehler := os.WriteFile(datei, []byte("x"), 0o600); fehler != nil {
					t.Fatal(fehler)
				}
			}

			speicher := plattenspeicher(ort)[name]
			if belegung := speicher.Belegung(); belegung.Eintraege != 1 {
				t.Errorf("Einträge = %d, erwartet 1", belegung.Eintraege)
			}

			speicher.Leeren()

			for _, datei := range halbe {
				if _, fehler := os.Stat(datei); fehler != nil {
					t.Errorf("Leeren nahm %s unter dem Lauf weg", filepath.Base(datei))
				}
			}
		})
	}
}
