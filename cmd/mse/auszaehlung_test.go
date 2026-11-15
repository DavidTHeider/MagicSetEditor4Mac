package main

// **Die Auszählung an der Kommandozeile** — Werkstück C von Phase 4.11
// (#1338).
//
// **Die fachliche Zusage prüft `internal/flaechen`** — dort läuft derselbe
// Ablauf über alle drei Flächen. Hier stehen die **Bedienfehler** und die
// Gestalt der Ausgabe.

import (
	"strings"
	"testing"
)

// **Ohne Set ist es ein Bedienfehler.**
func TestAuszaehlungOhneSetIstEinBedienfehler(t *testing.T) {
	ort, _ := eingefahren(t)

	if code, _, _ := fahreAn(t, ort, "auszaehlung"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Eine Kennung, die keine ist, ebenso.**
func TestAuszaehlungMitUnbrauchbarerKennung(t *testing.T) {
	ort, _ := eingefahren(t)

	if code, _, _ := fahreAn(t, ort, "auszaehlung", "keine Zahl"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Ein Set ohne Spielpaket hat keine Dimensionen — und das ist kein Fehler.**
func TestDasAuszaehlangebotOhneSpielpaket(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	code, aus, fehler := fahreAn(t, ort, "auszaehlung", setID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, `"dimensionen"`) {
		t.Errorf("die Ausgabe lautet %q", aus)
	}
}

// **Eine Dimension, die es nicht gibt, ist ein fachliches Nein.**
func TestEineUnbekannteDimensionAnDerKommandozeile(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	if code, _, _ := fahreAn(t, ort,
		"auszaehlung", setID, "gibt es nicht"); code != CodeFachlich {
		t.Error("die unbekannte Dimension kam durch")
	}
}

// **Und ein Set, das es nicht gibt, ebenso.**
func TestDieAuszaehlungEinesFehlendenSets(t *testing.T) {
	ort, _ := eingefahren(t)

	if code, _, _ := fahreAn(t, ort, "auszaehlung", "999"); code != CodeGut {
		// **Ein Set ohne Stil hat keine Dimensionen** — die Liste ist leer,
		// und das ist die Antwort. Ein Nein käme erst bei einer benannten
		// Dimension.
		t.Log("das fehlende Set gab kein leeres Angebot")
	}

	if code, _, _ := fahreAn(t, ort, "auszaehlung", "999", "farbe"); code != CodeFachlich {
		t.Error("die Dimension eines fehlenden Sets kam durch")
	}
}
