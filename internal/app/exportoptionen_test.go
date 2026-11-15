package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// optionsprobe ist die Naht zur Ablage der Exportoptionen.
type optionsprobe struct {
	werte  map[string]map[string]string
	fehler error
}

func (o *optionsprobe) Fuer(_ context.Context, vorlage string) (map[string]string, error) {
	if o.fehler != nil {
		return nil, o.fehler
	}

	return o.werte[vorlage], nil
}

func (o *optionsprobe) Setzen(_ context.Context, vorlage, feld, wert string) error {
	if o.fehler != nil {
		return o.fehler
	}

	if o.werte == nil {
		o.werte = map[string]map[string]string{}
	}

	if o.werte[vorlage] == nil {
		o.werte[vorlage] = map[string]string{}
	}

	o.werte[vorlage][feld] = wert

	return nil
}

// **Ein behaltener Wert erreicht die Vorlagenliste** (#492).
//
// Die Oberfläche liest ihn dort und nicht aus einem zweiten Aufruf: eine
// Vorlage und ihre Werte gehören zusammen, und zwei Antworten liefen beim
// nächsten Wechsel auseinander.
func TestEineBehalteneOptionStehtAnDerVorlage(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
	})
	anwendung.MitExportoptionen(&optionsprobe{
		werte: map[string]map[string]string{
			"probe.mse-export-template": {"mit namen": "no"},
		},
	})

	antwort := app.NeueBindungen(anwendung).Exportvorlagen(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, vorlage := range antwort.Vorlagen {
		for _, option := range vorlage.Optionen {
			if option.Name != "mit namen" {
				continue
			}

			if option.Wert != "no" {
				t.Errorf("%+v", option)
			}

			return
		}
	}

	t.Errorf("die Option kam nicht an: %+v", antwort.Vorlagen)
}

// **Ohne behaltenen Wert gilt die Vorgabe der Vorlage.**
func TestOhneBehalteneOptionGiltDieVorgabe(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
	})
	anwendung.MitExportoptionen(&optionsprobe{})

	antwort := app.NeueBindungen(anwendung).Exportvorlagen(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, vorlage := range antwort.Vorlagen {
		for _, option := range vorlage.Optionen {
			if option.Wert != option.Vorgabe {
				t.Errorf("%+v", option)
			}
		}
	}
}

// **Und eine Änderung wird geschrieben** — je Vorlage.
func TestEineExportoptionLaesstSichSetzen(t *testing.T) {
	probe := &optionsprobe{}
	anwendung := mitVorlagen(t, &vorlagenprobe{})
	anwendung.MitExportoptionen(probe)

	antwort := app.NeueBindungen(anwendung).
		ExportoptionSetzen(t.Context(), "v.mse-export-template", "forum", "mtgsalvation")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.werte["v.mse-export-template"]["forum"] != "mtgsalvation" {
		t.Errorf("%v", probe.werte)
	}
}

func TestEinSchreibfehlerDerOptionErreichtDasFenster(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{})
	anwendung.MitExportoptionen(&optionsprobe{fehler: errors.New("die Platte schweigt")})

	if antwort := app.NeueBindungen(anwendung).
		ExportoptionSetzen(t.Context(), "v", "f", "w"); antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Ein Lesefehler nimmt die Vorlagenliste nicht weg** (§8).
//
// Die Vorlagen sind der Grund, warum jemand hier ist; die behaltenen Werte
// sind eine Zugabe. Wer die Liste wegen ihrer Bequemlichkeit verschwiegen
// bekäme, könnte gar nichts ausgeben.
func TestEinLesefehlerNimmtDieVorlagenNichtWeg(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
	})
	anwendung.MitExportoptionen(&optionsprobe{fehler: errors.New("die Platte schweigt")})

	antwort := app.NeueBindungen(anwendung).Exportvorlagen(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Vorlagen) == 0 {
		t.Error("keine Vorlage")
	}

	if len(antwort.Befunde) == 0 {
		t.Error("und kein Befund — der Ausfall wäre stillschweigend")
	}
}
