// **Die Stelle eines Syntaxfehlers kommt durch den Dienst** — Werkstück 1 von
// Phase 3.3 (#1164).
//
// # Warum dieser Lauf hier steht und nicht nur in `internal/app`
//
// Dort wird der Fehler **ausgepackt**; ob er noch verpackt ankommt,
// entscheidet sich hier. Der erste Anlauf machte aus ihm eine Zeichenkette,
// und `errors.As` fand danach nichts mehr — die Fassade hätte sich nichts
// vorwerfen lassen müssen.
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Syntaxfehler behält Zeile und Spalte.**
func TestEinSyntaxfehlerBehaeltSeineStelle(t *testing.T) {
	t.Parallel()

	konsole := service.NeueSkriptkonsole(
		konsolenkarten{karte: &domain.CardRow{ID: 1, SetID: 7}}, nil)

	_, fehler := konsole.Auswerten(t.Context(), 1, "1 +\n  * 2")
	if fehler == nil {
		t.Fatal("der Syntaxfehler kam als Erfolg durch")
	}

	// **Beides**: die Fassade braucht `ErrValidation`, um den Ausgang zu
	// kennen, und die Stelle, um sie anzustreichen.
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Fehler ist kein ErrValidation: %v", fehler)
	}

	var stelle *msescript.Fehler
	if !errors.As(fehler, &stelle) {
		t.Fatalf("der Fehler trägt keine Stelle mehr: %v", fehler)
	}

	if stelle.Zeile != 2 || stelle.Spalte != 3 {
		t.Errorf("die Stelle ist Zeile %d, Spalte %d — erwartet war 2/3",
			stelle.Zeile, stelle.Spalte)
	}
}
