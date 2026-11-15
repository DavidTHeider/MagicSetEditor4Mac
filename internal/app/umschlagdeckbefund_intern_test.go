package app

// **Ein Platz, der sich nicht legen lässt, nimmt die übrigen nicht mit** —
// #1995,
// gefunden beim Festlegen der Zusagen von `FuzzUmschlagdeck` (QA-06c, #1976).
//
// Der Kommentar über der Schleife in `deckAnnehmen` sagt zu, ein solcher Platz
// sei ein Befund und kein Grund, das Deck zu verwerfen (§8). Der Rumpf gab
// beim ersten Fehler zurück — alles danach fehlte, ohne dass es jemand sah.

import (
	"strings"
	"testing"
)

const (
	druckEins = "11111111-1111-4111-8111-111111111111"
	druckZwei = "22222222-2222-4222-8222-222222222222"
	druckDrei = "33333333-3333-4333-8333-333333333333"
)

func TestEinGescheiterterPlatzNimmtDieUebrigenNichtMit(t *testing.T) {
	anwendung := mitIdentitaet(t)

	antwort := anwendung.deckAnnehmen(t.Context(), DeckDTO{
		Name: "Drei Plätze", Spiel: "probe",
		Plaetze: []DeckplatzDTO{
			{Printing: druckEins, Menge: 2, Bereich: "hauptdeck"},
			{Printing: druckZwei, Menge: -3, Bereich: "hauptdeck"},
			{Printing: druckDrei, Menge: 1, Bereich: "sideboard"},
		},
	})

	if antwort.Deck == nil {
		t.Fatalf("es kam kein Deck: %+v", antwort.Fehler)
	}

	gelegt := map[string]int{}
	for _, platz := range antwort.Deck.Plaetze {
		gelegt[platz.Printing] = platz.Menge
	}

	if gelegt[druckEins] != 2 || gelegt[druckDrei] != 1 {
		t.Errorf("gelegt %v — der erste und der dritte gehören ins Deck", gelegt)
	}

	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, druckZwei) {
		t.Errorf("der gescheiterte Platz ist nicht benannt: %+v", antwort.Fehler)
	}
}

// **Beide Befunde stehen in der Antwort**: der über eine fremde eigene Karte
// (FMT-04) und der über einen Platz, der sich nicht legen ließ.
func TestFremdeKarteUndGescheiterterPlatzStehenBeide(t *testing.T) {
	anwendung := mitIdentitaet(t)

	antwort := anwendung.deckAnnehmen(t.Context(), DeckDTO{
		Name: "Beides", Spiel: "probe",
		Plaetze: []DeckplatzDTO{
			{KarteID: "7", Bezeichnung: "Glutwind", Menge: 2, Bereich: "hauptdeck"},
			{Printing: druckZwei, Menge: -3, Bereich: "hauptdeck"},
		},
	})

	if antwort.Deck == nil || antwort.Fehler == nil {
		t.Fatalf("Deck %+v, Fehler %+v", antwort.Deck, antwort.Fehler)
	}

	for _, gesucht := range []string{"Glutwind", druckZwei} {
		if !strings.Contains(antwort.Fehler.Meldung, gesucht) {
			t.Errorf("die Meldung nennt %q nicht: %s", gesucht, antwort.Fehler.Meldung)
		}
	}
}
