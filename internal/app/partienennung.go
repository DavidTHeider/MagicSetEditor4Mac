package app

// **An einem fremden Tisch Platz nehmen, ohne die Partie zu kennen** —
// UX-08.AK-04 (#1680).
//
// # Die zwei Hälften
//
// **Der Wirt spricht zuerst** ([App.begruessen]), und **der Gast hört zu,
// bevor er spricht** ([App.partieErfragen]). Beides zusammen ersetzt ein
// Feld, in das bis hierher eine Kennung von einem anderen Bildschirm gehörte
// — der Grund, warum es so herum sein muss, steht bei
// [service.Partieerfragen].

import (
	"context"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// partieErfragen wartet auf das erste Wort des Wirts und setzt sich dann an
// seinen Tisch.
//
// **Endet die Leitung vorher, ist der Tisch verloren** wie bei einem Abriss
// ([App.tischVerloren]) — und wer inzwischen aufgestanden ist, bekommt davon
// nichts mit.
func (a *App) partieErfragen(
	ctx context.Context, griff service.Angenommene, lauf uint64,
) {
	partie, zugang, fehler := service.Partieerfragen(ctx, griff)
	if fehler != nil {
		slog.InfoContext(ctx, "der Wirt hat keine Partie genannt",
			"grund", fehler)
		schliessePartiesitzung(ctx, griff)
		a.tischVerloren(lauf, fehler)

		return
	}

	// **Wer inzwischen aufgestanden ist, hat geschlossen** — zweimal schadet
	// nicht, und ein Weg, der es ausließe, wäre der eine, der es braucht.
	sitzung, gilt := a.partieErfahren(lauf, partie)
	if !gilt {
		schliessePartiesitzung(ctx, griff)

		return
	}

	a.amTischSitzen(ctx, zugang, partie, sitzung, lauf)
}

// begruessen schickt dem Gast sein Tischbild, bevor er etwas gesagt hat.
//
// **Zwei Schritte, wie überall**: [service.Partiewirt.Begruesse] legt in die
// Warteschlange, hinaus geht es mit `Leere` — und genau dieser zweite Schritt
// hat schon einmal gefehlt (`rueckkehrMelden`).
func (a *App) begruessen(
	ctx context.Context, wirt *service.Partiewirt, griff service.Angenommene,
) error {
	if fehler := wirt.Begruesse(ctx); fehler != nil {
		return fehler
	}

	return griff.Leere(ctx)
}
