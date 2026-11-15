// **Die Formatpflege** — Werkstück 4 von Phase 4.6
// (#1293),
// nach Vermerk [0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md):
// „Daten, nicht Code".
//
// # Was diese Schicht dazutut
//
// Der Speicher schreibt Zeilen; hier steht, **was gilt** — und ein Weg, den es
// nur hier gibt: ein Format aus einer **Bauart** anlegen. Die drei Bauarten
// tragen Zahlen, die in den Comprehensive Rules stehen; wer eines von Hand
// anlegt, setzt sie selbst.

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Formatspeicher ist, was die Pflege zum Lesen und Schreiben braucht.
//
// **Die Schnittstelle steht hier und wird in `storage` erfüllt** (§2).
type Formatspeicher interface {
	Anlegen(ctx context.Context, format domain.Format) (domain.Format, error)
	Aendern(ctx context.Context, format domain.Format) (domain.Format, error)
	Format(ctx context.Context, kennung int64) (domain.Format, error)
	Formate(ctx context.Context, spiel string) ([]domain.Format, error)
	Loeschen(ctx context.Context, kennung int64) error
}

// Bauartvorschlag ist eine Bauart samt den Zahlen, die sie mitbringt.
//
// **Beides zusammen**, weil eine Fläche, die zur Wahl stellt, beides zeigt:
// „Konstruiert (60 Karten, 15 im Sideboard, höchstens 4 gleiche)".
type Bauartvorschlag struct {
	Bauart domain.Bauart
	Format domain.Format
}

// Formatpflege ist der Anwendungsfall „Formate führen".
type Formatpflege struct {
	speicher Formatspeicher
}

// NeueFormatpflege baut sie über einem Speicher.
func NeueFormatpflege(speicher Formatspeicher) *Formatpflege {
	return &Formatpflege{speicher: speicher}
}

// Bauarten gibt die Bauarten samt ihren Zahlen — **geordnet**.
func (f *Formatpflege) Bauarten() []Bauartvorschlag {
	namen := domain.Bauarten()
	vorschlaege := make([]Bauartvorschlag, 0, len(namen))

	for _, name := range namen {
		// **Ohne Zweig für „gibt es nicht"**: `Bauarten` nennt nur, was
		// `VonBauart` kennt, und ein Wächter in `domain` hält beide Richtungen
		// zusammen. Ein `if` hier wäre einer, den kein Lauf betreten kann (§3).
		format, _ := domain.VonBauart(name)

		vorschlaege = append(vorschlaege, Bauartvorschlag{Bauart: name, Format: format})
	}

	return vorschlaege
}

// AusBauart legt ein Format mit den Zahlen einer Bauart an.
//
// **Ein Vorschlag, keine Vorschrift**: wer die Zahlen danach ändert, ruft
// [Formatpflege.Aendern].
func (f *Formatpflege) AusBauart(
	ctx context.Context, bauart domain.Bauart, name, spiel, scryfallname string,
) (domain.Format, error) {
	format, da := domain.VonBauart(bauart)
	if !da {
		return domain.Format{}, domain.NewFieldError("bauart",
			domain.SchluesselValidierung, "Die Bauart %q gibt es nicht.", string(bauart))
	}

	format.Name = name
	format.Spiel = spiel
	format.Scryfallname = scryfallname

	return f.Anlegen(ctx, format)
}

// Anlegen legt ein Format an.
func (f *Formatpflege) Anlegen(ctx context.Context, format domain.Format) (
	domain.Format, error,
) {
	if err := formatVorbereiten(&format); err != nil {
		return domain.Format{}, err
	}

	return f.speicher.Anlegen(ctx, format)
}

// Aendern schreibt ein Format fort — **gegen seine Fassung** (§7, Regel 2).
func (f *Formatpflege) Aendern(ctx context.Context, format domain.Format) (
	domain.Format, error,
) {
	if err := formatVorbereiten(&format); err != nil {
		return domain.Format{}, err
	}

	return f.speicher.Aendern(ctx, format)
}

// Format gibt ein Format.
func (f *Formatpflege) Format(ctx context.Context, kennung int64) (
	domain.Format, error,
) {
	return f.speicher.Format(ctx, kennung)
}

// Formate gibt die Formate eines Spiels — ein leeres Spiel heißt: alle.
func (f *Formatpflege) Formate(ctx context.Context, spiel string) (
	[]domain.Format, error,
) {
	return f.speicher.Formate(ctx, spiel)
}

// Loeschen nimmt ein Format heraus — **und lässt die Decks stehen**.
//
// Ein Deck ohne Format ist **ungeprüft**, nicht verschwunden.
func (f *Formatpflege) Loeschen(ctx context.Context, kennung int64) error {
	return f.speicher.Loeschen(ctx, kennung)
}

// formatVorbereiten glättet ein Format und prüft es.
func formatVorbereiten(format *domain.Format) error {
	format.Normalisieren()

	if befunde := format.Validate(); len(befunde) > 0 {
		return befunde[0]
	}

	return nil
}
