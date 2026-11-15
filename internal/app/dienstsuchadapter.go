// **Wie jeder fremde Dienst auf eine Suche antwortet** — Werkstück 8 von
// Phase 4.8 (#1321).
//
// # Der Schnitt folgt der Frage
//
// Daneben in `dienstsuchnaht.go` steht, **welche Quellen es gibt** und warum
// eine fehlt. Hier steht, **wie** eine von ihnen antwortet: die Übersetzung
// zwischen der Fähigkeit `service.Suchdienst` und dem Klienten des Anbieters.
//
// Getrennt am 2026-09-02, weil die Datei mit den Korrekturen des Nachlaufs an
// die Zeilengrenze aus §2 kam.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgapi"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtggraphql"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mtgapisuche übersetzt zwischen der Fähigkeit und dem Klienten.
type mtgapisuche struct{ klient *mtgapi.Klient }

func (m mtgapisuche) Suche(ctx context.Context, frage string) (
	*service.Suchauskunft, error,
) {
	karten, _, err := m.klient.Karten(ctx, mtgapi.Abfrage("name", frage))
	if err != nil {
		return nil, err
	}

	treffer := make([]service.Suchtreffer, 0, len(karten))
	for _, karte := range karten {
		treffer = append(treffer, service.Suchtreffer{
			Kennung:    karte.Kennung,
			Name:       karte.Name,
			Setcode:    karte.Setcode,
			Setname:    karte.Setname,
			Nummer:     karte.Nummer,
			Seltenheit: karte.Seltenheit,
			Manakosten: karte.Manakosten,
			Typzeile:   karte.Typzeile,
			Bild:       karte.Bildadresse,
		})
	}

	return &service.Suchauskunft{Treffer: treffer}, nil
}

// mtgjsonsuche übersetzt zwischen der Fähigkeit und dem GraphQL-Klienten.
type mtgjsonsuche struct{ klient *mtggraphql.Klient }

func (m mtgjsonsuche) Suche(ctx context.Context, frage string) (
	*service.Suchauskunft, error,
) {
	karten, err := m.klient.Karten(ctx, frage)
	if err != nil {
		return nil, err
	}

	treffer := make([]service.Suchtreffer, 0, len(karten))
	for _, karte := range karten {
		treffer = append(treffer, service.Suchtreffer{
			Kennung:    karte.Kennung,
			Name:       karte.Name,
			Setcode:    karte.Setcode,
			Nummer:     karte.Nummer,
			Seltenheit: karte.Seltenheit,
			Manakosten: karte.Manakosten,
			Typzeile:   karte.Typzeile,
		})
	}

	return &service.Suchauskunft{Treffer: treffer}, nil
}

// scrydexsuche übersetzt zwischen der Fähigkeit und dem Scrydex-Klienten.
type scrydexsuche struct{ klient *scrydex.Klient }

func (s scrydexsuche) Suche(ctx context.Context, frage string) (
	*service.Suchauskunft, error,
) {
	karten, err := s.klient.Karten(ctx, frage)
	if err != nil {
		return nil, err
	}

	treffer := make([]service.Suchtreffer, 0, len(karten))
	for _, karte := range karten {
		treffer = append(treffer, service.Suchtreffer{
			Kennung:    karte.Kennung,
			Name:       karte.Name,
			Setcode:    karte.Setcode,
			Setname:    karte.Setname,
			Seltenheit: karte.Seltenheit,
			Manakosten: karte.Manakosten,
			Typzeile:   karte.Typzeile,
		})
	}

	return &service.Suchauskunft{Treffer: treffer}, nil
}
