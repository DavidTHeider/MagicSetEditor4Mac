package service_test

// **Jeder Fehlerweg der Zuordnung ist betreten** — Werkstück U9 von Phase 5.1
// (#1420).
//
// Die Pflege liest aus vier Quellen: Abgleich, Decks, Zuordnungen und
// Reservierungen. Schweigt eine, ist die Antwort nicht „nichts", sondern ein
// Fehler — dieselbe Zusage wie beim Abgleich selbst.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gestoert baut ein Werk mit einem Deck, einem Platz und einem Exemplar.
func gestoert(t *testing.T) (*zuordnungswerk, domain.Deckplatz) {
	t.Helper()

	werk := neuesZuordnungswerk(t)
	werk.besitze(druckA, 3)
	platz := werk.legePlatz(t, werk.deck.ID, druckA, 3)

	return werk, platz
}

func TestJedeKlemmendeQuelleHaeltDieZuordnungAn(t *testing.T) {
	faelle := map[string]func(w *zuordnungswerk){
		"die Sammlung":       func(w *zuordnungswerk) { w.sammlung.fehler = errKaputt },
		"die Decks":          func(w *zuordnungswerk) { w.decks.fehler = errKaputt },
		"die Zuordnungen":    func(w *zuordnungswerk) { w.ablage.fehler = errKaputt },
		"die Reservierungen": func(w *zuordnungswerk) { w.stoereReservierungen() },
	}

	for name, stoere := range faelle {
		t.Run(name, func(t *testing.T) {
			werk, platz := gestoert(t)
			stoere(werk)

			if _, err := werk.zuordnung.Stand(t.Context(), werk.deck.ID); err == nil {
				t.Error("Stand kam durch")
			}

			if _, err := werk.zuordnung.Fehlkarten(t.Context(), werk.deck.ID); err == nil &&
				(name == "die Sammlung" || name == "die Decks") {
				t.Error("Fehlkarten kamen durch")
			}

			if _, err := werk.zuordnung.DecksMitExemplar(t.Context(), werk.exemplarID); err == nil &&
				name == "die Zuordnungen" {
				t.Error("der Rückweg kam durch")
			}

			if _, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 1); err == nil {
				t.Error("Zuordnen kam durch")
			}
		})
	}
}

// **Ohne Naht zum Handeln ist reserviert null** — und der Stand steht trotzdem.
func TestOhneReservierungsnahtIstNichtsReserviert(t *testing.T) {
	werk, _ := gestoert(t)
	werk.ohneReservierungen()

	stand := werk.stand(t, werk.deck.ID)
	if stand.Zeilen[0].Reserviert != 0 || stand.Zeilen[0].Frei != 3 {
		t.Errorf("%+v", stand.Zeilen[0])
	}
}

// **Eine Zuordnung an einem Platz eines anderen Drucks zählt nicht mit** —
// sonst stünde sie in zwei Zeilen.
func TestEineZuordnungStehtNurInIhrerZeile(t *testing.T) {
	werk, platz := gestoert(t)
	zweiter := werk.legePlatz(t, werk.deck.ID, druckB, 1)

	if _, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 1); err != nil {
		t.Fatalf("zuordnen: %v", err)
	}

	stand := werk.stand(t, werk.deck.ID)
	if len(stand.Zeilen) != 2 {
		t.Fatalf("%d Zeilen", len(stand.Zeilen))
	}

	for _, zeile := range stand.Zeilen {
		erwartet := 0
		if zeile.Printing == druckA {
			erwartet = 1
		}

		if zeile.Zugeordnet != erwartet {
			t.Errorf("Zeile %s: zugeordnet %d, erwartet %d", zeile.Printing, zeile.Zugeordnet, erwartet)
		}
	}

	// **Und eine Zuordnung, deren Platz nicht mehr im Deck liegt, steht in
	// keiner Zeile** — gelöscht wird sie mit ihm (Kaskade im Bestand).
	if _, err := werk.ablage.Setzen(t.Context(), domain.Deckzuordnung{
		DeckID: werk.deck.ID, PlatzID: zweiter.ID + 99, ExemplarID: werk.exemplarID, Menge: 1,
	}); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if werk.stand(t, werk.deck.ID).Zeilen[0].Zugeordnet != 1 {
		t.Error("die verwaiste Zuordnung wurde mitgezählt")
	}
}

// **Ein vollständiges Deck hat keine Fehlkarten** — die Liste bleibt leer.
func TestEinVollstaendigesDeckHatKeineFehlkarten(t *testing.T) {
	werk, _ := gestoert(t)

	liste, err := werk.zuordnung.Fehlkarten(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("Fehlkarten: %v", err)
	}

	if liste != "" {
		t.Errorf("die Liste lautet %q", liste)
	}
}

// **Ein Fehler beim Nachschlagen des Exemplars hält den Stand an** — auch
// dann, wenn die Zuordnung schon dasteht.
func TestEinUnlesbaresExemplarHaeltDenStandAn(t *testing.T) {
	werk, platz := gestoert(t)

	if _, err := werk.ablage.Setzen(t.Context(), domain.Deckzuordnung{
		DeckID: werk.deck.ID, PlatzID: platz.ID, ExemplarID: 999, Menge: 1,
	}); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	_, err := werk.zuordnung.Stand(t.Context(), werk.deck.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Zwei Rufe an dieselbe Quelle, und der zweite scheitert** — der Weg, den
// ein Doppel mit einem einzigen Schalter nie erreicht.
func TestDerZweiteRufAnDieDecksHaeltDenStandAn(t *testing.T) {
	werk, platz := gestoert(t)

	if _, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 1); err != nil {
		t.Fatalf("zuordnen: %v", err)
	}

	werk.decks.deckrufe = 0
	werk.decks.deckfehlerAb = 2

	if _, err := werk.zuordnung.Stand(t.Context(), werk.deck.ID); !errors.Is(err, errKaputt) {
		t.Errorf("Stand: %v", err)
	}

	// **Und der Rückweg liest das Deck zu jeder Zuordnung.**
	werk.decks.deckrufe = 0
	werk.decks.deckfehlerAb = 1

	if _, err := werk.zuordnung.DecksMitExemplar(t.Context(), werk.exemplarID); !errors.Is(err, errKaputt) {
		t.Errorf("Rückweg: %v", err)
	}
}

// **Die eigene Menge einer Zuordnung wird über das Exemplar gelesen** — und
// wenn das klemmt, wird nichts zugeordnet.
func TestEinKlemmenderRueckwegHaeltDasZuordnenAn(t *testing.T) {
	werk, platz := gestoert(t)
	werk.ablage.jeExemplarFehler = errKaputt

	if _, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 1); !errors.Is(err, errKaputt) {
		t.Errorf("Zuordnen: %v", err)
	}
}
