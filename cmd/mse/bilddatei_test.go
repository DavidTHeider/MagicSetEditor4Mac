package main

// **Ein Kartenbild von der Platte — an der Kommandozeile** (#1291).

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// probebild schreibt ein winziges PNG und gibt seinen Pfad.
func probebild(t *testing.T, breite, hoehe int) string {
	t.Helper()

	bild := image.NewRGBA(image.Rect(0, 0, breite, hoehe))
	for y := range hoehe {
		for x := range breite {
			bild.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}

	var roh bytes.Buffer
	if err := png.Encode(&roh, bild); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	pfad := filepath.Join(t.TempDir(), "quelle.png")
	if err := os.WriteFile(pfad, roh.Bytes(), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}

// mitKunstfenster legt ein Set an, dessen Stil ein Bildfeld führt.
//
// **Nicht das Probeset**: `probe-stil` hat keine Fläche für ein Bild, und ein
// Test dagegen prüfte den Fehlerweg statt der Fähigkeit. `probe-kunst` aus der
// Saat hat eine — 315 × 230 (§11).
func mitKunstfenster(t *testing.T) (ort, setID, karteID string) {
	t.Helper()

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatBereitstellen(t, zuhause)

	ort = frischerBestand(t)

	_, aus, fehler := fahreAn(t, ort, "set-neu", "Kunstset", "probe", "probe-kunst", "--json")

	var angelegt setAnlegeAntwort

	if err := json.Unmarshal([]byte(aus), &angelegt); err != nil {
		t.Fatalf("Set anlegen: %v (%s)", err, fehler)
	}

	_, kartenaus, kartenfehler := fahreAn(t, ort, "karte", angelegt.ID, "anlegen", "--json")

	var karte karteAntwort

	if err := json.Unmarshal([]byte(kartenaus), &karte); err != nil {
		t.Fatalf("Karte anlegen: %v (%s)", err, kartenfehler)
	}

	return ort, angelegt.ID, karte.KarteID
}

// **`bild-setzen` schreibt ein Bild in ein Kartenfeld.**
func TestBildSetzenSchreibtEinFeld(t *testing.T) {
	ort, setID, karteID := mitKunstfenster(t)

	code, aus, fehler := fahreAn(t, ort,
		"bild-setzen", setID, karteID, "kunst", probebild(t, 400, 300), "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var gesetzt bildsetzung
	if err := json.Unmarshal([]byte(aus), &gesetzt); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, aus)
	}

	if gesetzt.QuelleBreite != 400 || gesetzt.QuelleHoehe != 300 {
		t.Errorf("die Quellmaße stimmen nicht: %+v", gesetzt)
	}

	if gesetzt.ZielBreite == 0 || gesetzt.ZielHoehe == 0 {
		t.Errorf("die Zielmaße fehlen: %+v", gesetzt)
	}

	// **Und das Feld trägt danach einen Namen** — nicht nur die Antwort.
	_, felder, _ := fahreAn(t, ort, "karten", setID, "--json")
	if !strings.Contains(felder, karteID) {
		t.Errorf("die Karte fehlt: %s", felder)
	}
}

// **Eine Datei, die kein Bild ist, ist ein fachliches Nein.**
func TestBildSetzenMitEinerDateiDieKeinBildIst(t *testing.T) {
	ort, setID, karteID := mitKunstfenster(t)

	kein := filepath.Join(t.TempDir(), "kein.png")
	if err := os.WriteFile(kein, []byte("hallo"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if code, _, _ := fahreAn(t, ort,
		"bild-setzen", setID, karteID, "kunst", kein); code == CodeGut {
		t.Error("eine Textdatei ging als Bild durch")
	}
}

// **Und die Bedienfehler.**
func TestBildSetzenBedienfehler(t *testing.T) {
	ort, setID, karteID := mitKunstfenster(t)

	faelle := map[string][]string{
		"ohne Pfad":            {"bild-setzen", setID, karteID, "kunst"},
		"krumme Setkennung":    {"bild-setzen", "x", karteID, "kunst", "/tmp/a.png"},
		"krumme Kartenkennung": {"bild-setzen", setID, "x", "kunst", "/tmp/a.png"},
	}

	for name, aufruf := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, aufruf...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
			}
		})
	}
}
