package main

// **Die Sicht auf die Sammlung an der Kommandozeile** — Werkstück
// #1500.
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. Das Fenster filtert und
// ordnet seit #1500 in Go — und was es kann, kann `mse sammlung` auch.
//
// **Der alte Vertrag bleibt**: ohne Schalter ist das erste Wort weiter die
// Druckkennung, und ohne jedes Wort kommt die ganze Sammlung. Deshalb steht
// hier beides — die Erweiterung **und** das, was sie nicht angetastet hat.

import (
	"encoding/json"
	"strings"
	"testing"
)

// dreierlei legt drei Exemplare an: zwei Nonfoil, eines Foil.
func dreierlei(t *testing.T) string {
	t.Helper()

	ort := eigenesZuhause(t)

	for _, zeile := range [][]string{
		{
			"b-druck", "1", "nearMint", "de", "nonfoil", "", "", "", "Zauber",
			"Blitzschlag", "m11", "1", "", "besitz",
		},
		{
			"a-druck", "4", "nearMint", "de", "nonfoil", "", "", "", "Zauber",
			"Ätherfluss", "m11", "2", "", "besitz",
		},
		{
			"c-druck", "2", "nearMint", "de", "foil", "", "", "", "Zauber",
			"Zauberpult", "m11", "3", "", "besitz",
		},
	} {
		if code, _, fehler := fahreAn(t, ort,
			append([]string{"exemplar"}, zeile...)...); code != CodeGut {
			t.Fatalf("anlegen %s: %d (%s)", zeile[0], code, fehler)
		}
	}

	return ort
}

// **Die Schalter filtern, ordnen und schneiden.**
func TestDieSammlungssichtFiltertOrdnetUndSchneidet(t *testing.T) {
	ort := dreierlei(t)

	// Nach Menge absteigend, ohne Foil: a-druck (4), b-druck (1).
	code, aus, fehler := fahreAn(t, ort, "sammlung", "--nonfoil", "--nach", "menge")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	zeilen := strings.Split(strings.TrimSpace(aus), "\n")
	if len(zeilen) != 2 || !strings.Contains(zeilen[0], "a-druck") {
		t.Errorf("die Ausgabe ist:\n%s", aus)
	}

	// **Der Ausschnitt schneidet** — dieselbe Sicht, je eine Zeile.
	//
	// **Zwei Stellen und nicht eine**: bei zwei Zeilen gäbe `--ab 1` allein
	// auch dann eine, wenn `--anzahl` gar nicht gelesen würde. Erst das Paar
	// unterscheidet die beiden Schalter voneinander.
	for name, fall := range map[string]struct{ ab, druck string }{
		"vorn":   {"0", "a-druck"},
		"hinten": {"1", "b-druck"},
	} {
		t.Run(name, func(t *testing.T) {
			code, aus, fehler := fahreAn(t, ort, "sammlung", "--nonfoil",
				"--nach", "menge", "--ab", fall.ab, "--anzahl", "1")
			if code != CodeGut {
				t.Fatalf("Rückgabewert %d (%s)", code, fehler)
			}

			zeilen := strings.Split(strings.TrimSpace(aus), "\n")
			if len(zeilen) != 1 || !strings.Contains(zeilen[0], fall.druck) {
				t.Errorf("der Ausschnitt ist:\n%s", aus)
			}
		})
	}
}

// **Die Suche geht über die Bezeichnung — und über die eigene Notiz.**
func TestDieSammlungssucheGehtUeberMehrereFelder(t *testing.T) {
	ort := dreierlei(t)

	for name, fall := range map[string]struct {
		suche string
		zahl  int
	}{
		"Bezeichnung": {"Blitz", 1},
		"Notiz":       {"Zauber", 3},
		"Setcode":     {"m11", 3},
		"nichts":      {"Drache", 0},
	} {
		t.Run(name, func(t *testing.T) {
			code, aus, fehler := fahreAn(t, ort, "sammlung", "--suche", fall.suche)
			if code != CodeGut {
				t.Fatalf("Rückgabewert %d (%s)", code, fehler)
			}

			gefunden := 0
			if getrimmt := strings.TrimSpace(aus); getrimmt != "" {
				gefunden = len(strings.Split(getrimmt, "\n"))
			}

			if gefunden != fall.zahl {
				t.Errorf("%d Zeilen statt %d:\n%s", gefunden, fall.zahl, aus)
			}
		})
	}
}

// **Eine Ordnung, die es nicht gibt, wird abgewiesen** — und zwar fachlich.
//
// **Anders als eine krumme Zahl**: `--ab neulich` heißt „von vorn", denn diese
// Deutung ist harmlos. Eine erfundene Ordnung ist es nicht — sie gäbe eine
// Reihenfolge, die niemand verlangt hat, und niemand sähe es.
func TestEineUnbekannteOrdnungWirdAbgewiesen(t *testing.T) {
	ort := dreierlei(t)

	code, aus, _ := fahreAn(t, ort, "sammlung", "--nach", "farbe")
	if code != CodeFachlich {
		t.Fatalf("Rückgabewert %d:\n%s", code, aus)
	}

	code, aus, fehler := fahreAn(t, ort, "sammlung", "--ab", "neulich", "--anzahl", "-2")
	if code != CodeGut {
		t.Fatalf("krumme Zahlen endeten mit %d (%s)", code, fehler)
	}

	if len(strings.Split(strings.TrimSpace(aus), "\n")) != 3 {
		t.Errorf("von vorn wären es drei Zeilen:\n%s", aus)
	}
}

// **Ohne Schalter bleibt das erste Wort die Druckkennung** — der Vertrag von
// vorher, erweitert und nicht gebrochen (§16).
func TestOhneSchalterBleibtDieDruckkennung(t *testing.T) {
	ort := dreierlei(t)

	code, aus, fehler := fahreAn(t, ort, "sammlung", "a-druck")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if zeilen := strings.Split(strings.TrimSpace(aus), "\n"); len(zeilen) != 1 ||
		!strings.Contains(zeilen[0], "a-druck") {
		t.Errorf("die Ausgabe ist:\n%s", aus)
	}
}

// **Ein Ausschnitt nennt seine Zahlen** — Befund 4 des Zweitblicks über Phase
// G. Dienst, Fenster, REST und MCP tragen seit U15
// (#2160)
// `gesamt`, `insgesamt` und die Stückzahl; die Kommandozeile warf sie weg, und
// wer einen Ausschnitt las, wusste nicht, wie viel es gibt.
func TestDerAusschnittNenntSeineZahlen(t *testing.T) {
	ort := dreierlei(t)

	code, aus, fehler := fahreAn(t, ort, "sammlung", "--anzahl", "1", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var gelesen struct {
		Exemplare []json.RawMessage `json:"exemplare"`
		Gesamt    *int              `json:"gesamt"`
		Insgesamt *int              `json:"insgesamt"`
		Stueck    *int              `json:"stueck"`
	}
	if err := json.Unmarshal([]byte(aus), &gelesen); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, aus)
	}

	if len(gelesen.Exemplare) != 1 || gelesen.Gesamt == nil || gelesen.Insgesamt == nil ||
		gelesen.Stueck == nil || *gelesen.Gesamt != 3 || *gelesen.Insgesamt != 3 ||
		*gelesen.Stueck != 7 {
		t.Errorf("erwartet ein Exemplar, gesamt 3, insgesamt 3, stueck 7 (1+4+2):\n%s", aus)
	}
}
