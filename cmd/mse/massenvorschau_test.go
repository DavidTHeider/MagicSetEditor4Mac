package main

import (
	"strings"
	"testing"
)

// **Die Vorschau einer Massenänderung an der Zeile** — Werkstück U6 von
// Phase 5.1 (#1417): `mse felder-vorschau <set> <feld> <wert> <karte>...`
// zeigt je Karte Alt und Neu und schreibt nichts.
func TestFelderVorschauZeigtAltUndNeuOhneZuSchreiben(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "felder-vorschau", setID, "artist", "Neu", karteID)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "Neu") || !strings.Contains(aus, karteID) {
		t.Errorf("die Ausgabe lautet %q", aus)
	}

	_, gelesen, _ := fahreAn(t, ort, "karten", setID, "--json")
	if strings.Contains(gelesen, `"Neu"`) {
		t.Errorf("die Vorschau hat geschrieben: %s", gelesen)
	}
}

func TestFelderVorschauAlsJSON(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "felder-vorschau", setID, "artist", "Neu", karteID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, `"zeilen"`) || !strings.Contains(aus, `"gabAlt"`) {
		t.Errorf("die Ausgabe lautet %q", aus)
	}
}

func TestFelderVorschauBraucht(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	for name, argumente := range map[string][]string{
		"ohne Karten":                  {"felder-vorschau", setID, "artist", "X"},
		"mit unbrauchbarer Setkennung": {"felder-vorschau", "keine-zahl", "artist", "X", "1"},
		"mit unbrauchbarer Karte":      {"felder-vorschau", setID, "artist", "X", "keine-zahl"},
	} {
		t.Run(name, func(t *testing.T) {
			if code, aus, _ := fahreAn(t, ort, argumente...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d (%s)", code, aus)
			}
		})
	}
}

// **`felder-setzen --erwartet` geht den vergleichenden Weg**: die Erwartung
// kommt als JSON, und eine Karte, die inzwischen anders dasteht, lässt den
// ganzen Aufruf scheitern.
func TestFelderSetzenMitErwartungWeistEinenKonfliktAb(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, aus, _ := fahreAn(t, ort, "felder-setzen", setID, "artist", "Neu", karteID,
		"--erwartet=[{\"karteID\":"+karteID+",\"wert\":\"Falsch\",\"gabEs\":true}]")
	if code != CodeFachlich {
		t.Fatalf("Rückgabewert %d (%s)", code, aus)
	}

	_, gelesen, _ := fahreAn(t, ort, "karten", setID, "--json")
	if strings.Contains(gelesen, `"Neu"`) {
		t.Errorf("trotz Konflikt geschrieben: %s", gelesen)
	}
}

func TestFelderSetzenMitPassenderErwartungSchreibt(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "felder-setzen", setID, "artist", "Neu", karteID,
		"--erwartet=[{\"karteID\":"+karteID+",\"gabEs\":false}]")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s %s)", code, aus, fehler)
	}

	_, gelesen, _ := fahreAn(t, ort, "karten", setID, "--json")
	if !strings.Contains(gelesen, "Neu") {
		t.Errorf("nicht geschrieben: %s", gelesen)
	}
}
