package service

// **Der Gast setzt sich mit seinem Deck** — P9b
// (#2377).
//
// # Warum kein Kernbefehl
//
// **Weil kein Sitz ihn stellt.** Ein Befehl kommt von einem Sitz und wird
// gegen das Regelwerk ausgelegt; das Hinsetzen geschieht **vor** dem Spiel
// dieses Sitzes und gehört dem Wirt, der den Gast hereinlässt. Ein Gast, der
// es als Wunsch schicken könnte, setzte sich mit einem Deck seiner Wahl an
// einen Sitz seiner Wahl.
//
// # Und warum es trotzdem in den Strom geht
//
// **Weil die Partie danach eine andere ist** — sie hat ein Deck mehr. Ein
// Gast, der zurückkommt, und ein Protokoll, das später gelesen wird, sehen
// sonst einen Sitz, der plötzlich Karten hat.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// EreignisPlatznahme ist das Ereignis, wenn sich ein Gast mit seinem Deck setzt.
const EreignisPlatznahme = "sitz.platznehmen"

// Platznehmen setzt den eingeladenen Gast mit seinem Deck an seinen Sitz.
//
// **Einmal je Sitz**: die Befehlskennung steht fest, und ein zweites Mal
// bekommt dieselbe Antwort statt eines zweiten Zuges — eine Leitung, die nach
// dem Hinsetzen abreißt, setzt sich beim Wiederkommen nicht noch einmal.
//
// **Der Abdruck wird hier gerechnet** und nicht übernommen: er geht in jede
// Bildkennung ein, und was der Gast mitschickt, ist eine Behauptung.
//
// **`verkuende` trägt es an den Tisch** (#2481), in der Reihenfolge, in der es
// angenommen wurde — `nil`, wenn es keinen gibt.
func (l *Partieleitung) Platznehmen(
	ctx context.Context, kennung, sitz, peer string, deck domain.Partiedeck,
	verkuende Verkuendung,
) (Befehlsausgang, error) {
	return l.inReihe(func() (Befehlsausgang, error) {
		return l.platznehmen(ctx, kennung, sitz, peer, deck)
	}, verkuende)
}

// platznehmen ist [Partieleitung.Platznehmen] unter dem Schloss des Zuges.
func (l *Partieleitung) platznehmen(
	ctx context.Context, kennung, sitz, peer string, deck domain.Partiedeck,
) (Befehlsausgang, error) {
	partie, fehler := l.partien.Partie(ctx, kennung)
	if fehler != nil {
		return Befehlsausgang{}, fehler
	}

	befehlID := "platz/" + sitz

	schon, gab, gefragt := l.partien.Befehlsergebnis(ctx, kennung, befehlID)
	if gefragt != nil {
		return Befehlsausgang{}, gefragt
	}

	if gab {
		return Befehlsausgang{Fassung: partie.Fassung, Ereignisse: schon}, nil
	}

	if partie.Stand != domain.StandLaeuft {
		return Befehlsausgang{}, fmt.Errorf(
			"%w: diese Partie ist %s — an ihr setzt sich niemand",
			domain.ErrConflict, partie.Stand)
	}

	deck.Abdruck = abdruckDesAbzugs(deck)

	// **Erst prüfen, dann legen** (Zweitblick über Phase 7, Befund 2) — das
	// tut `bibliothekAus` selbst.
	gelegt, ungueltig := bibliothekAus(deck, sitz)
	if ungueltig != nil {
		return Befehlsausgang{}, ungueltig
	}

	danach, fehler := partie.Platznehmen(sitz, peer, deck, gelegt)
	if fehler != nil {
		return Befehlsausgang{}, fehler
	}

	return l.schreiben(ctx, partie, befehlID, danach, []domain.Ereignis{{
		Art: EreignisPlatznahme,
		Oeffentlich: map[string]any{
			"sitz": sitz, "karten": float64(len(gelegt)), "abdruck": deck.Abdruck,
		},
	}})
}
