package main

import (
	"strings"
	"testing"
)

// **Eine Klage endet mit ihrer Zeile** — ENG-05.AK-03
// (#1730).
//
// Gemessen am 2026-09-23: ein zweiter Nutzer gleichen Namens endete mit
// `…liegt schon daVorgang: dad7dc7a` — 35 Stellen gaben die Meldung ohne
// Zeilenende aus, und die Vorgangskennung klebte an ihr. Wer den Satz las,
// las die Kennung mit, und wer die Kennung suchte, fand sie nicht am
// Zeilenanfang.
func TestEineKlageEndetMitIhrerZeile(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "nutzer-neu", "Zweimal"); code != CodeGut {
		t.Fatalf("der erste: %d (%s)", code, fehler)
	}

	code, _, fehler := fahreAn(t, ort, "nutzer-neu", "Zweimal")
	if code != CodeFachlich {
		t.Fatalf("Rückgabewert %d, erwartet %d (%s)", code, CodeFachlich, fehler)
	}

	if !strings.Contains(fehler, "\nVorgang: ") || !strings.Contains(fehler, "gibt es hier schon") {
		t.Errorf("die Klage: %q", fehler)
	}
}
