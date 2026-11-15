//go:build darwin

package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Am echten System**: der Fußabdruck dieses Prozesses ist gemessen und
// größer als sein Go-Heap — ein Testlauf hält keine Webansicht, also ist
// die Liste leer und nichts ungemessen.
func TestDerFussabdruckIstGemessen(t *testing.T) {
	t.Parallel()

	prozess := app.NeueApp("").Metriken(t.Context()).Prozess
	if prozess.FussabdruckBytes <= 0 || prozess.WebUngemessen != "" || prozess.Web == nil {
		t.Errorf("Prozess %+v", prozess)
	}
}
