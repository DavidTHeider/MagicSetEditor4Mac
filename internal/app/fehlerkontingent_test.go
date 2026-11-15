package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Das eigene Kontingent ist eine eigene Klasse** — Befund 5 des Zweitblicks
// über Phase 1 von 0.9.0 (F03, #2342): derselbe Code `drossel`, den der
// Dienst seit #1429 nennt, und die Wartezeit als Wert.
func TestDasKontingentHeisstDrosselMitWartezeit(t *testing.T) {
	t.Parallel()

	befund := app.Uebersetze(context.Background(), domain.NeueKontingentgrenze(4500*time.Millisecond))

	if befund.Code != "drossel" || befund.Schluessel != domain.SchluesselKontingent ||
		befund.Werte["sekunden"] != "5" {
		t.Errorf("übersetzt zu %+v", befund)
	}
}
