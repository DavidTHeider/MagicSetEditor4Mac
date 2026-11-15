package app_test

// **Die eigenen Namen eines Dienstes, an einer Stelle** — SEC-11
// (#2140).
// `mse-rest` und `mse-mcp` prüfen den `Host` gegen dieselbe Liste; bis dahin
// rechnete sie nur der Dienst, und der Werkzeugserver prüfte gar keinen.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieEigenenDienstnamen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		bindung      domain.Bindung
		horchadresse string
		erwartet     []string
	}{
		"auf der Schleife keine weiteren": {
			bindung: domain.Bindung{Adresse: "127.0.0.1"}, horchadresse: "127.0.0.1:7778",
		},
		"im Netz die gebundene Adresse und die Namen des Zeugnisses": {
			bindung: domain.Bindung{Adresse: "192.168.1.5"}, horchadresse: "192.168.1.5:7778",
			erwartet: []string{"192.168.1.5", "tisch.local"},
		},
		"eine unlesbare Horchadresse fügt nichts hinzu": {
			bindung: domain.Bindung{Adresse: "192.168.1.5"}, horchadresse: "unsinn",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			namen := app.EigeneDienstnamen(t.Context(), fall.bindung, fall.horchadresse, []string{"tisch.local"})
			if !slices.Equal(namen, fall.erwartet) {
				t.Errorf("gekommen ist %v, erwartet %v", namen, fall.erwartet)
			}
		})
	}
}
