// **Der Abschluss eines Zurückspielens** — Werkstück O7 von Phase 5.2
// (#1431).
//
// **Eine eigene Datei, weil `bestandsrueckspielen.go` an der Zeilengrenze
// steht** (§2), und der Schnitt folgt der Frage: dort der Tausch, hier, was
// **danach** gilt — der Vermerk über die Rückkehr und der Ort des alten
// Bestands.

package service

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// abschliessen vermerkt die Rückkehr und nennt den Ort des alten Bestands.
//
// **Der Vermerk hält den Weg nicht auf** (§7, Regel 3): getauscht ist
// getauscht, und ein Zurückspielen, das an seiner Buchführung scheitert, wäre
// schlimmer als eines ohne sie. **Aber er wird gemeldet** — und zwar
// **zusammen** mit dem Ort des alten Bestands, denn beides gilt: der Tausch
// hat stattgefunden, und über die unklaren Mengen steht nichts geschrieben.
func (b *Bestandsweg) abschliessen(
	ctx context.Context, lage domain.Ruecksprunglage, quelle string,
) (string, error) {
	beiseite := filepath.Join(filepath.Dir(b.orte.Datenbank), Rueckspielsicherung)

	if fehler := b.epocheVermerken(ctx, lage, filepath.Base(quelle), b.uhr()); fehler != nil {
		// **Die Marke bleibt liegen** (Befund 10): der nächste Start holt den
		// Vermerk nach. Sie hier wegzuräumen hieße, die Unsicherheit für
		// geklärt zu erklären, ohne dass jemand nachgesehen hat.
		return beiseite, fmt.Errorf(
			"der Bestand ist zurückgespielt, aber die Rückkehr ließ sich nicht "+
				"vermerken — sie wird beim nächsten Start nachgetragen; sieh "+
				"deine Sammlung von Hand nach: %w", fehler)
	}

	if fehler := b.epochenmarkeEntfernen(); fehler != nil {
		slog.Default().ErrorContext(ctx, "die Rückkehrmarke blieb liegen",
			"grund", fehler.Error())
	}

	return beiseite, nil
}
