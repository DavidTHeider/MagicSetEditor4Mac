package service

// **Ein abgewiesener Wunsch ist eine Antwort und kein Ende** —
// #2478.
//
// # Wogegen
//
// **Gegen die Sitzung, die an jedem Nein des Kerns endete.** `befehlen` gab
// jeden Fehler aus `Befehl` zurück, und `Partienetz.Bediene` legte daraufhin
// die Leitung nieder — ein Zug, den die Regeln nicht erlauben, oder eine
// Fassung, die sich inzwischen geändert hat, und der Gast saß an keinem
// Tisch mehr.
//
// # Wie die Absage reist
//
// **Am Tischbild, das der Gast ohnehin liest** ([domain.Wunschabsage]). Ein
// Gast vor 0.9.0 liest das Feld nicht und bekommt ein frisches Bild; eine
// eigene Kontrollart beendete gerade seine Sitzung (§16, derselbe Gedanke
// wie im Rückweg von Vermerk 0198).
//
// # Was weiter endet
//
// **Was kein Urteil des Kerns ist** — ein Bestand, der klemmt, eine Leitung,
// die bricht. Das ist kein Nein auf einen Wunsch, sondern ein Zustand dieses
// Rechners, und der Gast liest es als Abriss.

import (
	"context"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// abgewiesen beantwortet einen Wunsch, den der Kern abwies, mit dem Tischbild
// samt Grund — oder gibt einen Fehler zurück, der kein Urteil ist.
func (w *Partiewirt) abgewiesen(
	ctx context.Context, befehl domain.Spielbefehl, fehler error,
) error {
	if !einUrteil(fehler) {
		return fehler
	}

	// **In der Reihe** (#2481): das Bild der Absage darf kein neueres
	// überholen, das ein anderer Zug gerade verkündet.
	return w.nehmer.InReihe(func() error {
		bild, lesefehler := w.tischbild(ctx)
		if lesefehler != nil {
			return lesefehler
		}

		bild.Absage = &domain.Wunschabsage{
			Befehl: befehl.Kennung, Art: befehl.Art, Grund: fehler.Error(),
		}

		// **Kein Fehlerzweig** — derselbe Grund wie bei `tischbildSchicken`.
		inhalt, _ := domain.VerpackeKontrolle( //nolint:errcheck // siehe tischbildSchicken
			domain.KontrolleTischbild, bild)

		return w.sitzung.Sende(ctx, domain.ArtKontrolle, inhalt)
	})
}

// einUrteil sagt, ob ein Fehler ein Nein des Kerns auf einen Wunsch ist.
func einUrteil(fehler error) bool {
	return errors.Is(fehler, domain.ErrValidation) ||
		errors.Is(fehler, domain.ErrConflict) ||
		errors.Is(fehler, domain.ErrNotFound)
}
