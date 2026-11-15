// Was die Scryfall-Gegenstelle ausliefert und was die Läufe damit machen.
//
// **Eine eigene Datei**, weil `scryfallgegenstelle_test.go` an der
// Zeilengrenze aus §2 stand — und weil der Schnitt der Frage folgt: dort
// **die Gegenstelle**, hier **der Stoff**, den sie ausgibt, und die Griffe,
// mit denen ein Lauf sie benutzt.
package flaechen_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"testing"
)

// gepackteSammeldatei baut eine `.jsonl.gz` mit einer Karte.
func gepackteSammeldatei(t *testing.T, grund string) []byte {
	t.Helper()

	var gepackt bytes.Buffer

	schreiber := gzip.NewWriter(&gepackt)
	if _, fehler := schreiber.Write([]byte(
		einzeilig(fmt.Sprintf(probekarte, "Erste Probe", 1, grund, grund, 1)) + "\n")); fehler != nil {
		t.Fatalf("packen: %v", fehler)
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return gepackt.Bytes()
}

// einzeilig macht aus eingerücktem JSON eine Zeile — JSONL verlangt es.
func einzeilig(roh string) string {
	var wert any
	if err := json.Unmarshal([]byte(roh), &wert); err != nil {
		return roh
	}

	eine, err := json.Marshal(wert)
	if err != nil {
		return roh
	}

	return string(eine)
}

// winzigesPNG gibt ein 2×2-Bild.
func winzigesPNG(t *testing.T) []byte {
	t.Helper()

	bild := image.NewRGBA(image.Rect(0, 0, 2, 2))
	bild.Set(0, 0, color.RGBA{R: 200, G: 30, B: 30, A: 255})

	var roh bytes.Buffer
	if err := png.Encode(&roh, bild); err != nil {
		t.Fatalf("PNG: %v", err)
	}

	return roh.Bytes()
}

func schreibeText(t *testing.T, schreiben http.ResponseWriter, inhalt string) {
	t.Helper()

	schreiben.Header().Set("Content-Type", "application/json")
	schreibeBytes(t, schreiben, []byte(inhalt))
}

// schreibeBytes gibt einen Rumpf aus.
//
// **Der Inhalt stammt aus dieser Datei**, nicht aus einer Anfrage — die
// Taint-Analyse sieht nur, dass etwas Geschriebenes ankommt.
func schreibeBytes(t *testing.T, schreiben http.ResponseWriter, inhalt []byte) {
	t.Helper()

	// **G705 sieht eine Ausgabe, die aus der Anfrage stammen könnte** — hier
	// stammt sie aus dieser Datei, und die Gegenstelle antwortet keinem
	// Browser. Dieselbe Stelle steht in `app/scryfallbildnaht_test.go`.
	//nolint:gosec // G705: Prüfgegenstelle, keine Ausgabe an einen Browser
	if _, fehler := schreiben.Write(inhalt); fehler != nil {
		t.Errorf("schreiben: %v", fehler)
	}
}

// magicset legt ein Set mit dem Spiel `magic` an und gibt seine Kennung.
//
// # Warum nicht das Probeset aus `testdaten/`
//
// **Es führt das Spiel »probe«**, und für das gibt es keine Zuordnung auf
// Scryfall-Felder — jeder Scryfall-Weg endete daran, und der Lauf prüfte die
// Abweisung statt der Fähigkeit. Gemessen: „für das Spiel »probe« gibt es
// keine Zuordnung auf Scryfall-Felder", an allen drei Flächen gleichlautend.
//
// Angelegt wird es über den Satzimport — das ist der einzige Weg ohne Fenster,
// der ein Set mit einem gewählten Spiel anlegt.
func magicset(t *testing.T, umgebung []string) string {
	t.Helper()

	mussKommandozeile(t, umgebung, "scryfall-set", "prf", "magic", "--json")

	sets := setzeilenAus(t, mussKommandozeile(t, umgebung, "sets", "--json"))
	if len(sets) != 1 {
		t.Fatalf("%d Sets im Bestand, erwartet eines", len(sets))
	}

	return sets[0].ID
}

// mussKommandozeile ruft `mse` und bricht ab, wenn es nicht glückt.
func mussKommandozeile(t *testing.T, umgebung []string, argumente ...string) string {
	t.Helper()

	aus, fehler, code := kommandozeile(t, umgebung, argumente...)
	if code != 0 {
		t.Fatalf("mse %v: %d %s%s", argumente, code, aus, fehler)
	}

	return aus
}
