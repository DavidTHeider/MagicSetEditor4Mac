package service_test

// **Die Vorrichtung für die Zuordnung** — Werkstück U9 von Phase 5.1
// (#1420):
// Decks, Sammlung, Zuordnungen und Reservierungen im Arbeitsspeicher.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zuordnungsdoppel ist die Ablage der Zuordnungen im Arbeitsspeicher.
type zuordnungsdoppel struct {
	zeilen   []domain.Deckzuordnung
	naechste int64
	fehler   error

	// jeExemplarFehler trifft **nur** den Weg über das Exemplar. Ohne ihn
	// scheiterte schon die Summe davor, und der Weg dahinter bliebe
	// unbetreten (U9, #1420).
	jeExemplarFehler error
}

func (z *zuordnungsdoppel) Setzen(_ context.Context, zuordnung domain.Deckzuordnung) (
	domain.Deckzuordnung, error,
) {
	if z.fehler != nil {
		return domain.Deckzuordnung{}, z.fehler
	}

	// **Nicht großzügiger als der Bestand** (§3): dort prüft `Setzen` auch.
	if befunde := zuordnung.Validate(); len(befunde) > 0 {
		return domain.Deckzuordnung{}, befunde[0]
	}

	for stelle, zeile := range z.zeilen {
		if zeile.PlatzID == zuordnung.PlatzID && zeile.ExemplarID == zuordnung.ExemplarID {
			zuordnung.ID = zeile.ID
			z.zeilen[stelle] = zuordnung

			return zuordnung, nil
		}
	}

	z.naechste++
	zuordnung.ID = z.naechste
	z.zeilen = append(z.zeilen, zuordnung)

	return zuordnung, nil
}

func (z *zuordnungsdoppel) Loesen(_ context.Context, deckID, id int64) error {
	if z.fehler != nil {
		return z.fehler
	}

	for stelle, zeile := range z.zeilen {
		if zeile.ID == id && zeile.DeckID == deckID {
			z.zeilen = append(z.zeilen[:stelle], z.zeilen[stelle+1:]...)

			return nil
		}
	}

	return domain.ErrNotFound
}

func (z *zuordnungsdoppel) JeDeck(_ context.Context, deckID int64) ([]domain.Deckzuordnung, error) {
	return z.gefiltert(func(zeile domain.Deckzuordnung) bool { return zeile.DeckID == deckID })
}

func (z *zuordnungsdoppel) JeExemplar(_ context.Context, exemplarID int64) ([]domain.Deckzuordnung, error) {
	if z.jeExemplarFehler != nil {
		return nil, z.jeExemplarFehler
	}

	return z.gefiltert(func(zeile domain.Deckzuordnung) bool { return zeile.ExemplarID == exemplarID })
}

func (z *zuordnungsdoppel) gefiltert(passt func(domain.Deckzuordnung) bool) (
	[]domain.Deckzuordnung, error,
) {
	if z.fehler != nil {
		return nil, z.fehler
	}

	gefunden := []domain.Deckzuordnung{}

	for _, zeile := range z.zeilen {
		if passt(zeile) {
			gefunden = append(gefunden, zeile)
		}
	}

	return gefunden, nil
}

func (z *zuordnungsdoppel) ZugeordnetJeExemplar(context.Context) (map[int64]int, error) {
	if z.fehler != nil {
		return nil, z.fehler
	}

	summen := map[int64]int{}
	for _, zeile := range z.zeilen {
		summen[zeile.ExemplarID] += zeile.Menge
	}

	return summen, nil
}

// reservierungsdoppel steht für das, was T2 (#1467) einträgt.
type reservierungsdoppel struct {
	je     map[int64]int
	fehler error
}

func (r *reservierungsdoppel) ReserviertJeExemplar(context.Context) (map[int64]int, error) {
	return r.je, r.fehler
}

// zuordnungswerk hält Deck, Sammlung und die beiden Ablagen zusammen.
type zuordnungswerk struct {
	*pruefwerk

	ablage     *zuordnungsdoppel
	reserviert map[int64]int
	zuordnung  *service.Deckzuordnungspflege

	exemplarID int64
	naechste   int64
}

func neuesZuordnungswerk(t *testing.T) *zuordnungswerk {
	t.Helper()

	basis := neuesPruefwerk(t)
	werk := &zuordnungswerk{
		pruefwerk:  basis,
		ablage:     &zuordnungsdoppel{},
		reserviert: map[int64]int{},
	}

	werk.zuordnung = service.NeueDeckzuordnungspflege(
		basis.decks, werk.ablage, basis.sammlung, basis.abgleich,
	).MitReservierungen(&reservierungsdoppel{je: werk.reserviert})

	return werk
}

// stoereReservierungen lässt die Naht zum Handeln scheitern.
func (w *zuordnungswerk) stoereReservierungen() {
	w.zuordnung.MitReservierungen(&reservierungsdoppel{fehler: errKaputt})
}

// ohneReservierungen nimmt die Naht heraus — der Stand vor T2.
func (w *zuordnungswerk) ohneReservierungen() {
	w.zuordnung.MitReservierungen(nil)
}

// besitze legt ein Exemplar an und merkt sich seine Kennung.
func (w *zuordnungswerk) besitze(printing string, menge int) {
	w.naechste++
	w.exemplarID = w.naechste

	w.sammlung.je[printing] = append(w.sammlung.je[printing], &domain.Exemplar{
		ID: w.exemplarID, Printing: printing, Menge: menge,
		Name: "Blitz", Setcode: "m11", Nummer: "149",
	})
}

// wuenscheZeile legt eine Wunschzeile an und gibt ihre Kennung.
func (w *zuordnungswerk) wuenscheZeile(printing string, menge int) int64 {
	w.naechste++

	w.sammlung.je[printing] = append(w.sammlung.je[printing], &domain.Exemplar{
		ID: w.naechste, Printing: printing, Menge: menge, Wunsch: true,
	})

	return w.naechste
}

// legePlatz legt einen Platz in ein Deck.
func (w *zuordnungswerk) legePlatz(t *testing.T, deckID int64, printing string, menge int) domain.Deckplatz {
	t.Helper()

	platz, err := w.bau.PlatzLegen(t.Context(), deckID, domain.Deckplatz{
		Printing: printing, Menge: menge, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	return platz
}

// legeEigeneKarte legt einen Platz auf eine eigene Karte ohne Verweis.
func (w *zuordnungswerk) legeEigeneKarte(t *testing.T, deckID int64) domain.Deckplatz {
	t.Helper()

	platz, err := w.bau.PlatzLegen(t.Context(), deckID, domain.Deckplatz{
		KarteID: eigeneKarte, Menge: 1, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	return platz
}

// zweitesDeck legt ein zweites Deck an.
func (w *zuordnungswerk) zweitesDeck(t *testing.T) domain.Deck {
	t.Helper()

	deck, err := w.bau.Anlegen(t.Context(), domain.Deck{Name: "Zweites", Spiel: "magic"})
	if err != nil {
		t.Fatalf("zweites Deck: %v", err)
	}

	return deck
}

// stand holt den Zuordnungsstand eines Decks.
func (w *zuordnungswerk) stand(t *testing.T, deckID int64) service.Zuordnungsstand {
	t.Helper()

	stand, err := w.zuordnung.Stand(t.Context(), deckID)
	if err != nil {
		t.Fatalf("Stand: %v", err)
	}

	return stand
}
