// Die gebundene Oberfläche für die Bildausgabe.
//
// **Eine eigene Datei aus demselben Grund wie `bindungen_speichern_test.go`**
// (AGENTS.md §2): der Schnitt folgt der Frage, und hier steht genau eine — was
// die Oberfläche tut, wenn jemand Bilder ausgeben will.
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieBindungenZurBildausgabe(t *testing.T) {
	anwendung := app.NeueApp("")
	bindungen := app.NeueBindungen(anwendung)

	// **Ein Ordner, keine Datei.** Bilder kommen als viele Dateien heraus; wer
	// nach einem Dateinamen fragte, bekäme einen und bräuchte dreihundert.
	t.Run("Bildordner", func(t *testing.T) {
		var gewaehlt app.Dateiauswahl

		anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
			gewaehlt = auswahl

			return "/tmp/bilder", nil
		})

		antwort := bindungen.Bildordner(t.Context())
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Ordner != "/tmp/bilder" {
			t.Errorf("Ordner: %q", antwort.Ordner)
		}

		if !gewaehlt.Verzeichnis || gewaehlt.Titel == "" {
			t.Errorf("gefragt wurde nach %+v", gewaehlt)
		}
	})

	t.Run("ein abgebrochener Ordnerdialog", func(t *testing.T) {
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			return "", nil
		})

		if antwort := bindungen.Bildordner(t.Context()); !antwort.Abgebrochen {
			t.Error("nicht als Abbruch gemeldet")
		}
	})

	t.Run("ein Ordnerdialog, der nicht aufgeht", func(t *testing.T) {
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			return "", errors.New("der Dialog ging nicht auf")
		})

		if antwort := bindungen.Bildordner(t.Context()); antwort.Fehler == nil {
			t.Error("der Fehler ging verloren")
		}
	})

	t.Run("Bildnamen", func(t *testing.T) {
		probe := &bildprobe{namen: []service.Bildname{{KarteID: 1, Nummer: 1, Datei: "001.png"}}}
		anwendung.MitBildausfuhr(probe)

		if antwort := bindungen.Bildnamen(t.Context(), "7", "{nummer}", "png", domain.BildkonfliktNummerierenNeu, nil); len(antwort.Namen) != 1 {
			t.Errorf("%+v", antwort)
		}
	})

	t.Run("BildSchreiben", func(t *testing.T) {
		probe := &bildprobe{schreibung: &service.Bildschreibung{Pfad: "/tmp/bilder/001.png"}}
		anwendung.MitBildausfuhr(probe)

		antwort := bindungen.BildSchreiben(t.Context(), "1", "/tmp/bilder", "001.png", "png", 750, domain.BildkonfliktUeberschreiben, "")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if probe.zielbreite != 750 || probe.datei != "001.png" {
			t.Errorf("beim Dienst kam %q, %d an", probe.datei, probe.zielbreite)
		}
	})
}
