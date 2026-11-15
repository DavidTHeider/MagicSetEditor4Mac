package scryfall

import (
	"context"
	"net/url"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Rulings holt die Regelhinweise zu einer Karte.
//
// # Warum die allgemeine Drossel
//
// `/cards/:id/rulings` fällt nicht unter die 2/s der Kartensuche. Nachgelesen
// am 2026-08-23 unter `https://scryfall.com/docs/api`: die genannten Wege sind
// `/cards/search`, `/cards/named`, `/cards/random` und `/cards/collection`;
// alles Übrige darf 10/s.
//
// # Und warum derselbe Zwischenspeicher
//
// Regelhinweise ändern sich selten und die Bedingungen verlangen mindestens 24
// Stunden. Ein eigener Speicher daneben hätte eine zweite Frist, die niemand
// nachhält.
func (k *Klient) Rulings(ctx context.Context, kartenID string) (
	hinweise []Regelhinweis, alter time.Duration, befunde []domain.Befund, fehler error,
) {
	adresse := k.grundadresse + "/cards/" + url.PathEscape(kartenID) + "/rulings"

	roh, alter, holfehler := k.speicher.Hole(ctx, "rulings:"+kartenID, func() ([]byte, error) {
		return k.holeMit(ctx, k.allgemein, adresse,
			"die Regelhinweise zu dieser Karte")
	})
	if holfehler != nil {
		return nil, 0, nil, holfehler
	}

	gelesen, befunde, lesefehler := LiesRulings(roh)
	if lesefehler != nil {
		return nil, 0, nil, lesefehler
	}

	return gelesen, alter, befunde, nil
}

// Setangaben holt, was Scryfall über einen Satz sagt.
//
// **Dieselbe Klasse wie die Rulings** — `/sets/:code` darf 10/s.
func (k *Klient) Setangaben(ctx context.Context, code string) (
	satz *Setangabe, alter time.Duration, befunde []domain.Befund, fehler error,
) {
	adresse := k.grundadresse + "/sets/" + url.PathEscape(code)

	roh, alter, holfehler := k.speicher.Hole(ctx, "sets:"+code, func() ([]byte, error) {
		return k.holeMit(ctx, k.allgemein, adresse, "diesen Satz")
	})
	if holfehler != nil {
		return nil, 0, nil, holfehler
	}

	gelesen, befunde, lesefehler := LiesSetangabe(roh)
	if lesefehler != nil {
		return nil, 0, nil, lesefehler
	}

	return gelesen, alter, befunde, nil
}

// Symbologie holt die Liste aller Kartensymbole.
//
// # Warum sie einen eigenen Speicherschlüssel hat und keinen Parameter
//
// Weil es nur **eine** gibt. `/symbology` nimmt kein Argument, und die Antwort
// gilt für alle Karten — ein Schlüssel je Aufruf wäre ein Speicher, der nie
// trifft.
func (k *Klient) Symbologie(ctx context.Context) (
	symbole []Symbol, alter time.Duration, befunde []domain.Befund, fehler error,
) {
	roh, alter, holfehler := k.speicher.Hole(ctx, "symbology", func() ([]byte, error) {
		return k.holeMit(ctx, k.allgemein, k.grundadresse+"/symbology",
			"die Symbolliste")
	})
	if holfehler != nil {
		return nil, 0, nil, holfehler
	}

	gelesen, befunde, lesefehler := LiesSymbole(roh)
	if lesefehler != nil {
		return nil, 0, nil, lesefehler
	}

	return gelesen, alter, befunde, nil
}
