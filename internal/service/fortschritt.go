package service

// **Fortschritt und Abbruch aus einer Naht** — FMT-03.AK-02
// (#1692).
//
// Ein langer Lauf ruft zwischen zwei Stücken [Schritt]. Das tut zweierlei,
// und beides gehört zusammen: es **fragt**, ob der Lauf noch gewollt ist —
// der Kontext ist abgebrochen, wenn MCP `notifications/cancelled` bekam, der
// Dienst einen Auftrag löschte oder jemand an der Kommandozeile ⌃C drückte —,
// und es **meldet**, wie weit er ist. Wer zuhört, hängt die Fläche an, wo der
// Ruf ankommt; dieselbe Stelle wie die Vorgangskennung (`vorgang.go`).
//
// **Die Stelle des Schritts ist die Stelle, an der ein Abbruch nichts
// Halbes hinterlässt** (AK-04): vor dem nächsten Set, vor der nächsten Karte
// des Bogens, dessen Datei erst am Ende entsteht.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type fortschrittsschluessel struct{}

// MitFortschritt hängt einen Melder an den Kontext — der innerste gilt.
func MitFortschritt(ctx context.Context, melder domain.Fortschrittsmelder) context.Context {
	return context.WithValue(ctx, fortschrittsschluessel{}, melder)
}

// Schritt fragt, ob der Lauf weitergehen darf, und meldet seinen Stand.
//
// **Gefragt wird vor und nach der Meldung.** Davor, weil nach einem Abbruch
// niemand mehr zuhört und MCP verlangt, dass danach keine Meldung für diesen
// Ruf folgt; danach, weil ein Abbruch, der während der Meldung ankommt, sonst
// erst beim nächsten Schritt griffe — und einen Lauf mit einem einzigen gäbe
// es dann nicht mehr anzuhalten.
func Schritt(ctx context.Context, stand domain.Fortschritt) error {
	if fehler := ctx.Err(); fehler != nil {
		return fehler
	}

	if melder, da := ctx.Value(fortschrittsschluessel{}).(domain.Fortschrittsmelder); da {
		melder(stand)
	}

	return ctx.Err()
}

// Melde meldet einen Stand, ohne zu fragen — für den letzten, nach dem nichts
// mehr abzubrechen ist. Nach einem Abbruch schweigt es wie [Schritt].
func Melde(ctx context.Context, stand domain.Fortschritt) {
	_ = Schritt(ctx, stand) //nolint:errcheck // der Abbruch ist hier kein Fehler mehr
}
