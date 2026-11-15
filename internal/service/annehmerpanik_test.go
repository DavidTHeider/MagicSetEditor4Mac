package service_test

// **Ein Zusammenbruch im Annehmer erreicht den, der bedient** — SEC-04 mit
// Vermerk 0171: seit mehrere zugleich annehmen, läuft `Nimm` nicht mehr in
// der Goroutine des Aufrufers. Der Zusammenbruch wird dort weitergeworfen,
// wo ihn der Starter fängt — sonst endete der Prozess.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

type brechenderLauscher struct{}

func (brechenderLauscher) Nimm(context.Context) (service.Verbindung, error) {
	panic("planmäßiger Zusammenbruch im Lauscher")
}

func (brechenderLauscher) Adresse() string  { return "nirgends" }
func (brechenderLauscher) Schliesse() error { return nil }

func TestEinZusammenbruchImAnnehmerErreichtDenAufrufer(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	var gefangen any

	func() {
		defer func() { gefangen = recover() }()

		service.NeuerHorchposten(paar.dort, nil).Bediene(t.Context(), brechenderLauscher{})
	}()

	if gefangen != "planmäßiger Zusammenbruch im Lauscher" {
		t.Errorf("beim Aufrufer kam an: %v", gefangen)
	}
}
