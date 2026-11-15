package service

// **Aus einem Set ein Deck** — FMT-04 (R52,
// #1693).
//
// # Warum es das braucht
//
// Ein Set, ein Deck und ein Sammlungsstück sind drei Dinge
// (`docs/DECKUNDSAMMLUNG.md`): das Set **definiert** Karten, das Deck **stellt
// sie zusammen** — mit Bereich und Menge —, die Sammlung **besitzt** Drucke.
// Wer mit den Karten eines selbst entworfenen Sets spielen will, legte bis
// hierher jede einzeln ins Deck. Hier werden sie es in einem Zug: jede Karte
// ein Platz im Hauptdeck, jede mit derselben Menge.
//
// # Was nicht geschieht
//
// **Die Karten werden nicht kopiert.** Ein Platz verweist auf die Karte im
// Set ([0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md)); wer
// sie dort berichtigt, berichtigt sie im Deck. **Und nichts entsteht in der
// Sammlung**: eine eigene Karte hat keinen Druck, den jemand besitzen könnte.
//
// **new** — MSE2 kennt kein Deck (0061, dort nachgesehen); es gibt kein
// Altverhalten, das zu portieren wäre.

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setkarten liest, was ein Deck aus einem Set braucht.
type Setkarten interface {
	Setzeile(ctx context.Context, setID int64) (domain.SetHead, error)
	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)
}

// Deckaufnahme legt ein Deck samt Plätzen und Herkunft in einem Zug an —
// dieselbe Naht wie bei der Übernahme aus einem fremden Profil.
type Deckaufnahme interface {
	Uebernehmen(ctx context.Context, deck domain.Deck, plaetze []domain.Deckplatz,
		herkunft domain.Herkunft, quellDeck int64) (domain.Deck, error)
}

// Setdeck legt Decks aus Sets an.
type Setdeck struct {
	karten Setkarten
	decks  Deckaufnahme
}

// NeuesSetdeck baut den Anwendungsfall.
func NeuesSetdeck(karten Setkarten, decks Deckaufnahme) *Setdeck {
	return &Setdeck{karten: karten, decks: decks}
}

// Anlegen legt aus den Karten eines Sets ein Deck an — ohne Namen heißt es
// wie das Set.
//
// **Ganz oder gar nicht**: Deck, Plätze und Herkunft entstehen in einer
// Transaktion (`storage.Deckablage.Uebernehmen`); scheitert ein Platz, gibt
// es kein halbes Deck.
func (s *Setdeck) Anlegen(ctx context.Context, setID int64, name string, menge int) (domain.Deck, error) {
	kopf, err := s.karten.Setzeile(ctx, setID)
	if err != nil {
		return domain.Deck{}, err
	}

	karten, err := s.karten.Karten(ctx, setID)
	if err != nil {
		return domain.Deck{}, err
	}

	if len(karten) == 0 {
		return domain.Deck{}, domain.NewFieldError("setID", domain.SchluesselValidierung,
			"Das Set %q hat keine Karten — ein Deck daraus wäre leer.", kopf.Name)
	}

	plaetze := make([]domain.Deckplatz, 0, len(karten))

	for _, karte := range karten {
		platz := domain.Deckplatz{KarteID: karte.ID, Menge: menge, Bereich: domain.BereichHauptdeck}
		if befunde := platz.Validate(); len(befunde) > 0 {
			return domain.Deck{}, befunde[0]
		}

		plaetze = append(plaetze, platz)
	}

	if name = strings.TrimSpace(name); name == "" {
		name = kopf.Name
	}

	return s.decks.Uebernehmen(ctx, domain.Deck{Name: name, Spiel: kopf.Game}, plaetze,
		domain.Herkunft{Art: domain.HerkunftEinfuhr, Quelle: kopf.Name}, 0)
}
