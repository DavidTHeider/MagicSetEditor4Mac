package service_test

// **Die Grenzen des Kostenwächters** — Werkstück O12 von Phase 5.2
// (#1439).
//
// **Eine eigene Datei**, weil `kostenwache_test.go` über die Zeilengrenze aus
// §2 ging. Der Schnitt folgt der Frage: dort steht, wie ein Auftrag **läuft**
// — hier, woran er **nicht** vorbeikommt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein erfundener Zeitraum wird an jedem Weg abgewiesen** — statt still zum
// Tag zu werden.
//
// **Vier Wege und nicht einer**: jeder rechnet über ein Fenster, und jeder
// hätte ohne diese Prüfung ein anderes genommen als der Aufrufer meinte.
func TestEinErfundenerZeitraumWirdAnJedemWegAbgewiesen(t *testing.T) {
	t.Parallel()

	weg := kostenwache(&kostendoppel{}, wache())

	if _, fehler := weg.Stand(t.Context(), "anbieter", "stunde"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("der Stand: %v", fehler)
	}

	if _, fehler := weg.Posten(t.Context(), "anbieter", "stunde"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("die Posten: %v", fehler)
	}

	if fehler := weg.BudgetSetzen(t.Context(), "anbieter", "stunde",
		domain.Betrag{Wert: 500, Waehrung: "EUR"}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("das Budget: %v", fehler)
	}
}

// **Eine Absage der Wache hält auch das Setzen einer Grenze an.**
func TestEineAbsageHaeltDasSetzenDerGrenzeAn(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{}
	wacht := wache()
	wacht.absage = domain.ErrErlaubnis

	if fehler := kostenwache(doppel, wacht).BudgetSetzen(t.Context(), "anbieter",
		domain.ZeitraumTag, domain.Betrag{Wert: 500, Waehrung: "EUR"}); !errors.Is(
		fehler, domain.ErrErlaubnis) {
		t.Fatalf("mit Absage: %v", fehler)
	}

	if doppel.grenze.Wert != 0 {
		t.Error("gesetzt wurde trotzdem")
	}
}
