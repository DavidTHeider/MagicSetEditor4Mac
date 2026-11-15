package service_test

// **Die Zuordnung als Anwendungsfall** — Werkstück U9 von Phase 5.1
// (#1420).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Zwei Decks brauchen dieselben drei Exemplare** — die Abnahme des Issues.
//
// Der Bedarf sagt beiden „vollständig"; zugeordnet werden können sie nur
// einmal. Zwei erklärbar verschiedene Werte, und keiner davon ist falsch.
func TestZweiDecksBrauchenDasselbeUndNurEinesBekommtEs(t *testing.T) {
	werk := neuesZuordnungswerk(t)
	werk.besitze(druckA, 3)
	platz := werk.legePlatz(t, werk.deck.ID, druckA, 3)

	zweites := werk.zweitesDeck(t)
	zweiterPlatz := werk.legePlatz(t, zweites.ID, druckA, 3)

	// **Bedarf: beide vollständig.**
	for _, deck := range []int64{werk.deck.ID, zweites.ID} {
		stand := werk.stand(t, deck)

		if stand.Haben.Fehlend != 0 || stand.Zeilen[0].Zugeordnet != 0 || stand.Zeilen[0].Frei != 3 {
			t.Fatalf("Deck %d vor der Zuordnung: %+v", deck, stand.Zeilen[0])
		}
	}

	// **Zuordnung: nur einmal.**
	if _, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 3); err != nil {
		t.Fatalf("zuordnen: %v", err)
	}

	erstes := werk.stand(t, werk.deck.ID)
	if erstes.Zeilen[0].Zugeordnet != 3 || erstes.Zeilen[0].Frei != 0 {
		t.Errorf("erstes Deck: %+v", erstes.Zeilen[0])
	}

	zweiterStand := werk.stand(t, zweites.ID)
	if zweiterStand.Haben.Fehlend != 0 || zweiterStand.Zeilen[0].Frei != 0 || zweiterStand.Zeilen[0].Zugeordnet != 0 {
		t.Errorf("zweites Deck: %+v", zweiterStand.Zeilen[0])
	}

	_, err := werk.zuordnung.Zuordnen(t.Context(), zweites.ID, zweiterPlatz.ID, werk.exemplarID, 1)
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("die zweite Zuordnung: %v, erwartet ErrConflict", err)
	}
}

// **Eine Reservierung nimmt dasselbe weg wie eine Zuordnung** — und beides
// zusammen wird nie negativ (Vertrag mit T2, #1467).
func TestReservierungUndZuordnungGebenNieEineNegativeVerfuegbarkeit(t *testing.T) {
	werk := neuesZuordnungswerk(t)
	werk.besitze(druckA, 3)
	platz := werk.legePlatz(t, werk.deck.ID, druckA, 4)

	werk.reserviert[werk.exemplarID] = 2

	stand := werk.stand(t, werk.deck.ID)
	if stand.Zeilen[0].Reserviert != 2 || stand.Zeilen[0].Frei != 1 {
		t.Fatalf("mit Reservierung: %+v", stand.Zeilen[0])
	}

	if _, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 2); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("zwei bei einem freien: %v", err)
	}

	if _, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 1); err != nil {
		t.Fatalf("eines: %v", err)
	}

	// **Mehr reserviert als da**: frei ist null, nicht minus eins.
	werk.reserviert[werk.exemplarID] = 5

	danach := werk.stand(t, werk.deck.ID)
	if danach.Zeilen[0].Frei != 0 {
		t.Errorf("frei %d, erwartet 0", danach.Zeilen[0].Frei)
	}
}

// **Die eigene Zuordnung zählt beim Ändern nicht gegen sich selbst.**
func TestEineZuordnungLaesstSichErhoehenUndLoesen(t *testing.T) {
	werk := neuesZuordnungswerk(t)
	werk.besitze(druckA, 3)
	platz := werk.legePlatz(t, werk.deck.ID, druckA, 3)

	if _, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 1); err != nil {
		t.Fatalf("erste: %v", err)
	}

	gesetzt, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 3)
	if err != nil {
		t.Fatalf("erhöhen: %v", err)
	}

	if werk.stand(t, werk.deck.ID).Zeilen[0].Zugeordnet != 3 {
		t.Errorf("nach dem Erhöhen: %+v", werk.stand(t, werk.deck.ID).Zeilen[0])
	}

	if err := werk.zuordnung.Loesen(t.Context(), werk.deck.ID, gesetzt.ID); err != nil {
		t.Fatalf("lösen: %v", err)
	}

	if werk.stand(t, werk.deck.ID).Zeilen[0].Zugeordnet != 0 {
		t.Error("nach dem Lösen ist noch etwas zugeordnet")
	}
}

// **Was nicht zuordenbar ist, wird abgewiesen** — jede Klasse einmal.
func TestNichtZuordenbaresWirdAbgewiesen(t *testing.T) {
	werk := neuesZuordnungswerk(t)
	werk.besitze(druckA, 1)
	platz := werk.legePlatz(t, werk.deck.ID, druckA, 1)
	eigener := werk.legeEigeneKarte(t, werk.deck.ID)
	wunschID := werk.wuenscheZeile(druckB, 4)

	faelle := []struct {
		name                        string
		deckID, platzID, exemplarID int64
		menge                       int
		art                         error
	}{
		{"ein Wunsch ist kein Besitz", werk.deck.ID, platz.ID, wunschID, 1, domain.ErrValidation},
		{"eine eigene Karte hat keinen Druck", werk.deck.ID, eigener.ID, werk.exemplarID, 1, domain.ErrValidation},
		{"einen Platz aus einem fremden Deck", werk.deck.ID, 9999, werk.exemplarID, 1, domain.ErrNotFound},
		{"ein Exemplar, das es nicht gibt", werk.deck.ID, platz.ID, 9999, 1, domain.ErrNotFound},
		{"ein Deck, das es nicht gibt", 9999, platz.ID, werk.exemplarID, 1, domain.ErrNotFound},
		{"eine Menge unter eins", werk.deck.ID, platz.ID, werk.exemplarID, 0, domain.ErrValidation},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := werk.zuordnung.Zuordnen(t.Context(), fall.deckID, fall.platzID, fall.exemplarID, fall.menge)
			if !errors.Is(err, fall.art) {
				t.Errorf("Fehler %v, erwartet %v", err, fall.art)
			}
		})
	}
}

// **Ein anderer Druck ist eine ausdrückliche Abweichung, keine Ablehnung.**
func TestEinAndererDruckWirdAlsAbweichungGefuehrt(t *testing.T) {
	werk := neuesZuordnungswerk(t)
	werk.besitze(druckB, 1)
	platz := werk.legePlatz(t, werk.deck.ID, druckA, 1)

	if _, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 1); err != nil {
		t.Fatalf("zuordnen: %v", err)
	}

	stand := werk.stand(t, werk.deck.ID)
	if len(stand.Zeilen[0].Zuordnungen) != 1 || !stand.Zeilen[0].Zuordnungen[0].Abweichend {
		t.Errorf("%+v", stand.Zeilen[0])
	}

	if stand.Zeilen[0].Zuordnungen[0].Bezeichnung == "" {
		t.Error("die Zuordnung nennt ihr Exemplar nicht")
	}
}

// **Der Rückweg**: von einem Exemplar zu den Decks, die es benutzen.
func TestDerRueckwegNenntDieDecksEinesExemplars(t *testing.T) {
	werk := neuesZuordnungswerk(t)
	werk.besitze(druckA, 2)
	platz := werk.legePlatz(t, werk.deck.ID, druckA, 2)

	if _, err := werk.zuordnung.Zuordnen(t.Context(), werk.deck.ID, platz.ID, werk.exemplarID, 2); err != nil {
		t.Fatalf("zuordnen: %v", err)
	}

	decks, err := werk.zuordnung.DecksMitExemplar(t.Context(), werk.exemplarID)
	if err != nil {
		t.Fatalf("Rückweg: %v", err)
	}

	if len(decks) != 1 || decks[0].DeckID != werk.deck.ID || decks[0].Name == "" || decks[0].Menge != 2 {
		t.Errorf("%+v", decks)
	}
}

// **Die Fehlkarten gehen als Liste hinaus** — dieselbe Gestalt, die die
// Einfuhr liest.
func TestDieFehlkartenGehenAlsListeHinaus(t *testing.T) {
	werk := neuesZuordnungswerk(t)
	werk.besitze(druckA, 1)
	werk.legePlatz(t, werk.deck.ID, druckA, 4)

	liste, err := werk.zuordnung.Fehlkarten(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("Fehlkarten: %v", err)
	}

	if liste == "" || liste[0] != '3' {
		t.Errorf("die Liste lautet %q — erwartet drei fehlende", liste)
	}
}
