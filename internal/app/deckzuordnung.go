package app

// **Die Zuordnung an der Naht** — Werkstück U9 von Phase 5.1
// (#1420).
//
// **Eine eigene Datei neben `deckpruefung.go`.** Der Schnitt folgt der Frage:
// dort steht, was man über ein Deck erfährt; hier, welche Exemplare darin
// liegen sollen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Deckzuordnungsnaht ist, was das Fenster von der Zuordnung braucht.
type Deckzuordnungsnaht interface {
	Stand(ctx context.Context, deckID int64) (service.Zuordnungsstand, error)
	Zuordnen(ctx context.Context, deckID, platzID, exemplarID int64, menge int) (
		domain.Deckzuordnung, error)
	Loesen(ctx context.Context, deckID, id int64) error
	Fehlkarten(ctx context.Context, deckID int64) (string, error)
	DecksMitExemplar(ctx context.Context, exemplarID int64) ([]service.Deckbezug, error)
}

// ZuordnungDTO ist eine Zuordnung, wie eine Fläche sie sieht.
type ZuordnungDTO struct {
	ID         string `json:"id"`
	PlatzID    string `json:"platzID"`
	ExemplarID string `json:"exemplarID"`
	Menge      int    `json:"menge"`

	// Bezeichnung ist, wie die Sammlung das Exemplar nennt.
	Bezeichnung string `json:"bezeichnung"`

	// Abweichend sagt, dass das Exemplar einen **anderen** Druck trägt als
	// der Platz — ausdrücklich gewählt, nie über Namensgleichheit.
	Abweichend bool `json:"abweichend"`
}

// ZuordnungszeileDTO ist eine Zeile des Bedarfs samt ihren Zuordnungen.
type ZuordnungszeileDTO struct {
	HabenzeileDTO

	// Reserviert, Zugeordnet und Frei gelten für **alle** Exemplare dieses
	// Drucks, nicht nur für die dieses Decks — sonst sähe ein zweites Deck
	// Stücke als frei, die längst vergeben sind.
	Reserviert int `json:"reserviert"`
	Zugeordnet int `json:"zugeordnet"`
	Frei       int `json:"frei"`

	Zuordnungen []ZuordnungDTO `json:"zuordnungen"`
}

// ZuordnungsstandAntwort trägt den Bedarf samt den Zuordnungen.
type ZuordnungsstandAntwort struct {
	Zeilen []ZuordnungszeileDTO `json:"zeilen"`

	// Fehlend und Unbekannt stehen getrennt — wie beim Abgleich.
	Fehlend   int `json:"fehlend"`
	Unbekannt int `json:"unbekannt"`

	Fehler *Fehler `json:"fehler"`
}

// ZuordnungAntwort trägt eine gesetzte Zuordnung.
type ZuordnungAntwort struct {
	Zuordnung *ZuordnungDTO `json:"zuordnung"`
	Fehler    *Fehler       `json:"fehler"`
}

// FehlkartenAntwort trägt die Liste dessen, was zu besorgen wäre.
type FehlkartenAntwort struct {
	Liste  string  `json:"liste"`
	Fehler *Fehler `json:"fehler"`
}

// DeckbezugDTO sagt, welches Deck ein Exemplar benutzt.
type DeckbezugDTO struct {
	DeckID string `json:"deckID"`
	Name   string `json:"name"`
	Menge  int    `json:"menge"`
}

// DeckbezuegeAntwort trägt den Rückweg von der Sammlung zu den Decks.
type DeckbezuegeAntwort struct {
	Decks  []DeckbezugDTO `json:"decks"`
	Fehler *Fehler        `json:"fehler"`
}

// MitDeckzuordnung setzt die Naht.
func (a *App) MitDeckzuordnung(zuordnung Deckzuordnungsnaht) { a.deckzuordnung = zuordnung }

// DeckZuordnungen gibt den Bedarf eines Decks samt seinen Zuordnungen.
func (a *App) DeckZuordnungen(ctx context.Context, deckID string) ZuordnungsstandAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return ZuordnungsstandAntwort{Zeilen: []ZuordnungszeileDTO{}, Fehler: befund}
	}

	stand, err := a.deckzuordnung.Stand(ctx, kennung)
	if err != nil {
		return ZuordnungsstandAntwort{Zeilen: []ZuordnungszeileDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	antwort := ZuordnungsstandAntwort{
		Zeilen:    make([]ZuordnungszeileDTO, 0, len(stand.Zeilen)),
		Fehlend:   stand.Haben.Fehlend,
		Unbekannt: stand.Haben.Unbekannt,
	}

	for _, zeile := range stand.Zeilen {
		antwort.Zeilen = append(antwort.Zeilen, alsZuordnungszeile(zeile))
	}

	return antwort
}

// alsZuordnungszeile bildet eine Zeile ab.
func alsZuordnungszeile(zeile service.Zuordnungszeile) ZuordnungszeileDTO {
	dto := ZuordnungszeileDTO{
		HabenzeileDTO: alsHabenzeile(zeile.Habenzeile),
		Reserviert:    zeile.Reserviert,
		Zugeordnet:    zeile.Zugeordnet,
		Frei:          zeile.Frei,
		Zuordnungen:   make([]ZuordnungDTO, 0, len(zeile.Zuordnungen)),
	}

	for _, zuordnung := range zeile.Zuordnungen {
		dto.Zuordnungen = append(dto.Zuordnungen, ZuordnungDTO{
			ID:          zahltext(zuordnung.ID),
			PlatzID:     zahltext(zuordnung.PlatzID),
			ExemplarID:  zahltext(zuordnung.ExemplarID),
			Menge:       zuordnung.Menge,
			Bezeichnung: zuordnung.Bezeichnung,
			Abweichend:  zuordnung.Abweichend,
		})
	}

	return dto
}

// DeckZuordnen legt Exemplare in einen Platz.
func (a *App) DeckZuordnen(
	ctx context.Context, deckID, platzID, exemplarID string, menge int,
) ZuordnungAntwort {
	ctx = a.imProfil(ctx)

	kennungen, befund := dreiKennungen(deckID, platzID, exemplarID)
	if befund != nil {
		return ZuordnungAntwort{Fehler: befund}
	}

	gesetzt, err := a.deckzuordnung.Zuordnen(ctx, kennungen[0], kennungen[1], kennungen[2], menge)
	if err != nil {
		return ZuordnungAntwort{Fehler: Uebersetze(ctx, err)}
	}

	dto := ZuordnungDTO{
		ID:         zahltext(gesetzt.ID),
		PlatzID:    zahltext(gesetzt.PlatzID),
		ExemplarID: zahltext(gesetzt.ExemplarID),
		Menge:      gesetzt.Menge,
	}

	return ZuordnungAntwort{Zuordnung: &dto}
}

// dreiKennungen liest Deck, Platz und Exemplar — den ersten Befund als Fehler.
func dreiKennungen(deckID, platzID, exemplarID string) ([3]int64, *Fehler) {
	var kennungen [3]int64

	for stelle, feld := range [3]struct{ name, wert string }{
		{"deckID", deckID}, {"platzID", platzID}, {"exemplarID", exemplarID},
	} {
		kennung, befund := benannteKennungAus(feld.name, feld.wert)
		if befund != nil {
			return kennungen, befund
		}

		kennungen[stelle] = kennung
	}

	return kennungen, nil
}

// DeckZuordnungLoesen nimmt eine Zuordnung aus ihrem Deck.
func (a *App) DeckZuordnungLoesen(ctx context.Context, deckID, zuordnungID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	zuordnung, zuordnungsbefund := benannteKennungAus("zuordnungID", zuordnungID)
	if zuordnungsbefund != nil {
		return SchreibAntwort{Fehler: zuordnungsbefund}
	}

	if err := a.deckzuordnung.Loesen(ctx, kennung, zuordnung); err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return SchreibAntwort{}
}

// DeckFehlkarten gibt, was zu besorgen wäre — als Liste.
func (a *App) DeckFehlkarten(ctx context.Context, deckID string) FehlkartenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return FehlkartenAntwort{Fehler: befund}
	}

	liste, err := a.deckzuordnung.Fehlkarten(ctx, kennung)
	if err != nil {
		return FehlkartenAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return FehlkartenAntwort{Liste: liste}
}

// ExemplarDecks sagt, welche Decks ein Exemplar benutzen — der Rückweg.
func (a *App) ExemplarDecks(ctx context.Context, exemplarID string) DeckbezuegeAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("exemplarID", exemplarID)
	if befund != nil {
		return DeckbezuegeAntwort{Decks: []DeckbezugDTO{}, Fehler: befund}
	}

	bezuege, err := a.deckzuordnung.DecksMitExemplar(ctx, kennung)
	if err != nil {
		return DeckbezuegeAntwort{Decks: []DeckbezugDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	antwort := DeckbezuegeAntwort{Decks: make([]DeckbezugDTO, 0, len(bezuege))}

	for _, bezug := range bezuege {
		antwort.Decks = append(antwort.Decks, DeckbezugDTO{
			DeckID: zahltext(bezug.DeckID), Name: bezug.Name, Menge: bezug.Menge,
		})
	}

	return antwort
}
