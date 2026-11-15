package main

// **Auch ein Fehler beim Lesen der Eingabe trägt seine Form** — Befund 3 des
// Nachlaufs über Phase 1 von 0.9.0 (F03, #2342). Der Wächter über die Wege
// fand vier Stellen, an denen der erste nicht hinsah: drei mit zwei
// Rückgabewerten und `main`, das vor dem Lauf endete.

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"testing/iotest"
)

// formAus liest den Code aus `{"fehler": …}`.
func formAus(t *testing.T, aus string) string {
	t.Helper()

	var gelesen struct {
		Fehler struct {
			Code    string `json:"code"`
			Vorgang string `json:"vorgang"`
		} `json:"fehler"`
	}

	if err := json.Unmarshal([]byte(aus), &gelesen); err != nil || len(gelesen.Fehler.Vorgang) != 8 {
		t.Fatalf("keine Form auf der Ausgabe (%v):\n%s", err, aus)
	}

	return gelesen.Fehler.Code
}

func TestEinFehlerBeimLesenDerEingabeTraegtSeineForm(t *testing.T) {
	t.Run("unlesbares JSON", func(t *testing.T) {
		code, aus, fehler := fahreMitEingabe(t, frischerBestand(t), "{kaputt", "--json", "nutzer-aendern", "1")

		if code != CodeFachlich || !strings.Contains(fehler, "die Angaben lesen") {
			t.Fatalf("Rückgabewert %d, Fehlerausgabe %q", code, fehler)
		}

		if gelesen := formAus(t, aus); gelesen != "validierung" {
			t.Errorf("Code %q, erwartet validierung", gelesen)
		}
	})

	for name, argumente := range map[string][]string{
		"das Kennwort": {"--json", "konto-neu", "probe@beispiel.invalid"},
		"die Angaben":  {"--json", "nutzer-aendern", "1"},
	} {
		t.Run("eine Eingabe, die abreißt: "+name, func(t *testing.T) {
			var raus, rfehler bytes.Buffer

			code := lauf(probekontext(t), argumente, iotest.ErrReader(errors.New("die Leitung riss")),
				&raus, &rfehler, frischerBestand(t))

			if code != CodeFachlich || !strings.Contains(rfehler.String(), "die Standardeingabe lesen") {
				t.Fatalf("Rückgabewert %d, Fehlerausgabe %q", code, rfehler.String())
			}

			formAus(t, raus.String())
		})
	}
}

// **Ein Bestandsort, der sich nicht bestimmen lässt**, endet vor dem Lauf —
// und mit `--json` trotzdem mit der Form.
func TestOhneBestandsortTraegtDerFehlerSeineForm(t *testing.T) {
	t.Parallel()

	var raus, rfehler bytes.Buffer

	code := ohneBestandsort([]string{"--json", "sets"}, &raus, &rfehler, errors.New("MSE_BESTAND zeigt ins Leere"))

	if code != CodeFachlich || !strings.Contains(rfehler.String(), "MSE_BESTAND zeigt ins Leere") {
		t.Fatalf("Rückgabewert %d, Fehlerausgabe %q", code, rfehler.String())
	}

	formAus(t, raus.String())

	raus.Reset()

	if ohneBestandsort([]string{"sets"}, &raus, &rfehler, errors.New("x")) != CodeFachlich || raus.Len() != 0 {
		t.Errorf("ohne --json steht etwas auf der Ausgabe: %q", raus.String())
	}
}
