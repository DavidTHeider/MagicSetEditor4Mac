package main

// **Anmeldung an der Kommandozeile** — Werkstück 4 von Phase 4.12
// (#1351).

import (
	"strings"
	"testing"
)

// **`mse anbieter` nennt die vier — ohne Kontakt nach draußen.**
func TestAnbieterNenntDieVier(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "anbieter")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	for _, name := range []string{"github", "google", "microsoft", "x"} {
		if !strings.Contains(aus, name) {
			t.Errorf("%q fehlt in %q", name, aus)
		}
	}
}

// **`mse anmelden` ohne eingetragene Kennung ist ein fachliches Nein.**
//
// **Und die Meldung sagt, was zu tun ist** (§6): dieses Programm liefert keine
// Zugangsdaten mit.
func TestAnmeldenOhneKennungIstEinFachlichesNein(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, _, fehler := fahreAn(t, ort, "anmelden", "github")
	if code != CodeFachlich {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(fehler, "eigene Anwendung") {
		t.Errorf("die Meldung sagt nicht, was zu tun ist: %q", fehler)
	}
}

// **`mse konten` nennt die Konten — und ohne Konten ist die Liste leer.**
func TestKontenNenntDieKonten(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "konten", "1")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if strings.TrimSpace(aus) != "" {
		t.Errorf("die Liste ist nicht leer: %q", aus)
	}
}

// **Ein Konto, das es nicht gibt, zu lösen ist ein fachliches Nein.**
func TestEinFehlendesKontoZuLoesenIstEinNein(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "konten", "1", "github"); code != CodeFachlich {
		t.Errorf("Rückgabewert %d (%s), erwartet war CodeFachlich", code, fehler)
	}
}

// **Jeder der drei Befehle weist eine falsche Wortzahl ab.**
func TestDieAnmeldebefehleWeisenFalscheWortzahlenAb(t *testing.T) {
	ort, _, _ := mitKarte(t)

	faelle := map[string][]string{
		"anbieter mit Wort":  {"anbieter", "zuviel"},
		"anmelden ohne Wort": {"anmelden"},
		"anmelden mit zwei":  {"anmelden", "github", "zuviel"},
		"konten ohne Wort":   {"konten"},
		"konten mit drei":    {"konten", "1", "github", "zuviel"},
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, argumente...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet war CodeBedienung", code)
			}
		})
	}
}

// **Und `--json` gibt jede Antwort auch als JSON.**
func TestDieAnmeldebefehleGebenJSON(t *testing.T) {
	ort, _, _ := mitKarte(t)

	for _, argumente := range [][]string{{"anbieter"}, {"konten", "1"}} {
		code, aus, fehler := fahreAn(t, ort, append(argumente, "--json")...)
		if code != CodeGut {
			t.Fatalf("%v: %d (%s)", argumente, code, fehler)
		}

		if !strings.HasPrefix(strings.TrimSpace(aus), "{") {
			t.Errorf("%v gab kein JSON: %q", argumente, aus)
		}
	}
}
