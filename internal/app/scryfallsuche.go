// **Die Kartensuche bei Scryfall am Fenster** — Werkstück 1 von Phase 3.6
// (#1167).
//
// # Der Befund
//
// `scryfall.Klient.Suche` gab es seit Phase V, mit Tests und voller Abdeckung
// — und **keinem Aufrufer**. Zehn Bindungen sprachen mit Scryfall, und keine
// davon suchte: man tippte einen Kartennamen und bekam **einen** Vorschlag,
// keine Trefferliste. Wer nicht weiß, wie die Karte genau heißt, kam nicht
// weiter.

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Suchquelle ist, was das Fenster von der Kartensuche braucht.
type Suchquelle interface {
	Suche(ctx context.Context, frage string) (*service.Suchauskunft, error)
}

// MitScryfallsuche setzt die Naht — für Tests und für den Composition Root.
func (a *App) MitScryfallsuche(quelle Suchquelle) { a.scryfallsuche = quelle }

// SuchtrefferDTO ist eine gefundene Karte, wie das Fenster sie sieht.
type SuchtrefferDTO struct {
	// Kennung ist die Scryfall-Kennung des **Drucks** — der Bezug, an dem
	// Preise, Sammlung und Übernahme hängen.
	Kennung    string `json:"kennung"`
	Name       string `json:"name"`
	Setcode    string `json:"setcode"`
	Setname    string `json:"setname"`
	Nummer     string `json:"nummer"`
	Seltenheit string `json:"seltenheit"`
	Manakosten string `json:"manakosten"`
	Typzeile   string `json:"typzeile"`

	// Bild ist die Adresse eines kleinen Kartenbilds — leer, wenn es keines
	// gibt.
	Bild string `json:"bild"`
}

// ScryfallsuchAntwort trägt die Treffer samt ihrer Herkunft.
type ScryfallsuchAntwort struct {
	Treffer []SuchtrefferDTO `json:"treffer"`

	// Weitere sagt, ob Scryfall mehr hat, als hier steht.
	Weitere bool `json:"weitere"`

	// AlterStunden sagt, wie alt die Antwort ist — 0 heißt frisch geholt.
	//
	// **Eine Auskunft ohne ihr Alter ist eine Behauptung über den heutigen
	// Stand** (§16, Regel 5).
	AlterStunden int `json:"alterStunden"`

	Befunde []string `json:"befunde"`
	Fehler  *Fehler  `json:"fehler"`
}

// ScryfallSuche fragt Scryfall nach Karten.
//
// **Sie schreibt nichts.** Wer eine gefundene Karte übernehmen will, geht über
// den Abgleich — und dort steht vorher, was sich ändern würde (§16, Regel 6).
func (a *App) ScryfallSuche(ctx context.Context, frage string) ScryfallsuchAntwort {
	ctx = a.imProfil(ctx)

	leer := ScryfallsuchAntwort{Treffer: []SuchtrefferDTO{}, Befunde: []string{}}

	if a.scryfallsuche == nil {
		leer.Fehler = Uebersetze(ctx, fmt.Errorf(
			"%w: der Zugang zu Scryfall ist nicht eingerichtet", domain.ErrValidation))

		return leer
	}

	auskunft, err := a.scryfallsuche.Suche(ctx, frage)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	treffer := make([]SuchtrefferDTO, 0, len(auskunft.Treffer))
	for _, eines := range auskunft.Treffer {
		treffer = append(treffer, SuchtrefferDTO{
			Kennung: eines.Kennung, Name: eines.Name,
			Setcode: eines.Setcode, Setname: eines.Setname,
			Nummer: eines.Nummer, Seltenheit: eines.Seltenheit,
			Manakosten: eines.Manakosten, Typzeile: eines.Typzeile,
			Bild: eines.Bild,
		})
	}

	return ScryfallsuchAntwort{
		Treffer:      treffer,
		Weitere:      auskunft.Weitere,
		AlterStunden: int(auskunft.Alter.Hours()),
		Befunde:      befundtexte(auskunft.Befunde),
	}
}
