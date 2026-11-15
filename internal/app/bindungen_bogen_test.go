// Die gebundene Oberfläche für den Druckbogen.
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieBindungZumDruckbogen(t *testing.T) {
	anwendung := app.NeueApp("")
	bindungen := app.NeueBindungen(anwendung)

	t.Run("BogenLegen", func(t *testing.T) {
		probe := &bogenprobe{lauf: &service.Bogenlauf{Pfad: "/tmp/bogen.pdf", Seiten: 1}}
		anwendung.MitBogenleger(probe)

		var gewaehlt app.Dateiauswahl

		anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
			gewaehlt = auswahl

			return "/tmp/bogen.pdf", nil
		})

		antwort := bindungen.BogenLegen(t.Context(), "7", "a4", 3, 300, true, nil)
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if probe.pfad != "/tmp/bogen.pdf" || probe.masse.Anschnitt != 3 {
			t.Errorf("beim Leger kam %q, %+v an", probe.pfad, probe.masse)
		}

		// **Ein Sicherndialog mit Vorschlag** — wer eine Datei anlegt, sucht
		// keinen vorhandenen Namen aus.
		if !gewaehlt.Speichern || gewaehlt.Vorschlag == "" || gewaehlt.Muster != "*.pdf" {
			t.Errorf("gefragt wurde nach %+v", gewaehlt)
		}
	})

	t.Run("ein abgebrochener Dialog legt nichts", func(t *testing.T) {
		probe := &bogenprobe{lauf: &service.Bogenlauf{}}
		anwendung.MitBogenleger(probe)
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			return "", nil
		})

		antwort := bindungen.BogenLegen(t.Context(), "7", "a4", 0, 300, true, nil)
		if !antwort.Abgebrochen {
			t.Error("nicht als Abbruch gemeldet")
		}

		if probe.pfad != "" {
			t.Error("es wurde trotzdem gelegt")
		}
	})

	t.Run("ein Dialog, der nicht aufgeht", func(t *testing.T) {
		anwendung.MitBogenleger(&bogenprobe{lauf: &service.Bogenlauf{}})
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			return "", errors.New("der Dialog ging nicht auf")
		})

		antwort := bindungen.BogenLegen(t.Context(), "7", "a4", 0, 300, true, nil)
		if antwort.Fehler == nil {
			t.Fatal("der Fehler ging verloren")
		}

		if antwort.Abgebrochen {
			t.Error("als Abbruch gemeldet — dann sucht niemand nach der Ursache")
		}
	})
}
