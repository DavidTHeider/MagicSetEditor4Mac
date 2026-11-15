package mcp_test

// **Das Werkzeug der Kartenerkennung** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass die Antwort zwei Dinge getrennt hält**: was das Modell **gelesen**
// hat, und welche Drucke eine **Suche** dazu gefunden hat. Für ein
// Sprachmodell ist das der ganze Unterschied — es darf den Kandidaten nicht
// für seine eigene Auskunft halten.
//
// **Und dass keine Liste als `null` ankommt.** `null` sieht dort aus wie
// „konnte nicht beantwortet werden"; „nichts gelesen" ist etwas anderes.

import (
	"errors"
	"strings"
	"testing"
)

// **Gelesene Stellen und gefundene Drucke kommen beide an.**
func TestDasWerkzeugKIErkennungTrenntLesenUndFinden(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.erkennung",
		`{"pfad": "/tmp/karte.png", "anbieter": "ollama", "alias": "sieht:8b"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.erkennung: %v", err)
	}

	for _, teil := range []string{
		`"beobachtungen"`, `"art": "name"`, `"lesbar": true`,
		`"kandidaten"`, `"kennung": "lea-174"`, `"gestuetzt"`,
		`"drucke": 1`, `"eindeutig": true`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}

	// **Und es gibt keine Prozentangabe** — die Unsicherheit ist die Zahl
	// der passenden Drucke ([0110](../../docs/entscheidungen/0110-eine-bilderkennung-nennt-keinen-druck-den-die-suche-nicht-faende.md)).
	for _, verboten := range []string{"sicherheit", "konfidenz", "prozent"} {
		if strings.Contains(strings.ToLower(antwort), verboten) {
			t.Errorf("die Antwort trägt %q:\n%s", verboten, antwort)
		}
	}
}

// **Der Anbieter kommt an, statt fest zu stehen.**
//
// **Die Gegenprobe zum Lauf darüber**: ein Werkzeug, das die Argumente
// verwirft und eine feste Antwort gäbe, wäre dort grün.
func TestDasWerkzeugKIErkennungReichtSeinenBezugDurch(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.erkennung",
		`{"pfad": "/tmp/karte.png", "anbieter": "anthropic", "alias": "sieht-gross"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.erkennung: %v", err)
	}

	if !strings.Contains(antwort, `"anbieter": "anthropic"`) ||
		!strings.Contains(antwort, `"modell": "sieht-gross"`) {
		t.Errorf("der Bezug kam nicht durch:\n%s", antwort)
	}
}

// **Leere Listen bleiben Listen.**
func TestEineLeereErkennungGibtListenUndKeinNull(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, nilprobe{},
		"schreiben.ki.erkennung",
		`{"pfad": "/tmp/karte.png", "anbieter": "ollama", "alias": "sieht:8b"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.erkennung: %v", err)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("die Antwort trägt ein null:\n%s", antwort)
	}

	for _, liste := range []string{
		`"beobachtungen": []`, `"nichtGelesen": []`, `"befunde": []`,
		`"kandidaten": []`,
	} {
		if !strings.Contains(antwort, liste) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", liste, antwort)
		}
	}
}

// **Und die Liste *innerhalb* eines Kandidaten ebenso.**
//
// **Der Lauf darüber greift eine Ebene zu hoch**: er prüft `kandidaten`, und
// `gestuetzt` **darin** bleibt dabei unberührt. Dieselbe Falle wie bei der
// Gruppe (A13.1b) und bei der Balance (A13.2b) — dort ist sie zweimal
// aufgetreten.
func TestEinKandidatOhneStuetzeGibtEineListeUndKeinNull(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, ohneStuetze{},
		"schreiben.ki.erkennung",
		`{"pfad": "/tmp/karte.png", "anbieter": "ollama", "alias": "sieht:8b"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.erkennung: %v", err)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("die Antwort trägt ein null:\n%s", antwort)
	}

	if !strings.Contains(antwort, `"gestuetzt": []`) {
		t.Errorf("gestuetzt steht nicht als Liste da:\n%s", antwort)
	}
}

// **Ein Fehler der Werkstatt kommt als Fehler durch.**
func TestEinFehlerDerErkennungKommtAlsFehlerDurch(t *testing.T) {
	t.Parallel()

	_, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
		"schreiben.ki.erkennung",
		`{"pfad": "/tmp/karte.png", "anbieter": "ollama", "alias": "sieht:8b"}`)
	if !errors.Is(err, errGestolpertMCP) {
		t.Fatalf("der Fehler ist %v", err)
	}
}

// **Und ein kaputtes Argument ebenso.**
func TestEinKaputtesErkennungsargumentWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.erkennung", `{"pfad": 17}`); err == nil {
		t.Fatal("das kaputte Argument kam durch")
	}
}
