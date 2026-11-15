package app

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Melder einer Wurzel erreicht den Fachkern** — FMT-03 (#1692). Die
// Wurzeln ohne Fenster kennen `internal/service` nicht (§2) und hängen ihn
// über diese Naht an; gemeldet wird dort, wo ein Lauf fragt.
func TestDerMelderEinerWurzelErreichtDenFachkern(t *testing.T) {
	var gehoert []domain.Fortschritt

	ctx := MitFortschritt(context.Background(), func(stand domain.Fortschritt) {
		gehoert = append(gehoert, stand)
	})

	if fehler := service.Schritt(ctx, domain.Fortschritt{Erledigt: 2, Gesamt: 5}); fehler != nil {
		t.Fatal(fehler)
	}

	if len(gehoert) != 1 || gehoert[0].Erledigt != 2 {
		t.Errorf("gehört: %+v", gehoert)
	}
}
