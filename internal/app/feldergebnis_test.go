package app_test

// **Was das Fenster über das Gespeicherte erfährt** (#2535) — die Zweige, die
// die echte Naht in `fenstergespeichert_intern_test.go` nicht betritt: ein
// Verbund, dessen Wert niemand im Bestand findet, und ein Fehler auf dem Weg.

import (
	"context"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ergebnisschreiber sagt, was er gespeichert hätte — oder dass er es nicht weiß.
type ergebnisschreiber struct {
	*schreibprobe

	gespeichert string
	bekannt     bool
	fehler      error
}

func (e *ergebnisschreiber) FeldSetzenMitErgebnis(
	_ context.Context, _, karteID int64, feld, wert string,
) (gespeichert string, bekannt bool, fehler error) {
	e.karteID, e.feld, e.wert = karteID, feld, wert

	return e.gespeichert, e.bekannt, e.fehler
}

func TestFeldSetzenNenntDasGespeicherteNurWoBekannt(t *testing.T) {
	for _, fall := range []struct {
		name      string
		schreiber *ergebnisschreiber
		gewollt   string
		fehler    bool
	}{
		{name: "gerechnet", schreiber: &ergebnisschreiber{gespeichert: "[Glut]", bekannt: true}, gewollt: "[Glut]"},
		{name: "ein Verbund ist unbekannt", schreiber: &ergebnisschreiber{}},
		{name: "ein Fehler nennt nichts", schreiber: &ergebnisschreiber{
			gespeichert: "[Glut]", bekannt: true, fehler: fmt.Errorf("%w: inzwischen geändert", domain.ErrConflict),
		}, fehler: true},
	} {
		t.Run(fall.name, func(t *testing.T) {
			fall.schreiber.schreibprobe = &schreibprobe{}

			anwendung := app.NeueApp("")
			anwendung.MitEditor(fall.schreiber, karteprobe{}, service.NeuerPlanbauer(),
				render.NeuerRasterer(ohneBilder{}))

			antwort := anwendung.FeldSetzen(probekontext(t), "1", "7", "name", "Glut")

			if (antwort.Fehler != nil) != fall.fehler {
				t.Fatalf("Fehler %+v", antwort.Fehler)
			}

			switch {
			case fall.gewollt == "" && antwort.Gespeichert != nil:
				t.Errorf("gespeichert %q, erwartet nichts", *antwort.Gespeichert)
			case fall.gewollt != "" && (antwort.Gespeichert == nil || *antwort.Gespeichert != fall.gewollt):
				t.Errorf("gespeichert %v, erwartet %q", antwort.Gespeichert, fall.gewollt)
			}

			if fall.schreiber.wert != "Glut" {
				t.Errorf("geschrieben %+v", fall.schreiber)
			}
		})
	}
}
