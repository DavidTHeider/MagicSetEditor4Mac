package domain_test

// **Wann ein Turnier abgeschlossen werden kann** — DOM-04.AK-04
// (#1699),
// Vermerk 0139.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinTurnierSchliesstAbWennNichtsOffenIst(t *testing.T) {
	t.Parallel()

	sieg := &domain.Ergebnis{SpieleA: 2}
	gespielt := domain.Turnier{Runden: []domain.Runde{{Nummer: 1, Paarungen: []domain.Paarung{
		{ID: 1, A: 1, B: 2, Ergebnis: sieg},
		// **Ein Bye ist ein Ergebnis** (Vermerk 0062, Entscheidung 8).
		{ID: 2, A: 3},
	}}}}

	if err := gespielt.Abschliessbar(); err != nil {
		t.Errorf("ein gespieltes Turnier: %v", err)
	}

	faelle := map[string]domain.Turnier{
		"ohne Runde": {},
		"mit einem offenen Tisch": {Runden: []domain.Runde{{Nummer: 1, Paarungen: []domain.Paarung{
			{ID: 1, A: 1, B: 2},
		}}}},
	}

	for name, turnier := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var feldfehler *domain.FieldError

			err := turnier.Abschliessbar()
			if !errors.Is(err, domain.ErrValidation) || !errors.As(err, &feldfehler) ||
				feldfehler.Key != domain.SchluesselTurnierNichtFertig {
				t.Errorf("Fehler ist %v, erwartet turnierNichtFertig", err)
			}
		})
	}
}

// **Abgeschlossen ist, wer ein Datum trägt.**
func TestEinTurnierMitDatumIstAbgeschlossen(t *testing.T) {
	t.Parallel()

	if (domain.Turnier{}).IstAbgeschlossen() {
		t.Error("ein Turnier ohne Datum gilt als abgeschlossen")
	}

	if !(domain.Turnier{Abgeschlossen: "2026-09-27T12:00:00Z"}).IstAbgeschlossen() {
		t.Error("ein Turnier mit Datum gilt als offen")
	}

	// **Das Nein trägt seinen Schlüssel** — die Abhilfe ist das
	// Wiedereröffnen, nicht ein anderes Feld.
	var feldfehler *domain.FieldError

	err := domain.TurnierIstAbgeschlossen("Freitag")
	if !errors.Is(err, domain.ErrConflict) || !errors.As(err, &feldfehler) ||
		feldfehler.Key != domain.SchluesselTurnierAbgeschlossen {
		t.Errorf("Fehler ist %v, erwartet turnierAbgeschlossen", err)
	}
}
