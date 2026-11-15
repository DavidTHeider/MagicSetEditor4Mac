package app

// **Das Deck aus einem fremden Umschlag** — QA-06c
// (#1976),
// Fläche 29 aus `docs/FUZZZIELE.md` (Umschlaginhalt).
//
// # Was es neben `FuzzUmschlag` (27/28) zu sehen gibt
//
// **Was nach der Prüfung des Umschlags kommt.** `FuzzUmschlag` hält Hülle,
// Manifest und Größen; der Inhalt eines Objekts ist dort Bytes. Hier wird er
// gelesen, wie `UmschlagDeck` ihn liest — `json.Unmarshal` in ein [DeckDTO]
// —, und angenommen, mit echtem Bestand.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - eine Antwort ohne Deck **und** ohne Fehler
//   - **eine eigene Karte des Absenders liegt im Deck** (FMT-04.AK-04)
//   - ein Platz mit Menge null oder weniger
//   - **ohne Fehler fehlt ein Platz** — jede Menge der Eingabe, die keine
//     eigene Karte des Absenders ist, liegt nach Bereich gezählt im Deck
//   - **ein Platz fehlt still** — er liegt im Deck oder die Meldung nennt
//     seinen Druck. Dazugekommen mit
//     #1995:
//     die erste Fassung prüfte nur den Weg **ohne** Fehler, und gerade dort
//     verschwanden die Plätze — beim ersten gescheiterten brach die Annahme ab
//
//	go test ./internal/app -run Fuzz -fuzz FuzzUmschlagdeck -fuzztime 2m -fuzzminimizetime 5s

import (
	"encoding/json"
	"strings"
	"testing"
)

func FuzzUmschlagdeck(f *testing.F) {
	for _, saat := range []string{
		`{"name":"Rot","spiel":"probe","plaetze":[{"printing":"11111111-1111-4111-8111-111111111111","menge":4,"bereich":"hauptdeck"}]}`,
		`{"name":"Zwei","spiel":"probe","plaetze":[{"printing":"11111111-1111-4111-8111-111111111111","menge":2,"bereich":"hauptdeck"},{"printing":"11111111-1111-4111-8111-111111111111","menge":1,"bereich":"sideboard"}]}`,
		`{"name":"Fremd","spiel":"probe","plaetze":[{"karteID":"7","bezeichnung":"Glutwind","menge":2,"bereich":"hauptdeck"}]}`,
		`{"name":"Schief","spiel":"probe","plaetze":[{"printing":"11111111-1111-4111-8111-111111111111","menge":-3,"bereich":"irgendwo"},{"printing":"22222222-2222-4222-8222-222222222222","menge":1,"bereich":"hauptdeck"}]}`,
		`{"name":"","spiel":"","plaetze":null}`,
		`{"name":"Groß","spiel":"probe","plaetze":[{"printing":"11111111-1111-4111-8111-111111111111","menge":2147483647,"bereich":"hauptdeck"}]}`,
	} {
		f.Add([]byte(saat))
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		var gelesen DeckDTO
		if json.Unmarshal(roh, &gelesen) != nil {
			return
		}

		anwendung := mitIdentitaet(t)
		antwort := anwendung.deckAnnehmen(t.Context(), gelesen)

		if antwort.Deck == nil && antwort.Fehler == nil {
			t.Fatal("eine Antwort ohne Deck und ohne Fehler")
		}

		if antwort.Deck == nil {
			return
		}

		gelegt := map[string]int{}

		for _, platz := range antwort.Deck.Plaetze {
			if platz.KarteID != "" {
				t.Errorf("eine eigene Karte des Absenders liegt im Deck: %+v", platz)
			}

			if platz.Menge <= 0 {
				t.Errorf("ein Platz mit Menge %d: %+v", platz.Menge, platz)
			}

			gelegt[platz.Printing+"|"+platz.Bereich] += platz.Menge
		}

		for _, platz := range gelesen.Plaetze {
			if platz.KarteID != "" || gelegt[platz.Printing+"|"+platz.Bereich] > 0 {
				continue
			}

			if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, platz.Printing) {
				t.Errorf("ein Platz fehlt still: %+v", platz)
			}
		}

		if antwort.Fehler != nil {
			return
		}

		verlangt := map[string]int{}

		for _, platz := range gelesen.Plaetze {
			if platz.KarteID == "" {
				verlangt[platz.Printing+"|"+platz.Bereich] += platz.Menge
			}
		}

		for schluessel, menge := range verlangt {
			if gelegt[schluessel] != menge {
				t.Errorf("ohne Fehler fehlt etwas: %s verlangt %d, gelegt %d", schluessel, menge, gelegt[schluessel])
			}
		}
	})
}
