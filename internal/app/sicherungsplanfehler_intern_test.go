package app

// **Ein Plan, der scheitert, sagt es** — OPS01
// (#2347):
// am Fenster und nur dort, denn nur das Fenster führt ihn aus.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Stick ist beim Start nicht da** — der Plan wird fällig, findet seinen
// Ordner nicht, legt keinen an und meldet es. Ein neuer Ordner nimmt die
// Meldung weg.
func TestEinFehlenderPlanordnerMeldetSichNachDemStart(t *testing.T) { //nolint:paralleltest // der Verzug ist eine Paketvariable
	bestand := filepath.Join(t.TempDir(), "bestand.db")
	stick := filepath.Join(t.TempDir(), "Stick")

	if err := os.Mkdir(stick, 0o700); err != nil {
		t.Fatal(err)
	}

	erste := NeueApp(bestand)
	if err := erste.Starten(t.Context()); err != nil {
		t.Fatal(err)
	}

	if antwort := erste.EinstellungSetzen(t.Context(), domain.SicherungsplanOrdner, stick); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if err := erste.Beenden(); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(stick); err != nil {
		t.Fatal(err)
	}

	vorher := planverzug
	planverzug = time.Millisecond

	t.Cleanup(func() { planverzug = vorher })

	zweite := NeueApp(bestand)
	bilder := t.TempDir()
	zweite.MitBilderort(func() (string, error) { return bilder, nil })

	if err := zweite.Starten(t.Context()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = zweite.Beenden() }) //nolint:errcheck // nur zum Aufräumen

	var antwort SicherungsstandAntwort

	for range 400 {
		if antwort = zweite.Sicherungsstand(t.Context()); antwort.Planfehler != nil {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if antwort.Planfehler == nil || antwort.Planfehler.Code != CodeNichtGefunden || !antwort.Stand.Plan.Faellig {
		t.Fatalf("gemeldet: %+v", antwort)
	}

	if _, err := os.Stat(stick); err == nil {
		t.Error("der Plan legte den Ordner an")
	}

	zweite.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) { return t.TempDir(), nil })

	if neu := zweite.SicherungsplanEinrichten(t.Context()); neu.Planfehler != nil {
		t.Errorf("nach einem neuen Ordner: %+v", neu.Planfehler)
	}
}
