// `mse zeichnen` — eine Karte als Bild, ohne Fenster
// (#886).
//
// **Die Fähigkeit, die zeigt, ob der schwere Weg trägt.** Zwischen einer
// Kartenkennung und einer PNG-Datei liegen sieben Teile — Paketwurzeln,
// Stilspeicher, Schriftsuche, Bildquelle, Planbauer, Rasterer, Ausfuhr. Wer
// sie ohne Fenster zusammenbekommt, hat den Fachkern von der Oberfläche
// gelöst; wer nicht, hat ihn nur behauptet.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// eingefahren fährt das Probeset ein und gibt Ort, Set- und Kartenkennung.
func eingefahren(t *testing.T) (ort, karteID string) {
	t.Helper()

	ort = frischerBestand(t)

	_, aus, fehler := fahreAn(t, ort, "oeffnen", probeset(t), "--json")

	var eingang oeffnenAntwort
	if err := json.Unmarshal([]byte(aus), &eingang); err != nil {
		t.Fatalf("einfahren: %v (%s)", err, fehler)
	}

	_, kartenaus, _ := fahreAn(t, ort, "karten", eingang.SetID, "--json")

	var karten kartenAntwort
	if err := json.Unmarshal([]byte(kartenaus), &karten); err != nil {
		t.Fatalf("Karten lesen: %v", err)
	}

	if len(karten.Karten) == 0 {
		t.Fatal("das Probeset kam ohne Karten an")
	}

	return ort, karten.Karten[0].ID
}

// **`zeichnen` schreibt eine Bilddatei — und sie ist keine leere Datei.**
//
// Gefahren, nicht behauptet: der ganze Weg vom Bestand bis zu den Bytes.
func TestZeichnenSchreibtEinBild(t *testing.T) {
	ort, karteID := eingefahren(t)

	ziel := filepath.Join(t.TempDir(), "karte.png")

	code, _, fehler := fahreAn(t, ort, "zeichnen", karteID, ziel)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	angaben, err := os.Stat(ziel)
	if err != nil {
		t.Fatalf("die Bilddatei fehlt: %v", err)
	}

	if angaben.Size() == 0 {
		t.Error("die Bilddatei ist leer")
	}

	// **Und sie ist wirklich ein PNG.** Eine Datei mit Inhalt, die kein Bild
	// ist, sähe von außen genauso aus.
	roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !strings.HasPrefix(string(roh[:8]), "\x89PNG\r\n\x1a\n") {
		t.Errorf("die Datei beginnt nicht mit der PNG-Kennung: % x", roh[:8])
	}
}

// **Die Zielendung bestimmt das Format.**
//
// Eine Kommandozeile, die dafür einen eigenen Schalter verlangte, ließe
// `mse zeichnen 1 karte.svg --format png` zu — und schriebe ein PNG unter
// einem Namen, der SVG verspricht.
func TestZeichnenNimmtDasFormatAusDerEndung(t *testing.T) {
	ort, karteID := eingefahren(t)

	ziel := filepath.Join(t.TempDir(), "karte.svg")

	if code, _, fehler := fahreAn(t, ort, "zeichnen", karteID, ziel); code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !strings.Contains(string(roh[:200]), "<svg") {
		t.Errorf("die Datei ist kein SVG: %q", string(roh[:100]))
	}
}

// **Eine unbekannte Endung ist ein Bedienfehler.**
func TestZeichnenMitUnbekannterEndung(t *testing.T) {
	ort, karteID := eingefahren(t)

	code, _, fehler := fahreAn(t, ort, "zeichnen", karteID,
		filepath.Join(t.TempDir(), "karte.xyz"))

	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet war %d (%s)", code, CodeBedienung, fehler)
	}
}

// **Eine Karte, die es nicht gibt, ist ein fachliches Nein.**
func TestZeichnenEinerUnbekanntenKarte(t *testing.T) {
	code, _, _ := fahre(t, "zeichnen", "4711", filepath.Join(t.TempDir(), "karte.png"))

	if code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeFachlich)
	}
}

// **Ohne Ziel ist es ein Bedienfehler.**
func TestZeichnenOhneZiel(t *testing.T) {
	code, _, _ := fahre(t, "zeichnen", "1")

	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeBedienung)
	}
}
