// **Der Rückfall der Uhr** — Werkstück 3 von Phase 3.4
// (#1165).
package app

import (
	"testing"
	"time"
)

// **Ein `App` als Nullwert hat eine Uhr.**
//
// # Warum das eigens dasteht
//
// [NeueApp] setzt sie. Ein `App{}` als Nullwert kommt in Läufen trotzdem vor,
// und ein Aufruf über eine `nil`-Uhr wäre ein Absturz statt einer Antwort —
// genau die Klasse, an der in Phase 3.3 eine Kartenvorschau hing.
func TestEinAppAlsNullwertHatEineUhr(t *testing.T) {
	t.Parallel()

	var anwendung App

	if anwendung.jetzt().IsZero() {
		t.Error("die Uhr gibt den Nullzeitpunkt")
	}
}

// **Und eine gesetzte Uhr gilt.**
func TestEineGesetzteUhrGilt(t *testing.T) {
	t.Parallel()

	fest := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)

	var anwendung App

	anwendung.MitUhr(func() time.Time { return fest })

	if !anwendung.jetzt().Equal(fest) {
		t.Errorf("die Uhr zeigt %v", anwendung.jetzt())
	}
}
