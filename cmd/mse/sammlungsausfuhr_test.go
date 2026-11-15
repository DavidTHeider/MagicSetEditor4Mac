package main

// **Die Sammlung ausgeben und zurücklesen** — Werkstück 4 von Phase 4.5
// (#1292).

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// mitExemplar legt ein Exemplar mit jeder Angabe an.
func mitExemplar(t *testing.T) string {
	t.Helper()

	ort := eigenesZuhause(t)

	if code, _, fehler := fahreAn(t, ort, "exemplar", "probe-druck", "4",
		"played", "de", "foil", "2026-08-30", "4.17", "EUR", "getauscht, 2026",
		"Lightning Bolt", "m11", "149", "Kiste 4", "besitz"); code != CodeGut {
		t.Fatalf("anlegen: %d (%s)", code, fehler)
	}

	return ort
}

// **Das CSV trägt jede Angabe.**
func TestDasCSVTraegtJedeAngabe(t *testing.T) {
	ort := mitExemplar(t)

	code, aus, fehler := fahreAn(t, ort, "sammlung-ausgeben", "csv")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	for _, teil := range []string{
		"printing", "probe-druck", "played", "de", "true",
		"2026-08-30", "4.17", "EUR", "getauscht, 2026", "Kiste 4",
	} {
		if !strings.Contains(aus, teil) {
			t.Errorf("die Ausgabe nennt %q nicht:\n%s", teil, aus)
		}
	}
}

// **Die Liste ist einfügbar — und trägt den Zustand nicht.**
func TestDieListeIstEinfuegbar(t *testing.T) {
	code, aus, fehler := fahreAn(t, mitExemplar(t), "sammlung-ausgeben", "liste")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "4 Lightning Bolt (M11) 149 *F*") {
		t.Errorf("die Zeile fehlt:\n%s", aus)
	}

	// **Das ist die Zusage, nicht ein Versehen** — wer sichert, nimmt CSV.
	if strings.Contains(aus, "played") {
		t.Errorf("die Liste trägt den Zustand:\n%s", aus)
	}
}

// **Eine Form, die es nicht gibt, ist ein fachliches Nein.**
func TestEineErfundeneAusfuhrform(t *testing.T) {
	if code, _, _ := fahreAn(t, eigenesZuhause(t),
		"sammlung-ausgeben", "pergament"); code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeFachlich)
	}
}

// **Der Rundlauf: ausgeben, in einen frischen Bestand einlesen, alles da.**
func TestDerRundlaufUeberDieKommandozeile(t *testing.T) {
	_, text, _ := fahreAn(t, mitExemplar(t), "sammlung-ausgeben", "csv")

	frisch := eigenesZuhause(t)

	code, aus, fehler := fahreAnMitEingabe(t, frisch, text, "sammlung-einlesen")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "1 angelegt") {
		t.Errorf("die Ausgabe lautet %q", aus)
	}

	_, sammlung, _ := fahreAn(t, frisch, "sammlung", "--json")

	for _, teil := range []string{
		"played", "getauscht, 2026", "4.17", "Kiste 4",
	} {
		if !strings.Contains(sammlung, teil) {
			t.Errorf("nach dem Rundlauf fehlt %q:\n%s", teil, sammlung)
		}
	}
}

// **Als JSON geht beides auch.**
func TestDieAusfuhrAlsJSON(t *testing.T) {
	ort := mitExemplar(t)

	_, aus, _ := fahreAn(t, ort, "sammlung-ausgeben", "csv", "--json")
	if !strings.Contains(aus, `"form": "csv"`) {
		t.Errorf("die JSON-Ausgabe nennt die Form nicht: %s", aus)
	}

	_, eingelesen, _ := fahreAnMitEingabe(t, eigenesZuhause(t),
		"printing,menge\nabc,2\n", "sammlung-einlesen", "--json")
	if !strings.Contains(eingelesen, `"stueck": 2`) {
		t.Errorf("die JSON-Ausgabe nennt die Stückzahl nicht: %s", eingelesen)
	}
}

// **Ein kaputtes CSV wird zum Befund, nicht zum Abbruch** (§8).
func TestEinKaputtesCSVAnDerKommandozeile(t *testing.T) {
	code, aus, _ := fahreAnMitEingabe(t, eigenesZuhause(t),
		"printing,menge\nabc,1\ndef,keineZahl\n", "sammlung-einlesen")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d", code)
	}

	if !strings.Contains(aus, "1 angelegt") || !strings.Contains(aus, "!") {
		t.Errorf("die Ausgabe lautet %q", aus)
	}
}

// **Eine Standardeingabe, die klemmt, ist ein Bedienfehler.**
func TestEineKlemmendeEingabe(t *testing.T) {
	var raus, rfehler bytes.Buffer

	code := lauf(probekontext(t), []string{"sammlung-einlesen"}, klemmendeQuelle{},
		&raus, &rfehler, eigenesZuhause(t))

	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// klemmendeQuelle scheitert beim Lesen.
//
// **Der Zweig bliebe sonst unbetreten**: eine Standardeingabe, die klemmt,
// gibt es in keinem Lauf mit einer Zeichenkette.
type klemmendeQuelle struct{}

func (klemmendeQuelle) Read([]byte) (int, error) {
	return 0, errors.New("die Leitung ist zu")
}
