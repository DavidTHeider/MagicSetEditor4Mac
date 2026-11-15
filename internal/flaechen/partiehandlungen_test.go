package flaechen_test

// **Dieselbe Handlungsauskunft an jeder Fläche** — P3
// (#2371),
// Abnahme 1: je Objekt dieselben Handlungen, dieselbe Primär- und
// Zweithandlung, und keine Auskunft über eine fremde Hand. Das Fenster liest
// dieselbe Antwort der Naht wie die Kommandozeile.

import (
	"net/http"
	"strings"
	"testing"
)

func TestDieHandlungenSindAnJederFlaecheGleich(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	partieID := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck).ID

	ausCLI := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	ausMCP := bildAusFlach(t, mussMCP(t, umgebung, "lesen.partie",
		map[string]any{"partieID": partieID, "fuer": "anna"}))

	ausREST := bildAusFlach(t, mussREST(t, an, http.MethodGet,
		"/partien/"+partieID+"?fuer=anna", ""))

	gleichesBild(t, "die Handlungen derselben Partie", ausCLI, ausMCP)
	gleichesBild(t, "die Handlungen derselben Partie", ausCLI, ausREST)

	if len(ausCLI.Handlungen) == 0 {
		t.Fatal("nach dem Beginn steht keine Handlung da")
	}

	zonen := map[string]bildobjekt{}
	for _, objekt := range ausCLI.Objekte {
		zonen[objekt.Kennung] = objekt
	}

	primaer := ""
	for _, eine := range ausCLI.Handlungen {
		objekt := zonen[eine.Objekt]
		if objekt.Beherrscher != "anna" && objekt.Zone == "hand" {
			t.Errorf("zur fremden Handkarte %q steht %+v", eine.Objekt, eine)
		}

		if eine.Primaer != "" && primaer == "" {
			primaer = eine.Objekt + ": " + eine.Primaer
		}
	}

	// **Und die Textform nennt, was ein Klick täte** — dieselbe Auskunft
	// ohne `--json`.
	text := mussKommandozeile(t, umgebung, "partie-zeigen", partieID, "anna")
	if primaer == "" || !strings.Contains(text, primaer) {
		t.Errorf("die Textform nennt %q nicht:\n%s", primaer, text)
	}
}
