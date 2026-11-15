// Die Bindungen der Exportvorlagen.
//
// **Eine eigene Datei**, weil `bindungen_test.go` sonst über die Zeilengrenze
// wüchse (AGENTS.md §2). Der Schnitt folgt der Frage: dort die Bindungen der
// Anwendung insgesamt, hier die zwei, die
// #209
// hinzugefügt hat.
package app_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDieBindungenDerExportvorlagen(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
		ausgabe:  "fertig",
	})

	bindungen := app.NeueBindungen(anwendung)

	t.Run("Exportvorlagen", func(t *testing.T) {
		anwendung.MitVorlagen(&vorlagenprobe{
			vorlagen: []msepaket.Exportvorlage{probevorlage()},
		})

		if antwort := bindungen.Exportvorlagen(t.Context(), "1"); len(antwort.Vorlagen) != 1 {
			t.Errorf("%+v", antwort)
		}
	})

	t.Run("VorlageAusfuehren", func(t *testing.T) {
		anwendung.MitVorlagen(&vorlagenprobe{
			vorlagen: []msepaket.Exportvorlage{probevorlage()},
			ausgabe:  "fertig",
		})

		ziel := filepath.Join(t.TempDir(), "ausfuhr.txt")
		anwendung.MitDateiwahl(func(_ context.Context, wahl app.Dateiauswahl) (string, error) {
			// **Der Filter kommt aus der Vorlage** (#209): `*.txt|*.txt` ergibt
			// `*.txt`, und der Vorschlag trägt dieselbe Endung.
			if wahl.Muster != "*.txt" || !strings.HasSuffix(wahl.Vorschlag, ".txt") {
				t.Errorf("%+v", wahl)
			}

			return ziel, nil
		})

		if antwort := bindungen.VorlageAusfuehren(t.Context(), "1", "probe.mse-export-template",
			"*.txt|*.txt", nil, nil); antwort.Pfad != ziel {
			t.Errorf("%+v", antwort)
		}
	})

	t.Run("VorlageAusfuehren meldet einen Dialogfehler", func(t *testing.T) {
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			return "", errors.New("kein Dialog")
		})

		if antwort := bindungen.VorlageAusfuehren(t.Context(), "1", "p", "", nil, nil); antwort.Fehler == nil {
			t.Error("kein Fehler")
		}
	})
}
