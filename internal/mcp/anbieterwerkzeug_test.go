package mcp_test

// **Die Werkzeuge für Kontostand und Fremdverlauf** — Werkstück A5 von
// Phase 5.3 (#1444).

import (
	"errors"
	"strings"
	"testing"
)

// **Der Kontostand kommt an — samt der Frist, die ihn relativiert.**
func TestDasWerkzeugAnbieterstandNenntDieFrist(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"lesen.anbieterstand", `{"dienst":"scrydex"}`)
	if err != nil {
		t.Fatalf("lesen.anbieterstand: %v", err)
	}

	for _, teil := range []string{
		`"dienst": "scrydex"`, `"gesamt": 100000`, `"rest": 84213`,
		`"verbraucht": 15787`, `"ueberzugssatz": "0.0004"`,
		`"hoechstensAltMinuten": 30`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Der Fremdverlauf gibt seine Bilanz — und leere Befunde bleiben eine Liste.**
func TestDasWerkzeugFremdverlaufGibtSeineBilanz(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.fremdverlauf", `{"quelle":"scrydex","printing":"probe-druck"}`)
	if err != nil {
		t.Fatalf("schreiben.fremdverlauf: %v", err)
	}

	for _, teil := range []string{
		`"quelle": "scrydex"`, `"fassung": "mtg-m11-149"`,
		`"zugeordnet": true`, `"staende": 4`, `"befunde": []`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Ein Fehler der Naht kommt durch — bei beiden.**
func TestEinFehlerDerAnbieternaehteKommtDurch(t *testing.T) {
	t.Parallel()

	for name, argumente := range map[string]string{
		"lesen.anbieterstand":    `{"dienst":"scrydex"}`,
		"schreiben.fremdverlauf": `{"quelle":"scrydex","printing":"probe-druck"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
				name, argumente); !errors.Is(err, errGestolpertMCP) {
				t.Errorf("der Fehler ist %v", err)
			}
		})
	}
}

// **Und ein kaputtes Argument ebenso.**
func TestEinKaputtesAnbieterargumentWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, argumente := range map[string]string{
		"lesen.anbieterstand":    `{"dienst":42}`,
		"schreiben.fremdverlauf": `{"quelle":42}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
				name, argumente); err == nil {
				t.Error("eine Zahl kam durch")
			}
		})
	}
}
