package service

// **Der Wirt setzt einen eingeladenen Gast mit seinem Deck** — P9b
// (#2377).
//
// # Wann es geschieht
//
// **Beim ersten Rückkehrwunsch**, bevor der Gast irgendetwas sieht: der Stand,
// zu dem er zurückkommt, ist danach einer mit seinem Deck. Wartet sein Sitz
// nicht, bleibt alles Mitgebrachte liegen — wer nach dem Hinsetzen
// wiederkommt, setzt sich nicht noch einmal.
//
// # Und warum ein Gast ohne Deck vor dem Anfang geht
//
// **Weil er sonst an einem Sitz ohne Karten säße.** Ein Gast mit älterer
// Fassung kennt das Feld nicht; er bekommt den Grund, solange noch niemand
// gespielt hat, und nicht erst am Tisch.
//
// # Warum die Quelle hier gesetzt wird
//
// **Weil sie eine Behauptung des Gastes wäre.** Ein eigenes Deck zeigt in den
// Bestand dessen, der am anderen Ende der Leitung sitzt; eine Leihe dieses
// Wirts nimmt der Wirt aus seiner eigenen Ablage (`domain.Mitgebracht`).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Platzgeber setzt einen eingeladenen Gast mit seinem Deck.
type Platzgeber interface {
	Platznehmen(
		ctx context.Context, kennung, sitz, peer string, deck domain.Partiedeck,
		verkuende Verkuendung,
	) (Befehlsausgang, error)
}

// Leihgeber gibt den Abzug einer Leihe, die dieser Wirt ausgestellt hat.
type Leihgeber interface {
	VerliehenesDeck(
		ctx context.Context, kennung, entleiher, partie string,
	) (domain.Partiedeck, error)
}

// MitPlatzgeber setzt, wer Gäste setzt.
func (w *Partiewirt) MitPlatzgeber(platz Platzgeber) *Partiewirt {
	w.platz = platz

	return w
}

// MitLeihgeber setzt, wer die eigenen Leihen kennt.
func (w *Partiewirt) MitLeihgeber(leihen Leihgeber) *Partiewirt {
	w.leihen = leihen

	return w
}

// hinsetzen setzt den Gast, wenn sein Sitz wartet — oder sagt, warum nicht.
//
// **Ein Grund ist keine Störung**: er geht als Abbruch an den Gast. Ein
// Fehler dagegen ist einer dieses Rechners und geht zurück.
func (w *Partiewirt) hinsetzen(
	ctx context.Context, mit *domain.Mitgebracht,
) (string, error) {
	sicht, fehler := w.nehmer.Partie(ctx, w.partie, w.gast)
	if fehler != nil {
		return "", fehler
	}

	if !sitzWartet(sicht, w.gast) {
		return "", nil
	}

	if mit == nil {
		return "dieser Tisch wartet auf dein Deck, und du hast keines " +
			"mitgebracht — wähle beim Beitreten eines deiner Decks (ab 0.9.0)", nil
	}

	deck, grund := w.mitgebrachtesDeck(ctx, *mit, sicht.Spiel)
	if grund != "" {
		return grund, nil
	}

	// **Der Rest des Tisches erfährt es** — der Wirt an seinem Fenster
	// sähe sonst einen Sitz ohne Karten, bis er selbst etwas tut. **In der
	// Reihe** (#2481), wie jeder Zug.
	var verkuendet error

	if grund := w.setzen(ctx, deck, func(ausgang Befehlsausgang) {
		verkuendet = w.funk().Verkuende(ctx, w.gast, ausgang)
	}); grund != "" {
		return grund, nil
	}

	return "", verkuendet
}

// setzen gibt das Deck an den Platzgeber — und seine Absage als Grund.
func (w *Partiewirt) setzen(
	ctx context.Context, deck domain.Partiedeck, verkuende Verkuendung,
) (grund string) {
	if w.platz == nil {
		return "dieser Tisch nimmt keine mitgebrachten Decks an"
	}

	if _, fehler := w.platz.Platznehmen(
		ctx, w.partie, w.gast, w.gegenueber, deck, verkuende); fehler != nil {
		return fehler.Error()
	}

	return ""
}

// mitgebrachtesDeck gibt den Abzug, mit dem sich der Gast setzt.
func (w *Partiewirt) mitgebrachtesDeck(
	ctx context.Context, mit domain.Mitgebracht, spiel string,
) (deck domain.Partiedeck, grund string) {
	if fehler := mit.Pruefe(); fehler != nil {
		return domain.Partiedeck{}, fehler.Error()
	}

	if mit.Deck != nil && mit.Spiel != spiel {
		return domain.Partiedeck{}, fmt.Sprintf(
			"dein Deck gehört zu %q, dieser Tisch spielt %q", mit.Spiel, spiel)
	}

	if mit.Deck != nil {
		deck = *mit.Deck
		deck.Quelle = w.gegenueber

		return deck, ""
	}

	if w.leihen == nil {
		return domain.Partiedeck{}, "dieser Tisch kennt keine Leihen"
	}

	deck, fehler := w.leihen.VerliehenesDeck(ctx, mit.Leihe, w.gegenueber, w.partie)
	if fehler != nil {
		return domain.Partiedeck{}, fehler.Error()
	}

	deck.Quelle = ""

	return deck, ""
}

// sitzWartet sagt, ob der Sitz in dieser Sicht auf sein Deck wartet.
func sitzWartet(sicht domain.Partiesicht, sitz string) bool {
	for _, gesehen := range sicht.Sitze {
		if gesehen.Kennung == sitz {
			return gesehen.Wartet
		}
	}

	return false
}
