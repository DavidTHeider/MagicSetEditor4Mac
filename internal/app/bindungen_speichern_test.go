// Die gebundene Oberfläche für das Speichern.
//
// **Eine eigene Datei, weil `bindungen_test.go` über die Zeilengrenze lief**
// (AGENTS.md §2). Der Schnitt folgt der Frage: hier steht, was die Oberfläche
// beim Zurückschreiben tut — und das ist genau eine Entscheidung, nämlich ob
// gefragt wird.
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieBindungenZumSpeichern(t *testing.T) {
	anwendung := app.NeueApp("")
	bindungen := app.NeueBindungen(anwendung)

	// **⌘S fragt nicht.** Es gibt einen Ort, den niemand nennen muss: den, aus
	// dem das Set kam. Ein leeres Ziel heißt genau das — der Dienst setzt ihn
	// ein, nicht die Oberfläche.
	t.Run("SetSpeichern", func(t *testing.T) {
		probe := &speicherprobe{speicherung: &service.Speicherung{Pfad: "/tmp/probe.mse-set"}}
		anwendung.MitSpeicherer(probe)

		gefragt := false

		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			gefragt = true

			return "", nil
		})

		antwort := bindungen.SetSpeichern(t.Context(), "7")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if gefragt {
			t.Error("es wurde nach einem Ort gefragt — das wäre „Speichern unter“")
		}

		if probe.setID != 7 || probe.ziel != "" {
			t.Errorf("beim Speicherer kam Set %d, Ziel %q an", probe.setID, probe.ziel)
		}
	})

	// **⌘S fragt doch, wenn das Set keinen Ort hat** — wie jedes Programm auf
	// dem Mac beim ersten Sichern (FIX-F02, S01.NAT-02). Bis zum 2026-09-26
	// stand dann „Der gesuchte Eintrag wurde nicht gefunden" da (Hüllenprobe).
	t.Run("SetSpeichern ohne Ort", func(t *testing.T) {
		anwendung.MitSpeicherer(ortlos{})

		var gewaehlt app.Dateiauswahl

		anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
			gewaehlt = auswahl

			return "/tmp/erstmals.mse-set", nil
		})

		antwort := bindungen.SetSpeichern(t.Context(), "7")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if !gewaehlt.Speichern || antwort.Pfad != "/tmp/erstmals.mse-set" {
			t.Errorf("gefragt wurde nach %+v, geschrieben nach %q", gewaehlt, antwort.Pfad)
		}
	})

	// **„Speichern unter" fragt — und zwar mit dem Sicherndialog.**
	t.Run("SetSpeichernUnter", func(t *testing.T) {
		probe := &speicherprobe{speicherung: &service.Speicherung{Pfad: "/tmp/woanders.mse-set"}}
		anwendung.MitSpeicherer(probe)

		var gewaehlt app.Dateiauswahl

		anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
			gewaehlt = auswahl

			return "/tmp/woanders.mse-set", nil
		})

		antwort := bindungen.SetSpeichernUnter(t.Context(), "7")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if probe.ziel != "/tmp/woanders.mse-set" {
			t.Errorf("beim Speicherer kam Ziel %q an", probe.ziel)
		}

		if !gewaehlt.Speichern || gewaehlt.Vorschlag == "" {
			t.Errorf("gefragt wurde nach %+v", gewaehlt)
		}
	})

	// **Ein abgebrochener Dialog schreibt nichts** — und ist kein Fehler.
	t.Run("ein abgebrochenes Speichern unter", func(t *testing.T) {
		probe := &speicherprobe{speicherung: &service.Speicherung{}}
		anwendung.MitSpeicherer(probe)
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			return "", nil
		})

		antwort := bindungen.SetSpeichernUnter(t.Context(), "7")
		if !antwort.Abgebrochen {
			t.Error("nicht als Abbruch gemeldet")
		}

		if probe.ziel != "" || probe.setID != 0 {
			t.Error("es wurde trotzdem geschrieben — und zwar an die Herkunft")
		}
	})

	// **Ein Dialog, der gar nicht aufgeht, ist ein Fehler** — und keiner, den
	// ein leerer Pfad verschluckt: der hieße hier „Abbruch".
	t.Run("ein Sicherndialog, der nicht aufgeht", func(t *testing.T) {
		probe := &speicherprobe{speicherung: &service.Speicherung{}}
		anwendung.MitSpeicherer(probe)
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			return "", errors.New("der Dialog ging nicht auf")
		})

		antwort := bindungen.SetSpeichernUnter(t.Context(), "7")
		if antwort.Fehler == nil {
			t.Fatal("der Fehler ging verloren")
		}

		if antwort.Abgebrochen {
			t.Error("als Abbruch gemeldet — dann sucht niemand nach der Ursache")
		}
	})
}

// ortlos ist ein Speicherer für ein Set, das aus keiner Datei kam: ohne Ziel
// weiß er nicht wohin, mit einem schreibt er dorthin.
type ortlos struct{}

func (ortlos) Schreibe(_ context.Context, _ int64, ziel string) (*service.Speicherung, error) {
	if ziel == "" {
		return nil, service.ErrOhneOrt
	}

	return &service.Speicherung{Pfad: ziel}, nil
}
