// `mse oeffnen` und `mse sichern` — die beiden Enden des Arbeitsablaufs
// (#886).
//
// **Sie sind die zwei, die im Fenster an einem Dateidialog hängen.** Genau das
// macht sie für eine Kommandozeile interessant: dort ist der Pfad ein
// Argument, und was der Dialog verdeckt, muss die Fähigkeit benennen.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// probeset gibt den Pfad eines echten Setpakets aus `testdaten/`.
//
// **Selbst geschrieben und nicht fremd** (§11): `testdaten/` enthält keine
// fremden Vorlagen.
func probeset(t *testing.T) string {
	t.Helper()

	pfad, err := filepath.Abs(filepath.Join("..", "..", "testdaten", "probe.mse-set"))
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	if _, err := os.Stat(pfad); err != nil {
		t.Fatalf("das Probeset fehlt: %v", err)
	}

	return pfad
}

// **`oeffnen` fährt ein Set ein und nennt seine Kennung.**
func TestOeffnenFaehrtEinSetEin(t *testing.T) {
	ort := frischerBestand(t)

	code, aus, fehler := fahreAn(t, ort, "oeffnen", probeset(t), "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var gelesen oeffnenAntwort
	if err := json.Unmarshal([]byte(aus), &gelesen); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, aus)
	}

	if gelesen.SetID == "" {
		t.Error("die Antwort nennt keine Setkennung")
	}

	if gelesen.Karten == 0 {
		t.Error("das Set kam ohne Karten an")
	}

	// **Befunde sind eine Liste, auch wenn sie leer ist** (§3).
	if gelesen.Befunde == nil {
		t.Error("`befunde` ist null statt einer leeren Liste")
	}

	// Und danach steht es im Bestand.
	_, sets, _ := fahreAn(t, ort, "sets")
	if !strings.Contains(sets, gelesen.SetID) {
		t.Errorf("das eingefahrene Set steht nicht in `sets`: %q", sets)
	}
}

// **Ein Pfad, den es nicht gibt, ist ein fachliches Nein.**
func TestOeffnenEinesFehlendenPfads(t *testing.T) {
	code, _, fehler := fahre(t, "oeffnen", filepath.Join(t.TempDir(), "gibtsnicht.mse-set"))

	if code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet war %d (%s)", code, CodeFachlich, fehler)
	}
}

// **Ohne Pfad ist es ein Bedienfehler.**
func TestOeffnenOhnePfad(t *testing.T) {
	code, _, _ := fahre(t, "oeffnen")

	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeBedienung)
	}
}

// **`sichern` schreibt das Set zurück — an einen genannten Ort.**
func TestSichernSchreibtDasSet(t *testing.T) {
	ort := frischerBestand(t)

	_, aus, _ := fahreAn(t, ort, "oeffnen", probeset(t), "--json")

	var eingefahren oeffnenAntwort
	if err := json.Unmarshal([]byte(aus), &eingefahren); err != nil {
		t.Fatalf("kein JSON: %v", err)
	}

	ziel := filepath.Join(t.TempDir(), "wieder.mse-set")

	code, _, fehler := fahreAn(t, ort, "sichern", eingefahren.SetID, ziel)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if _, err := os.Stat(ziel); err != nil {
		t.Fatalf("die Zieldatei fehlt: %v", err)
	}
}

// **Ein Set, das es nicht gibt, ist ein fachliches Nein.**
func TestSichernEinesUnbekanntenSets(t *testing.T) {
	code, _, _ := fahre(t, "sichern", "4711", filepath.Join(t.TempDir(), "x.mse-set"))

	if code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeFachlich)
	}
}

// **Ohne Ziel ist es ein Bedienfehler.**
//
// **Und nicht „dorthin zurück, wo es herkam".** Im Fenster heißt ein leeres
// Ziel genau das (`SetSpeichern`), und es ist dort richtig: wer auf Speichern
// drückt, meint seine Datei. Auf der Kommandozeile wäre dasselbe ein
// Überschreiben ohne Rückfrage — der Aufruf sähe aus wie ein Ausdruck und
// änderte eine Datei.
func TestSichernOhneZiel(t *testing.T) {
	code, _, fehler := fahre(t, "sichern", "1")

	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeBedienung)
	}

	if !strings.Contains(fehler, "Ziel") {
		t.Errorf("die Meldung sagt nicht, was fehlt: %q", fehler)
	}
}

// **Ein zweites Einfahren derselben Datei legt kein zweites Set an.**
//
// Dieselbe Zusage wie im Fenster: eine Quelle, ein Set.
func TestOeffnenZweimalLegtEinSetAn(t *testing.T) {
	ort := frischerBestand(t)
	pfad := probeset(t)

	for i := range 2 {
		if code, _, fehler := fahreAn(t, ort, "oeffnen", pfad); code != CodeGut {
			t.Fatalf("Lauf %d: Rückgabewert %d (%s)", i+1, code, fehler)
		}
	}

	var raus, rfehler bytes.Buffer
	if code := lauf(probekontext(t), []string{"sets"}, nil, &raus, &rfehler, ort); code != CodeGut {
		t.Fatalf("sets: %d (%s)", code, rfehler.String())
	}

	if zeilen := strings.Count(strings.TrimSpace(raus.String()), "\n") + 1; zeilen != 1 {
		t.Errorf("%d Sets im Bestand, erwartet war eines:\n%s", zeilen, raus.String())
	}
}
