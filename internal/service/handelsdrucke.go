package service

// **Welche Karte eine Handelszeile meint** — P10
// (#2378),
// Befund C-13: am Fenster stand „1 × counterspell-7ed", und wer tauschte, sah
// weder Name noch Set noch Bild.
//
// # Nachgeschlagen beim Lesen, ohne zu fragen
//
// **Wie das Journal** (`handelsjournalsicht.go`): ein eingefrorener Name wäre
// falsch, sobald jemand den Druck berichtigt, und der Vertrag führt nur die
// Kennung. **Ins Netz geht dafür nichts** (§16, Regel 5): Name, Set und
// Nummer kommen aus dem eigenen Bestand, das Bild aus dem, was Scryfall schon
// einmal beantwortet hat — beim Preisholen liegt es unter demselben Schlüssel.
//
// # Und wo es keine Antwort gibt
//
// **Ein Druck ohne Scryfall-Kennung ist ein eigener** und hat nie ein Bild —
// keines zu zeigen ist ehrlicher als das einer anderen Karte. **Die
// Druckzeile des Gegenübers wird nicht nachgeschlagen**: sie ist eine Zeile
// seines Bestands, und unter derselben Zahl steht hier etwas anderes.

import (
	"context"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Handelsdruckleser liest Drucke des eigenen Bestands.
type Handelsdruckleser interface {
	NachScryfall(ctx context.Context, printing string) (*domain.Druck, error)
	Fuer(ctx context.Context, id int64) (*domain.Druck, error)
}

// Druckvorrat ist, was dieser Rechner über einen Scryfall-Druck schon weiß.
type Druckvorrat interface {
	KarteImVorrat(printing string) (*scryfall.Karte, bool)
}

// Handelsdruck ist, was eine Handelszeile über ihre Karte sagt.
//
// **Leer heißt unbekannt** — nicht erfunden.
type Handelsdruck struct {
	Name    string
	Setcode string
	Nummer  string

	// Bild ist die Adresse eines kleinen Kartenbilds — leer ohne.
	Bild string

	// Eigen heißt: der Druck hat keine Scryfall-Kennung.
	Eigen bool
}

// Handelsdrucke gibt je Position ihren Druck, in der Reihenfolge der
// Positionen; `ich` ist die eigene Partei.
func Handelsdrucke(
	ctx context.Context, leser Handelsdruckleser, vorrat Druckvorrat,
	positionen []domain.Handelsposition, ich string,
) ([]Handelsdruck, error) {
	// **Je Druck einmal nachgeschlagen** — ein Tausch nennt denselben Druck
	// oft in mehreren Zeilen.
	jeKennung := map[string]Handelsdruck{}
	jeZeile := map[int64]Handelsdruck{}
	drucke := make([]Handelsdruck, 0, len(positionen))

	for _, position := range positionen {
		var (
			druck  Handelsdruck
			fehler error
		)

		switch {
		case position.Printing != "":
			druck, fehler = gemerkt(jeKennung, position.Printing, func() (Handelsdruck, error) {
				return scryfalldruck(ctx, leser, vorrat, position.Printing)
			})

		case position.Von == ich:
			druck, fehler = gemerkt(jeZeile, position.DruckID, func() (Handelsdruck, error) {
				return eigenerDruck(ctx, leser, position.DruckID)
			})

		default:
			druck = Handelsdruck{Eigen: true}
		}

		if fehler != nil {
			return nil, fehler
		}

		drucke = append(drucke, druck)
	}

	return drucke, nil
}

// gemerkt schlägt einmal je Schlüssel nach.
func gemerkt[K comparable](
	gesehen map[K]Handelsdruck, schluessel K, nachschlagen func() (Handelsdruck, error),
) (Handelsdruck, error) {
	if druck, da := gesehen[schluessel]; da {
		return druck, nil
	}

	druck, fehler := nachschlagen()
	if fehler != nil {
		return Handelsdruck{}, fehler
	}

	gesehen[schluessel] = druck

	return druck, nil
}

// scryfalldruck nimmt Name, Set und Nummer aus dem eigenen Bestand und füllt,
// was fehlt, aus dem Vorrat — das Bild nur von dort.
func scryfalldruck(
	ctx context.Context, leser Handelsdruckleser, vorrat Druckvorrat, printing string,
) (Handelsdruck, error) {
	var druck Handelsdruck

	if karte, da := vorrat.KarteImVorrat(printing); da {
		druck = Handelsdruck{
			Name: karte.Name, Setcode: karte.Set, Nummer: karte.Nummer, Bild: kleinesBild(karte),
		}
	}

	eigener, fehler := leser.NachScryfall(ctx, printing)
	if fehler != nil {
		return Handelsdruck{}, fehler
	}

	// **Was der Bestand nennt, gilt; was er leer lässt, nicht** — der
	// Preisabruf legt eine Druckzeile mit Kennung und ohne Namen an, und die
	// überschrieb den Namen aus dem Vorrat (gefunden von `p10handel09`).
	if eigener != nil {
		druck.Name = gefuellt(eigener.Name, druck.Name)
		druck.Setcode = gefuellt(eigener.Setcode, druck.Setcode)
		druck.Nummer = gefuellt(eigener.Nummer, druck.Nummer)
	}

	return druck, nil
}

// gefuellt nimmt den ersten Wert, der nicht leer ist.
func gefuellt(bevorzugt, sonst string) string {
	if bevorzugt != "" {
		return bevorzugt
	}

	return sonst
}

// eigenerDruck liest eine Druckzeile des eigenen Bestands.
func eigenerDruck(ctx context.Context, leser Handelsdruckleser, id int64) (Handelsdruck, error) {
	druck, fehler := leser.Fuer(ctx, id)
	if errors.Is(fehler, domain.ErrNotFound) {
		return Handelsdruck{Eigen: true}, nil
	}

	if fehler != nil {
		return Handelsdruck{}, fehler
	}

	return Handelsdruck{Name: druck.Name, Setcode: druck.Setcode, Nummer: druck.Nummer, Eigen: true}, nil
}
