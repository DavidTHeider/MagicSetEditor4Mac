package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// loeschprobe merkt sich, welche Karte gelöscht wurde.
type loeschprobe struct {
	karteID int64
	fehler  error
}

func (l *loeschprobe) KarteLoeschen(_ context.Context, _, karteID int64, _ bool) error {
	l.karteID = karteID

	return l.fehler
}

// **Eine Karte lässt sich löschen** —
// #621.
//
// # Warum das erst jetzt geht
//
// #562 hat es
// ausdrücklich vertagt: *„eine gelöschte Karte ohne ⌘Z ist Datenverlust mit
// einem Klick."* Der Verlauf trug nur Feldänderungen; seit er eine eigene Art
// für ganze Karten hat, ist der Weg zurück da.
func TestKarteLoeschenReichtDieKennungDurch(t *testing.T) {
	t.Parallel()

	probe := &loeschprobe{}

	anwendung := app.NeueApp("")
	anwendung.MitKartenloescher(probe)

	if fehler := anwendung.KarteLoeschen(t.Context(), "1", "42", false).Fehler; fehler != nil {
		t.Fatalf("%+v", fehler)
	}

	if probe.karteID != 42 {
		t.Errorf("gelöscht wurde %d, erwartet 42", probe.karteID)
	}
}

// **Eine unbrauchbare Kennung ist ein Befund, kein Aufruf.**
func TestKarteLoeschenWeistEineUnbrauchbareKennungAb(t *testing.T) {
	t.Parallel()

	probe := &loeschprobe{}

	anwendung := app.NeueApp("")
	anwendung.MitKartenloescher(probe)

	if fehler := anwendung.KarteLoeschen(t.Context(), "1", "keine Zahl", false).Fehler; fehler == nil {
		t.Error("kein Befund")
	}

	if probe.karteID != 0 {
		t.Errorf("es wurde gelöscht (%d), obwohl die Kennung unbrauchbar ist", probe.karteID)
	}
}

// **Und ein Befund des Speichers wird übersetzt**, nicht durchgereicht (§8).
func TestKarteLoeschenUebersetztDenBefund(t *testing.T) {
	t.Parallel()

	probe := &loeschprobe{fehler: domain.ErrNotFound}

	anwendung := app.NeueApp("")
	anwendung.MitKartenloescher(probe)

	fehler := anwendung.KarteLoeschen(t.Context(), "1", "42", false).Fehler
	if fehler == nil {
		t.Fatal("kein Befund")
	}

	if fehler.Code != app.CodeNichtGefunden {
		t.Errorf("Code %q, erwartet %q", fehler.Code, app.CodeNichtGefunden)
	}

	_ = errors.Is(domain.ErrNotFound, domain.ErrNotFound)
}
