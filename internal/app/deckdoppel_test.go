package app_test

// **Die Doppel der Decknähte** — Werkstück 7 von Phase 4.6
// (#1293).

import (
	"context"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// deckprobe ist ein Deckbau im Arbeitsspeicher.
type deckprobe struct {
	deck   domain.Deck
	platz  domain.Deckplatz
	fehler error

	// gewaehlt merkt sich, welches Format zuletzt gestellt wurde — für die
	// Zusage, dass eine leere Kennung die Null bedeutet.
	gewaehlt int64

	// umbenennfehler scheitert **nur** beim Umbenennen.
	//
	// **Eigens getrennt von `fehler`**: `DeckUmbenennen` liest erst und
	// schreibt dann. Scheitert beides, wird der Schreibweg nie betreten — und
	// sein Fehlerzweig wäre einer, den kein Lauf erreicht (§3).
	umbenennfehler error
}

func (d *deckprobe) Anlegen(_ context.Context, deck domain.Deck) (domain.Deck, error) {
	if d.fehler != nil {
		return domain.Deck{}, d.fehler
	}

	deck.ID, deck.Fassung = 7, 1

	return deck, nil
}

func (d *deckprobe) Umbenennen(_ context.Context, deck domain.Deck) (domain.Deck, error) {
	if d.umbenennfehler != nil {
		return domain.Deck{}, d.umbenennfehler
	}

	if d.fehler != nil {
		return domain.Deck{}, d.fehler
	}

	deck.Fassung++
	d.deck = deck

	return deck, nil
}

func (d *deckprobe) Deck(_ context.Context, kennung int64) (domain.Deck, error) {
	if d.fehler != nil {
		return domain.Deck{}, d.fehler
	}

	deck := d.deck
	deck.ID = kennung

	// **Wie der Bestand**: er zählt beim Lesen eines einzelnen Decks aus den
	// Plätzen. Ein Doppel ohne diese Zeile wäre großzügiger als der Betrieb
	// (§3) — und hätte den Befund der Sichtprüfung mitgetragen.
	deck.Zaehlen()

	return deck, nil
}

func (d *deckprobe) Decks(_ context.Context, _ string) ([]domain.Deck, error) {
	if d.fehler != nil {
		return nil, d.fehler
	}

	return []domain.Deck{d.deck}, nil
}

func (d *deckprobe) PlatzLegen(_ context.Context, _ int64, platz domain.Deckplatz) (
	domain.Deckplatz, error,
) {
	if d.fehler != nil {
		return domain.Deckplatz{}, d.fehler
	}

	platz.ID = 3
	d.platz = platz

	return platz, nil
}

func (d *deckprobe) PlatzNehmen(_ context.Context, _, _ int64) error { return d.fehler }

// **Das Doppel verschiebt wirklich** — es gibt den Platz mit neuem Bereich
// zurück, statt den Aufruf nur zu zählen (§3).
func (d *deckprobe) PlatzVerschieben(
	_ context.Context, _, platzID int64, bereich domain.Deckbereich,
) (domain.Deckplatz, error) {
	if d.fehler != nil {
		return domain.Deckplatz{}, d.fehler
	}

	return domain.Deckplatz{ID: platzID, Bereich: bereich, Menge: 4}, nil
}

func (d *deckprobe) FormatWaehlen(_ context.Context, _, formatID int64) (
	domain.Deck, error,
) {
	if d.fehler != nil {
		return domain.Deck{}, d.fehler
	}

	d.gewaehlt = formatID
	d.deck.FormatID = formatID

	return d.deck, nil
}

func (d *deckprobe) Loeschen(_ context.Context, _ int64) error { return d.fehler }

// formatprobe ist eine Formatpflege im Arbeitsspeicher.
type formatprobe struct {
	format domain.Format
	fehler error
}

func (f *formatprobe) Bauarten() []service.Bauartvorschlag {
	vorschlaege := make([]service.Bauartvorschlag, 0, 3)

	for _, name := range domain.Bauarten() {
		format, _ := domain.VonBauart(name)
		vorschlaege = append(vorschlaege,
			service.Bauartvorschlag{Bauart: name, Format: format})
	}

	return vorschlaege
}

func (f *formatprobe) AusBauart(
	_ context.Context, bauart domain.Bauart, name, spiel, scryfallname string,
) (domain.Format, error) {
	if f.fehler != nil {
		return domain.Format{}, f.fehler
	}

	format, da := domain.VonBauart(bauart)
	if !da {
		return domain.Format{}, errors.New("die Bauart gibt es nicht")
	}

	format.ID, format.Fassung = 5, 1
	format.Name, format.Spiel, format.Scryfallname = name, spiel, scryfallname

	return format, nil
}

func (f *formatprobe) Anlegen(_ context.Context, format domain.Format) (
	domain.Format, error,
) {
	if f.fehler != nil {
		return domain.Format{}, f.fehler
	}

	format.ID, format.Fassung = 5, 1
	f.format = format

	return format, nil
}

func (f *formatprobe) Aendern(_ context.Context, format domain.Format) (
	domain.Format, error,
) {
	if f.fehler != nil {
		return domain.Format{}, f.fehler
	}

	format.Fassung++
	f.format = format

	return format, nil
}

func (f *formatprobe) Format(_ context.Context, kennung int64) (domain.Format, error) {
	if f.fehler != nil {
		return domain.Format{}, f.fehler
	}

	format := f.format
	format.ID = kennung

	return format, nil
}

func (f *formatprobe) Formate(_ context.Context, _ string) ([]domain.Format, error) {
	if f.fehler != nil {
		return nil, f.fehler
	}

	return []domain.Format{f.format}, nil
}

func (f *formatprobe) Loeschen(_ context.Context, _ int64) error { return f.fehler }

// pruefprobe gibt ein festes Urteil.
type pruefprobe struct {
	urteil  domain.Deckurteil
	befunde []domain.Befund
	fehler  error
}

func (p *pruefprobe) Pruefe(_ context.Context, _ int64) (
	domain.Deckurteil, []domain.Befund, error,
) {
	return p.urteil, p.befunde, p.fehler
}

// abgleichprobe gibt einen festen Stand.
type abgleichprobe struct {
	stand  service.Habenstand
	fehler error
}

func (a *abgleichprobe) Fuer(_ context.Context, _ int64) (service.Habenstand, error) {
	return a.stand, a.fehler
}
