package service_test

// **Die Turnierleitung** — Werkstück 4 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Turnier ohne Namen wird abgewiesen — an der Naht, nicht erst unten.**
func TestEinTurnierOhneNamenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)

	if _, err := werk.leitung.Anlegen(t.Context(),
		domain.Turnier{Spiel: "magic"}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}

	if werk.turniere.angelegt != 0 {
		t.Errorf("der Speicher wurde %dmal gerufen — erwartet keinmal",
			werk.turniere.angelegt)
	}
}

// **Ein Teilnehmer ohne Namen ebenso.**
func TestEinTeilnehmerOhneNamenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	if _, err := werk.leitung.Anmelden(t.Context(), turnier.ID,
		domain.Teilnehmer{}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **In ein Turnier, das es nicht gibt, meldet sich niemand an.**
//
// Ohne diese Frage entstünde aus einer falschen Kennung ein
// Fremdschlüsselfehler — und der sagt dem Aufrufer nichts.
func TestInEinFehlendesTurnierMeldetSichNiemandAn(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)

	if _, err := werk.leitung.Anmelden(t.Context(), 999,
		domain.Teilnehmer{Name: "Ada"}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Ein angemeldetes Deck wird abgezogen, nicht verwiesen.**
//
// Vermerk 0062, Entscheidung 3: die Liste ist ein **eingereichtes Dokument**.
// Was hier entsteht, ist Text — und überlebt jedes Löschen.
func TestEinAngemeldetesDeckWirdAbgezogen(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	werk.decks.deck = domain.Deck{
		ID: 7, Spiel: "magic", Name: "Aggro",
		Plaetze: []domain.Deckplatz{
			{KarteID: 3, Menge: 4, Bereich: domain.BereichHauptdeck},
			{Printing: "a1b2", Menge: 2, Bereich: domain.BereichSideboard},
		},
	}
	werk.karten.namen[3] = "Lightning Bolt"

	angemeldet, err := werk.leitung.Anmelden(t.Context(), turnier.ID,
		domain.Teilnehmer{Name: "Ada", DeckID: 7})
	if err != nil {
		t.Fatalf("anmelden: %v", err)
	}

	if len(angemeldet.Deckabzug) != 2 {
		t.Fatalf("%d Zeilen im Abzug, erwartet 2", len(angemeldet.Deckabzug))
	}

	// **Die eigene Karte steht mit ihrem Namen da**, nicht mit ihrer Kennung.
	if angemeldet.Deckabzug[0].Bezeichnung != "Lightning Bolt" {
		t.Errorf("die erste Zeile ist %+v", angemeldet.Deckabzug[0])
	}

	// **Der echte Druck steht mit seiner Kennung da** — wie er heißt, weiß
	// nur Scryfall, und eine erfundene Bezeichnung wäre für immer falsch.
	if angemeldet.Deckabzug[1].Printing != "a1b2" ||
		angemeldet.Deckabzug[1].Bezeichnung != "" {
		t.Errorf("die zweite Zeile ist %+v", angemeldet.Deckabzug[1])
	}
}

// **Eine namenlose eigene Karte bekommt ihre Kennung.**
//
// Der Abzug verlangt Text — „Karte 7" ist eine Auskunft, eine leere Zeile
// ist keine.
func TestEineNamenloseKarteBekommtIhreKennung(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	werk.decks.deck = domain.Deck{
		ID: 7, Spiel: "magic", Plaetze: []domain.Deckplatz{
			{KarteID: 9, Menge: 1, Bereich: domain.BereichHauptdeck},
		},
	}

	angemeldet, err := werk.leitung.Anmelden(t.Context(), turnier.ID,
		domain.Teilnehmer{Name: "Ada", DeckID: 7})
	if err != nil {
		t.Fatalf("anmelden: %v", err)
	}

	if angemeldet.Deckabzug[0].Bezeichnung != "Karte 9" {
		t.Errorf("die Zeile ist %+v", angemeldet.Deckabzug[0])
	}
}

// **Ein Deck, das es nicht gibt, ist ein Nein** — keine stille Anmeldung
// ohne Liste. Wer ein Deck nennt, meint eines.
func TestEinFehlendesDeckVerhindertDieAnmeldung(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	werk.decks.fehler = domain.ErrNotFound

	if _, err := werk.leitung.Anmelden(t.Context(), turnier.ID,
		domain.Teilnehmer{Name: "Ada", DeckID: 7}); !errors.Is(
		err, domain.ErrNotFound,
	) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}

	if werk.turniere.angemeldet != 0 {
		t.Error("angemeldet wurde trotzdem")
	}
}
