package service_test

// **Ein Platz nennt seine Karte beim Namen** —
// #1376.
//
// # Der Befund
//
// Die Deckliste zeigte, was ein Platz trägt: eine Kennung. Bei einer eigenen
// Karte eine Zahl, bei einem Druck eine UUID —
// `c58096e5-3cd4-4edd-b71a-3bdb47ab2536`. **Direkt darunter** stand nach einem
// Klick auf „Gegen das Format prüfen" dieselbe Karte mit Namen; nur die untere
// Liste war lesbar.
//
// # Warum es jetzt anders geht
//
// Seit #1373
// fragt die Prüfung über `POST /cards/collection` in Schüben zu 75: hundert
// Drucke sind zwei Anfragen statt hundert.
//
// # Was hier geprüft wird
//
// Der Dienst, nicht das Fenster: `Deckbau.Deck` gibt die Namen mit. **An einer
// Stelle**, damit jede Fläche sie hat (§16) — MCP und die Kommandozeile zeigen
// sonst weiter UUIDs.
//
// # Und was er ohne Netz tut
//
// **Die Kennung stehen lassen** (§16, Regel 5): eine Anwendung, die ohne Netz
// keine Deckliste mehr zeigt, ist falsch gebaut.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// namensspeicher ist ein Deckspeicher mit einem Deck darin.
type namensspeicher struct {
	deckdoppel

	deck domain.Deck
}

func (s *namensspeicher) Deck(context.Context, int64) (domain.Deck, error) {
	return s.deck, nil
}

// druckdoppel antwortet auf die Sammelabfrage.
type deckdruckdoppel struct {
	namen  map[string]string
	fehler error
	rufe   int
}

func (d *deckdruckdoppel) Sammlung(
	_ context.Context, kennungen []scryfall.Kennung,
) (scryfall.Sammelantwort, time.Duration, []domain.Befund, error) {
	d.rufe++

	if d.fehler != nil {
		return scryfall.Sammelantwort{}, 0, nil, d.fehler
	}

	gefunden := make([]*scryfall.Karte, 0, len(kennungen))

	for _, kennung := range kennungen {
		name, da := d.namen[kennung.ID]
		if !da {
			continue
		}

		gefunden = append(gefunden, &scryfall.Karte{Kennung: kennung.ID, Name: name})
	}

	return scryfall.Sammelantwort{Karten: gefunden}, 0, nil, nil
}

// probedeck gibt ein Deck mit einer eigenen Karte und einem Druck.
func probedeck() domain.Deck {
	return domain.Deck{
		ID: 1, Name: "Probe", Spiel: "magic",
		Plaetze: []domain.Deckplatz{
			{ID: 1, KarteID: 7, Menge: 4, Bereich: domain.BereichHauptdeck},
			{ID: 2, Printing: "c58096e5", Menge: 2, Bereich: domain.BereichHauptdeck},
		},
	}
}

func TestEinPlatzNenntSeineKarteBeimNamen(t *testing.T) {
	t.Parallel()

	speicher := &namensspeicher{deck: probedeck()}
	drucke := &deckdruckdoppel{namen: map[string]string{"c58096e5": "Fires of Victory"}}

	bau := service.NeuerDeckbau(speicher, nil, nil).
		MitKartennamen(&kartennamendoppel{
			namen:  map[int64]string{7: "Lightning Bolt"},
			felder: map[string]string{"casting cost": "R"},
		}).
		MitDrucknamen(drucke)

	deck, err := bau.Deck(t.Context(), 1)
	if err != nil {
		t.Fatalf("das Deck lesen: %v", err)
	}

	if deck.Plaetze[0].Bezeichnung != "Lightning Bolt" {
		t.Errorf("die eigene Karte heißt %q", deck.Plaetze[0].Bezeichnung)
	}

	if deck.Plaetze[1].Bezeichnung != "Fires of Victory" {
		t.Errorf("der Druck heißt %q", deck.Plaetze[1].Bezeichnung)
	}
}

// **Ein Ruf für alle Drucke** — nicht einer je Platz.
func TestDieDruckeKommenInEinemRuf(t *testing.T) {
	t.Parallel()

	deck := probedeck()
	deck.Plaetze = append(deck.Plaetze,
		domain.Deckplatz{ID: 3, Printing: "b57c3674", Menge: 1, Bereich: domain.BereichHauptdeck})

	speicher := &namensspeicher{deck: deck}
	drucke := &deckdruckdoppel{namen: map[string]string{
		"c58096e5": "Fires of Victory",
		"b57c3674": "Flowstone Infusion",
	}}

	bau := service.NeuerDeckbau(speicher, nil, nil).
		MitKartennamen(&kartennamendoppel{}).
		MitDrucknamen(drucke)

	if _, err := bau.Deck(t.Context(), 1); err != nil {
		t.Fatalf("das Deck lesen: %v", err)
	}

	if drucke.rufe != 1 {
		t.Errorf("%d Rufe an Scryfall, erwartet war einer", drucke.rufe)
	}
}

// **Ohne Netz bleibt die Kennung stehen** (§16, Regel 5).
//
// Eine Anwendung, die ohne Netz keine Deckliste mehr zeigt, ist falsch gebaut —
// und ein Deck, das dann leer aussieht, ist schlimmer als eine UUID.
func TestOhneNetzBleibtDieKennungStehen(t *testing.T) {
	t.Parallel()

	speicher := &namensspeicher{deck: probedeck()}
	drucke := &deckdruckdoppel{fehler: errors.New("kein Netz")}

	bau := service.NeuerDeckbau(speicher, nil, nil).
		MitKartennamen(&kartennamendoppel{
			namen:  map[int64]string{7: "Lightning Bolt"},
			felder: map[string]string{"casting cost": "R"},
		}).
		MitDrucknamen(drucke)

	deck, err := bau.Deck(t.Context(), 1)
	if err != nil {
		t.Fatalf("das Deck lesen: %v", err)
	}

	// **Die eigene Karte steht trotzdem da** — der Bestand braucht kein Netz.
	if deck.Plaetze[0].Bezeichnung != "Lightning Bolt" {
		t.Errorf("die eigene Karte heißt %q", deck.Plaetze[0].Bezeichnung)
	}

	if deck.Plaetze[1].Bezeichnung != "" {
		t.Errorf("der Druck heißt %q, erwartet war leer", deck.Plaetze[1].Bezeichnung)
	}
}

// **Ohne die beiden Nähte tut der Deckbau, was er immer tat.**
//
// Die Kommandozeile und der MCP-Server verdrahten sie nicht alle; ein
// `nil`-Zugriff dort wäre ein Absturz für eine Bequemlichkeit.
func TestOhneNamensnaehteBleibtAllesLeer(t *testing.T) {
	t.Parallel()

	speicher := &namensspeicher{deck: probedeck()}

	deck, err := service.NeuerDeckbau(speicher, nil, nil).Deck(t.Context(), 1)
	if err != nil {
		t.Fatalf("das Deck lesen: %v", err)
	}

	for _, platz := range deck.Plaetze {
		if platz.Bezeichnung != "" {
			t.Errorf("Platz %d heißt %q, erwartet war leer", platz.ID, platz.Bezeichnung)
		}
	}
}

// **Ein Bestandsfehler lässt den Platz stehen, nicht das Deck fallen.**
//
// Beide Wege eigens: die Kennung ist der Rückfall, wenn der Name nicht kommt,
// und der Name bleibt, wenn nur die Manakosten nicht kommen. Hingen beide an
// einem Fehler, prüfte ein Test sie nur zusammen — und der zweite fiele still
// aus (§3).
func TestEinBestandsfehlerLaesstDenPlatzStehen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name        string
		doppel      *kartennamendoppel
		bezeichnung string
		manakosten  string
	}{
		{
			name: "der Name kommt nicht",
			doppel: &kartennamendoppel{
				fehler: errors.New("die Karte führt der Bestand nicht"),
				felder: map[string]string{"casting cost": "R"},
			},
		},
		{
			name: "nur die Manakosten kommen nicht",
			doppel: &kartennamendoppel{
				namen:      map[int64]string{7: "Lightning Bolt"},
				feldfehler: errors.New("das Feld gibt es nicht"),
			},
			bezeichnung: "Lightning Bolt",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			speicher := &namensspeicher{deck: probedeck()}

			deck, err := service.NeuerDeckbau(speicher, nil, nil).
				MitKartennamen(fall.doppel).
				Deck(t.Context(), 1)
			if err != nil {
				t.Fatalf("das Deck lesen: %v", err)
			}

			if deck.Plaetze[0].Bezeichnung != fall.bezeichnung {
				t.Errorf("die Karte heißt %q, erwartet war %q",
					deck.Plaetze[0].Bezeichnung, fall.bezeichnung)
			}

			if deck.Plaetze[0].Manakosten != fall.manakosten {
				t.Errorf("die Kosten sind %q, erwartet war %q",
					deck.Plaetze[0].Manakosten, fall.manakosten)
			}
		})
	}
}

// **Ein Deck ohne Drucke fragt Scryfall nicht.**
//
// Nicht Sparsamkeit, sondern §16, Regel 4: eine Anfrage mit leerer Liste zählt
// gegen die Drossel und beantwortet nichts. Der Dienst hat sie beantwortet, als
// ein Deck nur eigene Karten führte — das ist der Normalfall beim Bauen.
func TestEinDeckOhneDruckeFragtScryfallNicht(t *testing.T) {
	t.Parallel()

	deck := probedeck()
	deck.Plaetze = deck.Plaetze[:1] // nur die eigene Karte

	drucke := &deckdruckdoppel{}

	_, err := service.NeuerDeckbau(&namensspeicher{deck: deck}, nil, nil).
		MitKartennamen(&kartennamendoppel{namen: map[int64]string{7: "Lightning Bolt"}}).
		MitDrucknamen(drucke).
		Deck(t.Context(), 1)
	if err != nil {
		t.Fatalf("das Deck lesen: %v", err)
	}

	if drucke.rufe != 0 {
		t.Errorf("%d Rufe an Scryfall, erwartet war keiner", drucke.rufe)
	}
}
