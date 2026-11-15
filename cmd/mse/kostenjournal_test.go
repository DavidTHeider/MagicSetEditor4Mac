package main

// **Das Kostenjournal auf der Kommandozeile** — Werkstück O12 von Phase 5.2
// (#1439).

import (
	"strings"
	"testing"
)

func TestDasKostenkontoNenntVierZahlen(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "kosten", "anbieter")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	// **Ohne Budget steht überall null — und keine Währung.** Es gibt keine,
	// solange niemand eine freigegeben hat; eine erfundene wäre schlimmer.
	for _, erwartet := range []string{
		"konto\tanbieter", "zeitraum\ttag", "grenze\t0",
		"reserviert\t0", "verbraucht\t0", "unklar\t0", "verfuegbar\t0",
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("die Ausgabe nennt %q nicht:\n%s", erwartet, aus)
		}
	}

	// **Und als JSON dasselbe.**
	code, aus, _ = fahreAn(t, ort, "kosten", "anbieter", "--json")
	if code != CodeGut || !strings.Contains(aus, `"verfuegbar"`) {
		t.Errorf("als JSON: %d %s", code, aus)
	}
}

func TestDasBudgetLaesstSichSetzenUndWirktSofort(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "kostenbudget", "anbieter", "monat", "50000", "EUR")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "gesetzt\tanbieter\tmonat\t50000\tEUR") {
		t.Errorf("die Ausgabe ist %q", aus)
	}

	code, aus, _ = fahreAn(t, ort, "kosten", "anbieter", "monat")
	if code != CodeGut || !strings.Contains(aus, "verfuegbar\t50000\tEUR") {
		t.Errorf("nach dem Setzen: %d\n%s", code, aus)
	}

	// **Die Posten sind leer, und das ist eine Antwort.**
	code, aus, _ = fahreAn(t, ort, "kostenposten", "anbieter", "monat")
	if code != CodeGut || strings.TrimSpace(aus) != "" {
		t.Errorf("die Posten: %d %q", code, aus)
	}

	code, aus, _ = fahreAn(t, ort, "kostenposten", "anbieter", "--json")
	if code != CodeGut || !strings.Contains(aus, `"posten": []`) {
		t.Errorf("die Posten als JSON: %d %s", code, aus)
	}

	code, aus, _ = fahreAn(t, ort, "kostenbudget", "anbieter", "monat", "0", "EUR", "--json")
	if code != CodeGut || !strings.Contains(aus, `"fehler": null`) {
		t.Errorf("wegnehmen als JSON: %d %s", code, aus)
	}
}

func TestEinKrummerKostenaufrufWirdAbgewiesen(t *testing.T) {
	ort, _, _ := mitKarte(t)

	for name, argumente := range map[string][]string{
		"kosten ohne Konto":      {"kosten"},
		"kosten mit zu vielen":   {"kosten", "a", "b", "c"},
		"posten ohne Konto":      {"kostenposten"},
		"posten mit zu vielen":   {"kostenposten", "a", "b", "c"},
		"budget mit zu wenigen":  {"kostenbudget", "anbieter", "tag"},
		"budget ohne Zahl":       {"kostenbudget", "anbieter", "tag", "viel", "EUR"},
		"kosten mit Zeitraum hm": {"kosten", "anbieter", "stunde"},
		"budget mit Zeitraum hm": {"kostenbudget", "anbieter", "stunde", "1", "EUR"},
		"budget ohne Währung":    {"kostenbudget", "anbieter", "tag", "1", ""},
		"posten mit Zeitraum hm": {"kostenposten", "anbieter", "stunde"},
	} {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, argumente...); code == CodeGut {
				t.Error("kam durch")
			}
		})
	}
}
