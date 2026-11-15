package service_test

// **Die Doppel der Übernahme** — Werkstück 12 von Phase 4.12
// (#1351).

import (
	"context"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// uebernahmedoppel ist Entnahme, Speicher und Herkunft in einem.
//
// **Eines und nicht drei**: die Läufe fragen es als **einen** Zeugen —
// „was wurde gelesen, was geschrieben, was steht als Herkunft da?" —, und
// drei Doppel hießen drei Handreichungen je Lauf.
type uebernahmedoppel struct {
	// fremde sind die Decks je Profil.
	fremde map[int64][]domain.Deck

	// entnommen sagt, ob überhaupt gelesen wurde.
	entnommen bool

	// angelegt ist, was der Speicher bekommen hat.
	angelegt domain.Deck

	// deckherkunft und quellDeck sind, was die Herkunft bekommen hat.
	deckherkunft domain.Herkunft
	quellDeck    int64

	lesefehler   error
	schreibfehl  error
	herkunftfehl error

	// platzfehler klemmt **nur** beim Platz und nicht schon beim Anlegen —
	// sonst käme der Weg nie bis dorthin.
	platzfehler error

	// ablage hält die Nutzerablage, damit ein Lauf ihren Zeiger stören kann.
	ablage *nutzerdoppel
}

// zeigerfehler lässt den Profilzeiger klemmen.
func (d *uebernahmedoppel) zeigerfehler(grund error) {
	d.ablage.zeigerfehler = grund
}

func (d *uebernahmedoppel) DeckAusProfil(
	_ context.Context, profilID, deckID int64,
) (domain.Deck, error) {
	d.entnommen = true

	if d.lesefehler != nil {
		return domain.Deck{}, d.lesefehler
	}

	for _, eines := range d.fremde[profilID] {
		if eines.ID == deckID {
			return eines, nil
		}
	}

	return domain.Deck{}, fmt.Errorf(
		"%w: das Deck %d", domain.ErrNotFound, deckID)
}

// DeckeAusProfil nennt die Decks eines Profils — und merkt, dass gelesen
// wurde, wie beim einzelnen Deck.
func (d *uebernahmedoppel) DeckeAusProfil(
	_ context.Context, profilID int64,
) ([]domain.Deck, error) {
	d.entnommen = true

	if d.lesefehler != nil {
		return nil, d.lesefehler
	}

	return d.fremde[profilID], nil
}

// Uebernehmen schreibt **wie eine Transaktion**: entweder alles, oder
// nichts (#1820). Welcher Fehler greift, sagen die drei Felder — sie klemmen
// an verschiedenen Stellen desselben Zugs, und keine hinterlässt etwas.
func (d *uebernahmedoppel) Uebernehmen(
	_ context.Context, deck domain.Deck, plaetze []domain.Deckplatz,
	herkunft domain.Herkunft, quellDeck int64,
) (domain.Deck, error) {
	for _, fehler := range []error{d.schreibfehl, d.platzfehler, d.herkunftfehl} {
		if fehler != nil {
			return domain.Deck{}, fehler
		}
	}

	deck.ID = 99
	for stelle, platz := range plaetze {
		platz.ID = int64(stelle + 1)
		deck.Plaetze = append(deck.Plaetze, platz)
	}

	d.angelegt = deck
	d.deckherkunft, d.quellDeck = herkunft, quellDeck

	return deck, nil
}

// zweiProfileMitDeck baut eine Übernahme mit vier Profilen und vier Decks.
//
// **Der Zeiger steht auf Profil 1** — der arbeitende Mensch ist Nutzer 1.
// Profil 2, 3 und 4 gehören einem anderen und tragen je ein Deck, das eine
// andere Eigenschaft prüft: echte Drucke, eine eigene Karte, ein Format.
func zweiProfileMitDeck(t *testing.T, rolle domain.Rolle) (
	*service.Deckuebernahme, *uebernahmedoppel,
) {
	t.Helper()

	ablage := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}, {ID: 2, Name: "Zweiter"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 1, NutzerID: 1, Name: "Meins"}},
			2: {
				{ID: 2, NutzerID: 2, Name: "Fremd"},
				{ID: 3, NutzerID: 2, Name: "Eigenbau"},
				{ID: 4, NutzerID: 2, Name: "Turnier"},
			},
		},
		zeiger: 1,
	}

	freigaben := neuesFreigabendoppel()

	if rolle != "" {
		for _, profil := range []int64{2, 3, 4} {
			freigaben.abgelegt = append(freigaben.abgelegt,
				domain.Profilfreigabe{ProfilID: profil, NutzerID: 1, Rolle: rolle})
		}
	}

	doppel := &uebernahmedoppel{ablage: ablage, fremde: map[int64][]domain.Deck{
		2: {{
			ID: 7, Name: "Nachtlauf", Spiel: "probe", Notiz: "geliehen",
			Plaetze: []domain.Deckplatz{
				{ID: 1, Printing: uebernahmedruck, Menge: 4, Bereich: domain.BereichHauptdeck},
				{ID: 2, Printing: uebernahmedruck, Menge: 1, Bereich: domain.BereichSideboard},
			},
		}},
		3: {{
			ID: 8, Name: "Eigenbau", Spiel: "probe",
			Plaetze: []domain.Deckplatz{
				{ID: 1, Printing: uebernahmedruck, Menge: 4, Bereich: domain.BereichHauptdeck},
				{ID: 2, KarteID: 12, Menge: 1, Bereich: domain.BereichHauptdeck},
			},
		}},
		4: {{
			ID: 9, Name: "Turnier", Spiel: "probe", FormatID: 5,
			Plaetze: []domain.Deckplatz{
				{ID: 1, Printing: uebernahmedruck, Menge: 4, Bereich: domain.BereichHauptdeck},
			},
		}},
	}}

	// **Mit einem Faktor, der jeden Nachweis abweist** — dieselbe Vorrichtung
	// wie in `wechselerlaubnis_test.go`, und aus demselben Grund: erst wo
	// einer steht, entscheidet die Freigabe etwas. Ohne ihn wäre jeder Lauf
	// hier grün, auch wenn die Erlaubnis gar nicht geprüft würde (§10).
	verwaltet := service.NeueFreigabenverwaltung(freigaben, ablage)

	// **Dieselbe Freigabenverwaltung an beiden Stellen** — so wie in beiden
	// Composition Roots. Zwei wären zwei Stände, und der Zutritt fragte einen
	// anderen als die Übernahme.
	verwaltung := service.NeueNutzerverwaltung(ablage).
		MitFaktor(verschlossen()).
		MitFreigaben(verwaltet)

	uebernahme := service.NeueDeckuebernahme(
		doppel, doppel, verwaltung, verwaltet)

	return uebernahme, doppel
}

// uebernahmedruck ist eine Scryfall-Kennung in der richtigen Gestalt.
//
// **Selbst erfunden und nicht bezogen** (§11): geprüft wird der Weg, nicht der
// Inhalt einer fremden Antwort.
const uebernahmedruck = "a1b2c3d4-0000-4000-8000-00000000000a"
