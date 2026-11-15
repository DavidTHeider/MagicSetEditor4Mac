// **Ein Deck aus einem fremden Profil übernehmen, an der Fassade** —
// Werkstück 12 von Phase 4.12
// (#1351).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Deckuebernehmer ist die Naht zur Übernahme.
type Deckuebernehmer interface {
	DeckUebernehmen(ctx context.Context, quellProfil, quellDeck int64) (
		deck domain.Deck, befunde []string, fehler error)

	// DeckeZurUebernahme nennt die Decks eines fremden Profils — hinter
	// derselben Erlaubnis (UX-08.AK-04, #1680).
	DeckeZurUebernahme(ctx context.Context, quellProfil int64) (
		[]domain.Deck, error)
}

// MitDeckuebernahme setzt die Naht.
func (a *App) MitDeckuebernahme(uebernehmer Deckuebernehmer) {
	a.deckuebernahme = uebernehmer
}

// DeckUebernehmen holt ein Deck aus einem fremden Profil in das eigene.
//
// **Die Befunde sind Teil der Antwort und kein Fehler**: was nicht mitkonnte —
// ein Platz auf eine eigene Karte des anderen Profils, ein Format — steht
// darin, und das Deck ist trotzdem da (§8).
func (a *App) DeckUebernehmen(ctx context.Context, quellProfil, quellDeck string) (
	antwort Deckuebernahmeantwort,
) {
	ctx = a.imProfil(ctx)

	leer := []string{}

	if a.deckuebernahme == nil {
		return Deckuebernahmeantwort{
			Befunde: leer, Fehler: Uebersetze(ctx, keineDeckuebernahme()),
		}
	}

	profil, befund := benannteKennungAus("quellProfil", quellProfil)
	if befund != nil {
		return Deckuebernahmeantwort{Befunde: leer, Fehler: befund}
	}

	kennung, deckbefund := benannteKennungAus("quellDeck", quellDeck)
	if deckbefund != nil {
		return Deckuebernahmeantwort{Befunde: leer, Fehler: deckbefund}
	}

	roh, befunde, fehler := a.deckuebernahme.DeckUebernehmen(ctx, profil, kennung)
	if fehler != nil {
		return Deckuebernahmeantwort{Befunde: leer, Fehler: Uebersetze(ctx, fehler)}
	}

	if befunde == nil {
		befunde = leer
	}

	dto := alsDeck(roh)

	return Deckuebernahmeantwort{Deck: &dto, Befunde: befunde}
}

// DeckeZurUebernahme nennt die Decks eines fremden Profils — zur Wahl.
//
// **Statt zweier Kennungen** (UX-08.AK-04,
// #1680):
// das Profil steht in der Nutzerliste, und seine Decks nennt dieser Weg —
// **hinter derselben Erlaubnis wie die Übernahme**, geprüft, bevor gelesen
// wird (Vermerk 0071). Ohne Plätze: gewählt wird ein Deck, nicht angesehen.
func (a *App) DeckeZurUebernahme(
	ctx context.Context, quellProfil string,
) DecksAntwort {
	ctx = a.imProfil(ctx)

	if a.deckuebernahme == nil {
		return DecksAntwort{
			Decks: []DeckDTO{}, Fehler: Uebersetze(ctx, keineDeckuebernahme()),
		}
	}

	profil, befund := benannteKennungAus("quellProfil", quellProfil)
	if befund != nil {
		return DecksAntwort{Decks: []DeckDTO{}, Fehler: befund}
	}

	decks, fehler := a.deckuebernahme.DeckeZurUebernahme(ctx, profil)
	if fehler != nil {
		return DecksAntwort{Decks: []DeckDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	return DecksAntwort{Decks: alsDeckliste(decks)}
}

// DeckeZurUebernahme reicht die Wahl über die Fläche — für MCP und
// Kommandozeile, neben [Nutzerflaeche.DeckUebernehmen].
func (f *Nutzerflaeche) DeckeZurUebernahme(
	ctx context.Context, quellProfil string,
) DecksAntwort {
	return f.app.DeckeZurUebernahme(ctx, quellProfil)
}

// keineDeckuebernahme meldet eine Fläche ohne die Naht.
func keineDeckuebernahme() error {
	return fmt.Errorf("%w: die Deckübernahme ist nicht eingerichtet",
		domain.ErrValidation)
}
