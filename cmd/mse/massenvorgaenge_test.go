package main

// **Massenvorgänge an der Kommandozeile** — Werkstück 6 von Phase 4.5
// (#1292).

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

// dreiExemplare legt drei an und gibt Ort und Zeilen als `id:fassung`.
func dreiExemplare(t *testing.T) (ort string, zeilen []string) {
	t.Helper()

	ort = eigenesZuhause(t)

	for zaehler, zustand := range []string{"nearMint", "played", "good"} {
		code, aus, fehler := fahreAn(t, ort, "exemplar",
			"druck-"+strconv.Itoa(zaehler), "1", zustand, "en", "nonfoil",
			"", "", "", "", "", "", "", "Kiste 4", "besitz", "--json")
		if code != CodeGut {
			t.Fatalf("anlegen: %d (%s)", code, fehler)
		}

		var angelegt struct {
			ID int64 `json:"id"`
		}

		if err := json.Unmarshal([]byte(aus), &angelegt); err != nil {
			t.Fatalf("die Antwort lesen: %v (%s)", err, aus)
		}

		zeilen = append(zeilen, strconv.FormatInt(angelegt.ID, 10)+":1")
	}

	return ort, zeilen
}

// **Ein Lagerort für drei Zeilen auf einmal.**
func TestEinLagerortFuerDreiZeilenAnDerKommandozeile(t *testing.T) {
	ort, zeilen := dreiExemplare(t)

	code, aus, fehler := fahreAn(t, ort,
		append([]string{"masse-aendern", "lagerort=Ordner Rot"}, zeilen...)...)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "3 geändert") {
		t.Errorf("die Ausgabe lautet %q", aus)
	}

	_, sammlung, _ := fahreAn(t, ort, "sammlung", "--json")
	if strings.Contains(sammlung, "Kiste 4") {
		t.Errorf("der alte Lagerort steht noch da: %s", sammlung)
	}
}

// **Jedes Feld geht durch — und `wunsch` nimmt ja oder nein.**
func TestJedesMassenfeldGehtDurch(t *testing.T) {
	ort, zeilen := dreiExemplare(t)

	code, _, fehler := fahreAn(t, ort, "masse-aendern",
		"menge=7", "zustand=poor", "lagerort=Kiste 9", "wunsch=ja", zeilen[0])
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	_, sammlung, _ := fahreAn(t, ort, "sammlung", "--json")

	for _, teil := range []string{
		`"menge": 7`, `"zustand": "poor"`,
		`"lagerort": "Kiste 9"`, `"wunsch": true`,
	} {
		if !strings.Contains(sammlung, teil) {
			t.Errorf("%q fehlt:\n%s", teil, sammlung)
		}
	}
}

// **Eine veraltete Fassung wird übergangen und benannt** (§7, Regel 2).
func TestEineVeralteteFassungAnDerKommandozeile(t *testing.T) {
	ort, zeilen := dreiExemplare(t)

	zeilen[1] = strings.Split(zeilen[1], ":")[0] + ":99"

	code, aus, fehler := fahreAn(t, ort,
		append([]string{"masse-aendern", "lagerort=Kiste 9"}, zeilen...)...)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "2 geändert") || !strings.Contains(aus, "!") {
		t.Errorf("die Ausgabe lautet %q", aus)
	}
}

// **Löschen in Masse.**
func TestLoeschenInMasseAnDerKommandozeile(t *testing.T) {
	ort, zeilen := dreiExemplare(t)

	code, aus, fehler := fahreAn(t, ort,
		append([]string{"masse-loeschen"}, zeilen...)...)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "3 geändert") {
		t.Errorf("die Ausgabe lautet %q", aus)
	}
}

// **Die Bedienfehler eines Massenvorgangs.**
func TestDieBedienfehlerEinesMassenvorgangs(t *testing.T) {
	ort := eigenesZuhause(t)

	for _, fall := range []struct {
		name      string
		argumente []string
	}{
		{name: "ohne Zeilen", argumente: []string{"masse-aendern", "menge=3"}},
		{name: "Zeile ohne Fassung", argumente: []string{"masse-aendern", "menge=3", "12"}},
		{name: "Zeile ist keine Zahl", argumente: []string{"masse-aendern", "menge=3", "x:1"}},
		{name: "Menge ist keine Zahl", argumente: []string{"masse-aendern", "menge=viele", "1:1"}},
		{name: "unbekanntes Feld", argumente: []string{"masse-aendern", "farbe=rot", "1:1"}},
		{name: "wunsch ist weder ja noch nein", argumente: []string{"masse-aendern", "wunsch=vielleicht", "1:1"}},
		{name: "Löschen ohne Zeilen", argumente: []string{"masse-loeschen"}},
		{name: "Stunden sind keine Zahl", argumente: []string{"preise-alle", "viele"}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, fall.argumente...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
			}
		})
	}
}

// **Eine Änderung, die nichts ändert, ist ein fachliches Nein.**
func TestEineMassenaenderungOhneAenderung(t *testing.T) {
	ort, zeilen := dreiExemplare(t)

	if code, _, _ := fahreAn(t, ort,
		append([]string{"masse-aendern"}, zeilen...)...); code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeFachlich)
	}
}

// **Als JSON geht es auch.**
func TestMassenvorgaengeAlsJSON(t *testing.T) {
	ort, zeilen := dreiExemplare(t)

	_, aus, _ := fahreAn(t, ort, "masse-aendern", "menge=2", zeilen[0], "--json")
	if !strings.Contains(aus, `"geaendert": 1`) {
		t.Errorf("die JSON-Ausgabe lautet %s", aus)
	}
}
