package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Set ohne Ort sagt es mit einem eigenen Fehler** — FIX-F02
// (#1746).
//
// Ein Set, das im Fenster angelegt wurde, kam aus keiner Datei; ⌘S weiß nicht,
// wohin. Die Hüllenprobe sah am 2026-09-26, dass das Fenster dann „Der gesuchte
// Eintrag wurde nicht gefunden" zeigte, statt zu fragen — der Fehler war von
// „das Set gibt es nicht" nicht zu unterscheiden. **Er bleibt ein
// `ErrNotFound`**, damit jede Fläche ihn wie bisher übersetzt, und trägt
// zusätzlich [service.ErrOhneOrt], an dem das Fenster den Dialog öffnet.
func TestEinSetOhneOrtSagtEsMitEigenemFehler(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuerSetschreiber(mitQuelle(), &ortsschreiber{}).
		Schreibe(t.Context(), 7, "")

	if !errors.Is(fehler, service.ErrOhneOrt) {
		t.Errorf("kein ErrOhneOrt: %v", fehler)
	}

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("kein ErrNotFound mehr: %v", fehler)
	}
}
