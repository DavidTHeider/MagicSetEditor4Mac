package main

// **Die Sitzung an der Kommandozeile** — ID-02
// (#1701).

import (
	"strings"
	"testing"
)

func TestDieSitzungsbefehleWeisenFalscheAufrufeAb(t *testing.T) {
	ort, _, _ := mitKarte(t)

	for name, fall := range map[string]struct {
		eingabe   string
		argumente []string
	}{
		"sitzung mit einem Wort":                    {"", []string{"sitzung", "zuviel"}},
		"kennwort-wechseln mit einem Wort":          {"a\nb", []string{"kennwort-wechseln", "geheim"}},
		"kennwort-wechseln mit einer Zeile":         {"nur-eine", []string{"kennwort-wechseln"}},
		"kennwort-wiederherstellen ohne Nutzer":     {"a\nb", []string{"kennwort-wiederherstellen"}},
		"kennwort-wiederherstellen mit drei Zeilen": {"a\nb\nc", []string{"kennwort-wiederherstellen", "1"}},
	} {
		t.Run(name, func(t *testing.T) {
			code, _, fehler := fahreAnMitEingabe(t, ort, fall.eingabe, fall.argumente...)
			if code != CodeBedienung {
				t.Errorf("Rückgabewert %d (%s)", code, fehler)
			}
		})
	}

	// **Ohne Standardeingabe sagt es, woher die Kennwörter kommen.**
	if code, _, fehler := fahreAn(t, ort, "kennwort-wechseln"); code != CodeBedienung || !strings.Contains(fehler, "Standardeingabe") {
		t.Errorf("ohne Eingabe: %d %q", code, fehler)
	}
}

// **Stand, Wechsel und ein falscher Zettel** — der ganze Weg an einem
// frischen Bestand.
func TestKennwortWechselnUndDerStandDerSitzung(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, aus, fehler := fahreAn(t, ort, "sitzung"); code != CodeGut || !strings.HasPrefix(aus, "angemeldet\t") {
		t.Fatalf("Stand: %d %q %q", code, aus, fehler)
	}

	if code, _, fehler := fahreAnMitEingabe(t, ort, "\nkurz", "kennwort-wechseln"); code != CodeFachlich {
		t.Errorf("zu kurz: %d %q", code, fehler)
	}

	if code, aus, fehler := fahreAnMitEingabe(t, ort, "\nein-langes-kennwort", "kennwort-wechseln"); code != CodeGut || aus != "gesetzt\n" {
		t.Fatalf("gesetzt: %d %q %q", code, aus, fehler)
	}

	// **Jetzt verlangt der Wechsel das bisherige.**
	if code, _, fehler := fahreAnMitEingabe(t, ort, "falsch\nnoch-ein-langes", "kennwort-wechseln"); code != CodeFachlich {
		t.Errorf("ohne das bisherige: %d %q", code, fehler)
	}

	if code, _, fehler := fahreAnMitEingabe(t, ort, "code\nnoch-ein-langes", "kennwort-wiederherstellen", "1"); code != CodeFachlich {
		t.Errorf("angemeldet gibt es nichts wiederherzustellen: %d %q", code, fehler)
	}

	if code, _, fehler := fahreAnMitEingabe(t, ort, "code\nnoch-ein-langes", "kennwort-wiederherstellen", "x"); code != CodeFachlich {
		t.Errorf("keine Kennung: %d %q", code, fehler)
	}

	if code, aus, _ := fahreAn(t, ort, "sitzung", "--json"); code != CodeGut || !strings.Contains(aus, `"zustand": "angemeldet"`) {
		t.Errorf("als JSON: %d %q", code, aus)
	}
}
