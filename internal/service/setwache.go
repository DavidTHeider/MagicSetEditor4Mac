package service

// **Gibt es dieses Set?** —
// #1336.
//
// # Der Befund
//
// `SetLeser.Karten(ctx, setID)` und `Setwerte(ctx, setID)` geben für ein
// **leeres** Set und für ein Set, **das es nicht gibt**, dieselbe leere Liste.
// Fünf Wege haben daraus dieselbe Antwort gemacht und Erfolg gemeldet —
// gemessen am 2026-08-30 an der gebauten `mse`:
//
//	mse verweise 999        → 0, leer
//	mse vorbehalte 999      → 0, „Kein Vorbehalt — alles gedeutet."
//	mse zuruecknehmen 999   → 0, „Nichts zu bewegen."
//	mse wiederherstellen 999→ 0, „Nichts zu bewegen."
//	mse suche x --set=999   → 0, „Kein Treffer."
//
// **Zwei Sachverhalte, eine Antwort.** Wer ein Skript schreibt, das über die
// Setkennungen läuft, merkt einen Tippfehler nie.
//
// # Warum das hier steht und nicht in jedem Dienst
//
// Der Druckbogen hat dieselbe Frage am 2026-08-30 **einzeln** gelöst
// (`Bogenleger.setDaSein`), weil er der erste war. Das Issue sagt, warum das
// nicht reicht: „**oder besser eine Stufe tiefer**, damit die zwölfte Stelle
// sie nicht wieder vergisst."
//
// Hier steht sie einmal. Wer einen sechsten Weg baut, findet sie — und der
// Wächter `TestJederSetwegFragtNachDemDasein` hält die Liste.
//
// # Was sie **nicht** kann
//
// **Sie sagt nicht, ob ein Set lesbar ist**, nur ob es im Bestand steht. Ein
// Set, dessen Karten nicht zu lesen sind, kommt hier durch und scheitert
// später — und das ist richtig: der Fehler ist dann ein anderer.

import (
	"context"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setkopfleser gibt die Sets, die es gibt.
//
// **Nur die Köpfe**: wer fragt, ob ein Set existiert, braucht seine Karten
// nicht (§2 — die Schnittstelle sagt, was gebraucht wird).
type Setkopfleser interface {
	Setkoepfe(ctx context.Context) ([]domain.SetHead, error)
}

// SetDaSein weist eine Kennung ab, die auf nichts zeigt.
//
// **`ErrNotFound` und nicht `ErrValidation`**: die Kennung ist wohlgeformt,
// sie zeigt nur auf nichts (§8).
func SetDaSein(ctx context.Context, leser Setkopfleser, setID int64) error {
	koepfe, err := leser.Setkoepfe(ctx)
	if err != nil {
		return fmt.Errorf("die Sets lesen: %w", err)
	}

	if slices.ContainsFunc(koepfe, func(kopf domain.SetHead) bool {
		return kopf.ID == setID
	}) {
		return nil
	}

	return fmt.Errorf("%w: ein Set mit der Kennung %d gibt es nicht",
		domain.ErrNotFound, setID)
}
