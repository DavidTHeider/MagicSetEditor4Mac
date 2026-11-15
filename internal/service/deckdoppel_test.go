package service_test

// **Die Doppel des Deckbaus** — Werkstück 3 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `deckbau_test.go` unter der Zeilengrenze aus §2
// bleibt: dort steht, was zugesagt ist, hier, wogegen geprüft wird.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gueltigerDruck ist eine Scryfall-Kennung in der richtigen Gestalt.
const gueltigerDruck = "a1b2c3d4-0000-0000-0000-000000000000"

// deckdoppel ist ein Deckspeicher im Arbeitsspeicher.
//
// **Er hält die Fassung wie der Bestand** (§3, „eine Vorrichtung, die
// großzügiger ist als der Betrieb"): `Anlegen` setzt 1, `Aendern` prüft gegen
// die mitgebrachte und erhöht.
type deckdoppel struct {
	decks    map[int64]domain.Deck
	naechste int64

	// gelegt zählt die Rufe an PlatzSetzen — für die Zusage, dass ein
	// abgewiesener Platz den Speicher gar nicht erreicht.
	gelegt int

	// naechsterPlatz vergibt die Kennungen der Plätze — **über alle Decks**,
	// wie `AUTOINCREMENT` im Bestand. Hier stand `len(Plaetze)+1`, und damit
	// hatte der erste Platz **jedes** Decks die Eins: eine Vorrichtung, die
	// großzügiger ist als der Betrieb (§3). Aufgefallen ist es U9 (#1420) —
	// die Zuordnung sucht ihre eigene Menge über die Platzkennung und fand
	// die eines fremden Decks.
	naechsterPlatz int64

	// deckfehlerAb lässt `Deck` erst ab dem n-ten Ruf scheitern — 0 heißt
	// nie. **Für die Wege, die zweimal lesen** (U9, #1420): der Abgleich
	// holt das Deck, und danach holt es die Zuordnung noch einmal. Ein
	// Doppel, das immer scheitert, erreicht den zweiten Ruf nie.
	deckfehlerAb int
	deckrufe     int

	// kopiert sind die Namen, unter denen kopiert werden sollte (DOM-03).
	kopiert []string

	// fehler ist, was jeder Weg zurückgibt — für die Fehlerläufe.
	fehler error
}

func (d *deckdoppel) Anlegen(_ context.Context, deck domain.Deck) (domain.Deck, error) {
	if d.fehler != nil {
		return domain.Deck{}, d.fehler
	}

	if d.decks == nil {
		d.decks = map[int64]domain.Deck{}
	}

	d.naechste++
	deck.ID = d.naechste
	deck.Fassung = 1
	d.decks[deck.ID] = deck

	return deck, nil
}

func (d *deckdoppel) Aendern(_ context.Context, deck domain.Deck) (domain.Deck, error) {
	if d.fehler != nil {
		return domain.Deck{}, d.fehler
	}

	stand, da := d.decks[deck.ID]
	if !da {
		return domain.Deck{}, fmt.Errorf("%w: das Deck gibt es nicht", domain.ErrNotFound)
	}

	if stand.Fassung != deck.Fassung {
		return domain.Deck{}, fmt.Errorf(
			"%w: das Deck hat sich seit dem Lesen geändert", domain.ErrConflict)
	}

	deck.Fassung++
	d.decks[deck.ID] = deck

	return deck, nil
}

func (d *deckdoppel) Deck(_ context.Context, kennung int64) (domain.Deck, error) {
	d.deckrufe++
	if d.deckfehlerAb != 0 && d.deckrufe >= d.deckfehlerAb {
		return domain.Deck{}, errKaputt
	}

	if d.fehler != nil {
		return domain.Deck{}, d.fehler
	}

	deck, da := d.decks[kennung]
	if !da {
		return domain.Deck{}, fmt.Errorf("%w: das Deck gibt es nicht", domain.ErrNotFound)
	}

	return deck, nil
}

func (d *deckdoppel) Decks(_ context.Context, spiel string) ([]domain.Deck, error) {
	if d.fehler != nil {
		return nil, d.fehler
	}

	decks := []domain.Deck{}

	for _, deck := range d.decks {
		if spiel == "" || deck.Spiel == spiel {
			decks = append(decks, deck)
		}
	}

	return decks, nil
}

func (d *deckdoppel) PlatzSetzen(
	_ context.Context, deck int64, platz domain.Deckplatz,
) (domain.Deckplatz, error) {
	d.gelegt++

	if d.fehler != nil {
		return domain.Deckplatz{}, d.fehler
	}

	stand := d.decks[deck]
	d.naechsterPlatz++
	platz.ID = d.naechsterPlatz
	stand.Plaetze = append(stand.Plaetze, platz)
	d.decks[deck] = stand

	return platz, nil
}

func (d *deckdoppel) PlatzLoeschen(_ context.Context, _, _ int64) error {
	return d.fehler
}

// **Das Doppel verschiebt wirklich** — es schreibt den Bereich um und legt
// zusammen, wenn dort schon derselbe Bezug liegt. Ein Doppel, das nur den
// Aufruf zählte, prüfte den Weg und nicht die Wirkung (§3).
func (d *deckdoppel) PlatzVerschieben(
	_ context.Context, deck, platz int64, bereich domain.Deckbereich,
) (domain.Deckplatz, error) {
	if d.fehler != nil {
		return domain.Deckplatz{}, d.fehler
	}

	stand := d.decks[deck]

	stelle := -1

	for i, eintrag := range stand.Plaetze {
		if eintrag.ID == platz {
			stelle = i
		}
	}

	if stelle < 0 {
		return domain.Deckplatz{}, fmt.Errorf(
			"%w: diesen Deckplatz gibt es nicht", domain.ErrNotFound)
	}

	verschoben := stand.Plaetze[stelle]
	verschoben.Bereich = bereich

	uebrig := make([]domain.Deckplatz, 0, len(stand.Plaetze))

	for i, eintrag := range stand.Plaetze {
		switch {
		case i == stelle:
			continue
		case eintrag.Bereich == bereich && eintrag.Printing == verschoben.Printing &&
			eintrag.KarteID == verschoben.KarteID:
			verschoben.ID = eintrag.ID
			verschoben.Menge += eintrag.Menge
		default:
			uebrig = append(uebrig, eintrag)
		}
	}

	stand.Plaetze = append(uebrig, verschoben) //nolint:gocritic // uebrig ist die neue Liste
	d.decks[deck] = stand

	return verschoben, nil
}

func (d *deckdoppel) Loeschen(_ context.Context, kennung int64) error {
	if d.fehler != nil {
		return d.fehler
	}

	delete(d.decks, kennung)

	return nil
}

// kartenspieldoppel sagt, zu welchem Spiel eine eigene Karte gehört.
type kartenspieldoppel struct {
	spiel  string
	fehler error
}

func (k *kartenspieldoppel) SpielZurKarte(_ context.Context, _ int64) (string, error) {
	return k.spiel, k.fehler
}
