package mcp_test

// **Das Werkzeug für die Marktangebote** — Werkstück A4 von Phase 5.3
// (#1443).

import (
	"errors"
	"strings"
	"testing"
)

// **Ein Abruf gibt die Posten samt Zuordnung.**
func TestDasWerkzeugMarktangeboteGibtSeinePosten(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"lesen.marktangebote",
		`{"quelle":"cardtrader","printing":"probe-druck"}`)
	if err != nil {
		t.Fatalf("lesen.marktangebote: %v", err)
	}

	for _, teil := range []string{
		`"quelle": "cardtrader"`, `"fassung": "89"`, `"zugeordnet": true`,
		`"kennung": "101862104"`, `"wert": "2.17"`, `"waehrung": "EUR"`,
		`"menge": 3`, `"zustand": "Moderately Played"`, `"sprache": "en"`,
		`"verkaeufer": "einer"`, `"land": "FI"`, `"befunde": []`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Eine leere Antwort bleibt eine Liste und wird kein `null`.**
//
// Ein `null` im JSON sähe für ein Sprachmodell aus wie „konnte nicht
// beantwortet werden" — und genau das ist es nicht: dort steht gerade nichts
// zum Verkauf.
func TestEineLeereMarktantwortBleibtEineListe(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, nilprobe{},
		"lesen.marktangebote",
		`{"quelle":"cardtrader","printing":"probe-druck"}`)
	if err != nil {
		t.Fatalf("lesen.marktangebote: %v", err)
	}

	for _, teil := range []string{`"angebote": []`, `"befunde": []`} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestEinFehlerDesMarktabrufsKommtDurchsWerkzeug(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
		"lesen.marktangebote",
		`{"quelle":"cardtrader","printing":"probe-druck"}`); !errors.Is(
		err, errGestolpertMCP) {
		t.Errorf("der Fehler ist %v", err)
	}
}

// **Und ein kaputtes Argument ebenso.**
func TestEinKaputtesMarktargumentWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"lesen.marktangebote", `{"quelle":42}`); err == nil {
		t.Error("eine Zahl als Quelle kam durch")
	}
}
