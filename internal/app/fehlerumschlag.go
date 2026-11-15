package app

// **Die drei Ausgänge eines Umschlags** — Werkstück N4 von Phase 5.4
// (#1457).
//
// **Drei und nicht einer**, weil ein Mensch an drei verschiedenen Stellen
// nachsähe: bei der Fassung an der Gegenseite, bei der Unterschrift am
// Absender, beim Inhalt an der Datei.
//
// **Eine eigene Datei**, weil `fehler.go` mit ihnen über die Zeilengrenze aus
// §2 ginge.

import (
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// umschlagausgang gibt die Übersetzung — oder nil, wenn es keiner ist.
//
// **`mitSatz` gibt schon einen Zeiger**, und ein zweiter darum wäre einer zu
// viel — der Übersetzer gibt `*Fehler`, weil „kein Fehler" dort `nil` heißt.
func umschlagausgang(fehler error) *Fehler {
	switch {
	case errors.Is(fehler, domain.ErrUmschlagFassung):
		return mitSatz(
			CodeValidierung, domain.SchluesselUmschlagFassung, fehler)

	case errors.Is(fehler, domain.ErrUmschlagUnterschrift):
		return mitSatz(
			CodeValidierung, domain.SchluesselUmschlagUnterschrift, fehler)

	case errors.Is(fehler, domain.ErrUmschlagInhalt):
		return mitSatz(
			CodeValidierung, domain.SchluesselUmschlagInhalt, fehler)

	default:
		return nil
	}
}
