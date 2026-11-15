package app_test

// **Eine Sicherungsauswahl nennt ihre Bilderlage auch ohne Datei** — C01
// (#2226).
//
// Der Vertragsrundgang fand `bilder: ""` nach einem abgebrochenen Dialog, wo
// das Fenster eine von drei Lagen erwartet (`antwortenbestand.ts`). Ohne
// gewählte Datei geschieht mit den Bildern nichts: `unberuehrt` — dieselbe
// Pflicht, aus der die Listen derselben Antwort leer und nicht `null` sind.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestOhneGewaehlteDateiBleibenDieBilderUnberuehrt(t *testing.T) {
	for _, fall := range []struct {
		name   string
		dialog func(context.Context, app.Dateiauswahl) (string, error)
		pfad   string
	}{
		{"der Dialog wird abgebrochen", func(context.Context, app.Dateiauswahl) (string, error) { return "", nil }, ""},
		{"der Dialog scheitert", func(context.Context, app.Dateiauswahl) (string, error) {
			return "", errors.New("das Fenster antwortet nicht")
		}, ""},
	} {
		t.Run(fall.name, func(t *testing.T) {
			anwendung := app.NeueApp("")
			anwendung.MitDateiwahl(fall.dialog)

			if antwort := anwendung.SicherungWaehlen(t.Context()); antwort.Bilder != "unberuehrt" {
				t.Errorf("die Bilderlage ist %q", antwort.Bilder)
			}
		})
	}

	// **Und eine Datei, die nicht daliegt**, ebenso — und sie ist **nicht
	// gefunden**, kein unerwarteter Fehler (bis C01 kam der rohe Fehler des
	// Systems an).
	anwendung := app.NeueApp("")
	antwort := anwendung.SicherungBeschauen(t.Context(), "/gibt/es/nicht.msebackup", "")
	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNichtGefunden || antwort.Bilder != "unberuehrt" {
		t.Errorf("Fehler %+v, Bilderlage %q", antwort.Fehler, antwort.Bilder)
	}
}
