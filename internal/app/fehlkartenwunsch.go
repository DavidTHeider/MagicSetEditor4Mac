package app

// **Was einem Deck fehlt, auf die Wunschliste** — DOM-03.AK-02
// (#1698).
//
// **Die Regel steht im Dienst** (§2): was neu darauf kommt und was bleibt,
// entscheidet `service.Fehlkartenwunsch`. Hier wird übersetzt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Fehlkartenwunschnaht ist, was die Flächen für den Wunschweg brauchen.
type Fehlkartenwunschnaht interface {
	Wuenschen(ctx context.Context, deckID int64) (service.Fehlkartenwuensche, error)
}

// MitFehlkartenwunsch setzt die Naht — an **beiden** Composition Roots.
func (a *App) MitFehlkartenwunsch(naht Fehlkartenwunschnaht) { a.fehlkartenwunsch = naht }

// VorhandenerWunschDTO ist ein Druck, der fehlt und schon gewünscht ist.
type VorhandenerWunschDTO struct {
	HabenzeileDTO

	// Gewuenscht ist, wie viele davon schon auf der Wunschliste stehen.
	Gewuenscht int `json:"gewuenscht"`
}

// FehlkartenwunschAntwort sagt, was neu gewünscht ist und was schon war.
type FehlkartenwunschAntwort struct {
	Neu       []ExemplarDTO          `json:"neu"`
	Vorhanden []VorhandenerWunschDTO `json:"vorhanden"`

	// Unbekannt zählt eigene Entwürfe ohne Verweis — ohne Druck kein Wunsch.
	Unbekannt int     `json:"unbekannt"`
	Fehler    *Fehler `json:"fehler"`
}

// DeckFehlkartenWuenschen setzt, was einem Deck fehlt und noch nicht
// gewünscht ist, auf die Wunschliste.
func (a *App) DeckFehlkartenWuenschen(ctx context.Context, deckID string) FehlkartenwunschAntwort {
	ctx = a.imProfil(ctx)

	// **Leere Listen statt `null`** — die Fläche zählt sie, auch im Fehlerfall.
	antwort := FehlkartenwunschAntwort{Neu: []ExemplarDTO{}, Vorhanden: []VorhandenerWunschDTO{}}

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		antwort.Fehler = befund

		return antwort
	}

	ergebnis, err := a.fehlkartenwunsch.Wuenschen(ctx, kennung)
	if err != nil {
		antwort.Fehler = Uebersetze(ctx, err)

		return antwort
	}

	for stelle := range ergebnis.Neu {
		antwort.Neu = append(antwort.Neu, zuExemplarDTO(&ergebnis.Neu[stelle]))
	}

	for _, vorhanden := range ergebnis.Vorhanden {
		antwort.Vorhanden = append(antwort.Vorhanden, VorhandenerWunschDTO{
			HabenzeileDTO: alsHabenzeile(vorhanden.Habenzeile),
			Gewuenscht:    vorhanden.Gewuenscht,
		})
	}

	antwort.Unbekannt = ergebnis.Unbekannt

	return antwort
}
