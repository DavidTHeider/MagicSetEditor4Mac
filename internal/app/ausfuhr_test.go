package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ausfuhrprobe ist ein Double des Exporteurs.
type ausfuhrprobe struct {
	ausfuhr *service.Ausfuhr
	fehler  error
	setID   int64
	spiel   string
	pfad    string
}

func (a *ausfuhrprobe) MTGJSON(
	_ context.Context, setID int64, spiel, pfad string,
) (*service.Ausfuhr, error) {
	a.setID, a.spiel, a.pfad = setID, spiel, pfad

	return a.ausfuhr, a.fehler
}

func mitExporteur(probe *ausfuhrprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitExporteur(probe)

	return anwendung
}

// Ein Set geht als MTG JSON an den gewählten Ort.
func TestMTGJSONSchreibenGehtAnDenGewaehltenOrt(t *testing.T) {
	probe := &ausfuhrprobe{ausfuhr: &service.Ausfuhr{Pfad: "/tmp/PRB.json", Karten: 3}}

	antwort := mitExporteur(probe).MTGJSONSchreiben(probekontext(t), "7", "magic", "/tmp/PRB.json")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if probe.setID != 7 || probe.spiel != "magic" || probe.pfad != "/tmp/PRB.json" {
		t.Errorf("beim Exporteur kam %d, %q, %q an", probe.setID, probe.spiel, probe.pfad)
	}

	if antwort.Karten != 3 || antwort.Pfad != "/tmp/PRB.json" {
		t.Errorf("%+v", antwort)
	}
}

// Was nicht abbildbar war, kommt in der Antwort an.
func TestMTGJSONSchreibenNenntDieVerluste(t *testing.T) {
	probe := &ausfuhrprobe{ausfuhr: &service.Ausfuhr{
		Karten:  1,
		Befunde: []domain.Befund{{Ort: "PRB → super type", Meldung: "MSE führt beides in einem Feld"}},
	}}

	antwort := mitExporteur(probe).MTGJSONSchreiben(probekontext(t), "7", "magic", "/tmp/PRB.json")

	if len(antwort.Befunde) != 1 {
		t.Fatalf("%+v", antwort.Befunde)
	}

	if antwort.Befunde[0] != "PRB → super type: MSE führt beides in einem Feld" {
		t.Errorf("%q", antwort.Befunde[0])
	}
}

// Ein leerer Pfad ist der Abbruch des Speicherdialogs — kein Fehler.
func TestMTGJSONSchreibenMitLeeremPfadIstEinAbbruch(t *testing.T) {
	antwort := mitExporteur(&ausfuhrprobe{}).MTGJSONSchreiben(probekontext(t), "7", "magic", "")

	if antwort.Fehler != nil || !antwort.Abgebrochen {
		t.Errorf("%+v", antwort)
	}
}

// Befunde kommen als übersetzte Fehler an.
func TestMTGJSONSchreibenMeldetBefunde(t *testing.T) {
	faelle := map[string]struct {
		setID string
		probe *ausfuhrprobe
	}{
		"eine Kennung, die keine ist": {setID: "keine Zahl", probe: &ausfuhrprobe{}},
		"das Set gibt es nicht": {setID: "7", probe: &ausfuhrprobe{
			fehler: domain.ErrNotFound,
		}},
		"der Bestand streikt": {setID: "7", probe: &ausfuhrprobe{
			fehler: errors.New("die Platte ist voll"),
		}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			antwort := mitExporteur(fall.probe).
				MTGJSONSchreiben(probekontext(t), fall.setID, "magic", "/tmp/x.json")
			if antwort.Fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}
