package mcp_test

// **Das Werkzeug der Einschätzung** — Werkstück A13.2a von Phase 5.3
// (#1452).

import (
	"errors"
	"strings"
	"testing"
)

// **Die Einschätzung kommt an — samt Belegen, Befunden und Vorbehalt.**
func TestDasWerkzeugKIEinschaetzungGibtBelegeUndBefunde(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.einschaetzung",
		`{"karteId": 4, "art": "regel", "alias": "gemma4:8b",`+
			` "frage": "Wie wirkt sie?"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.einschaetzung: %v", err)
	}

	for _, teil := range []string{
		`"art": "regel"`, `"belegt": false`,
		`keine amtliche Regelauskunft`,

		// **Die erfundene Nummer steht als Befund** und nicht als Quelle —
		// das ist der Satz, für den es dieses Werkzeug gibt.
		`"stelle": "702.19b"`,
		`"kennung": "B1"`, `"quelldatum": "2026-06-07"`,
		`"revision": "2026-06-07"`,
		`Nur die mitgegebenen Stellen.`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Leere Listen bleiben Listen.**
//
// **Und hier ist es die gefährlichste Verwechslung**: „keine Belege" heißt
// etwas anderes als „die Frage ging schief".
func TestEineLeereEinschaetzungGibtListenUndKeinNull(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, nilprobe{},
		"schreiben.ki.einschaetzung",
		`{"karteId": 4, "art": "regel", "alias": "gemma4:8b",`+
			` "frage": "Wie wirkt sie?"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.einschaetzung: %v", err)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("die Antwort trägt ein null:\n%s", antwort)
	}

	for _, teil := range []string{
		`"befunde": []`, `"einschraenkungen": []`, `"belege": []`,
		`"korpusstaende": []`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}

	// **Und `balance` steht gar nicht da** — die dritte Gestalt neben Liste
	// und `null`. Eine Regeleinschätzung hat diesen Abschnitt nicht; ein
	// leerer Verbund behauptete eine Suche, die keine Karte fand.
	if strings.Contains(antwort, `"balance"`) {
		t.Errorf("die Regeleinschätzung nennt balance:\n%s", antwort)
	}
}

// **Auch im Verbund bleiben leere Listen Listen.**
//
// Der Lauf darüber greift eine Ebene zu hoch: `balance` fehlt dort ganz, und
// damit betritt er den Weg nicht, auf dem aus den Listen **darin** leere
// werden.
func TestEineLeereBalanceGibtListenUndKeinNull(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, balanceprobe{},
		"schreiben.ki.einschaetzung",
		`{"karteId": 4, "art": "balance", "alias": "gemma4:8b",`+
			` "frage": "Ist sie zu stark?"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.einschaetzung: %v", err)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("die Antwort trägt ein null:\n%s", antwort)
	}

	for _, teil := range []string{
		`"vergleiche": []`, `"nichtBetrachtet": []`, `"staende": []`,
		`kein Reranker`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Ein Fehler der Werkstatt kommt als Fehler durch.**
func TestEinFehlerDerEinschaetzungKommtAlsFehlerDurch(t *testing.T) {
	t.Parallel()

	_, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
		"schreiben.ki.einschaetzung",
		`{"karteId": 4, "art": "regel", "alias": "gemma4:8b",`+
			` "frage": "Wie wirkt sie?"}`)
	if !errors.Is(err, errGestolpertMCP) {
		t.Fatalf("der Fehler ist %v", err)
	}
}

// **Und ein unlesbares Argument ebenso.**
func TestEinUnlesbaresArgumentDerEinschaetzungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.einschaetzung", `{"karteId": "vier"}`); err == nil {
		t.Fatal("das unlesbare Argument kam durch")
	}
}
