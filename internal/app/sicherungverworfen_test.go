package app_test

// **Eine verworfene Sicherung sagt es** — F05
// (#2344).
// Ohne eigenen Schlüssel käme sie als „ein unerwarteter Fehler" an, und wer
// sie liest, wüsste nicht, dass nichts gesichert ist.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineVerworfeneSicherungHatIhrenSatz(t *testing.T) {
	t.Parallel()

	fehler := app.Uebersetze(t.Context(), domain.MitGrund(domain.ErrSicherungVerworfen,
		errors.New("quick_check: Page 7: btreeInitPage() returns error code 11"), "noch einmal"))

	if fehler.Schluessel != domain.SchluesselSicherungVerworfen || fehler.Code != app.CodeValidierung {
		t.Errorf("übersetzt zu %+v", fehler)
	}
}

// **Eine Sicherung ohne Vermerk sagt, dass sie da ist** — OPS01 (#2347): der
// allgemeine Satz zu „belegt" sagte, es habe nicht geklappt, und wer ihn las,
// hätte eine gute Datei gelöscht.
func TestEineUnvermerkteSicherungHatIhrenSatz(t *testing.T) {
	t.Parallel()

	fehler := app.Uebersetze(t.Context(), domain.MitGrund(domain.ErrSicherungUnvermerkt, domain.ErrBelegt, "noch einmal"))

	if fehler.Schluessel != domain.SchluesselSicherungUnvermerkt || fehler.Code != app.CodeKonflikt {
		t.Errorf("übersetzt zu %+v", fehler)
	}
}
