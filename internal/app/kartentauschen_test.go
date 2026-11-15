package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// tauschprobe merkt sich, welche beiden Karten getauscht wurden.
type tauschprobe struct {
	eine, andere int64
	fehler       error
}

func (t *tauschprobe) KartenTauschen(_ context.Context, _, eineID, andereID int64) error {
	t.eine, t.andere = eineID, andereID

	return t.fehler
}

func mitTauscher(probe *tauschprobe) *app.App {
	anwendung := app.NeueApp(":memory:")
	anwendung.MitKartentauscher(probe)

	return anwendung
}

func TestKartenTauschenReichtBeideKennungenDurch(t *testing.T) {
	probe := &tauschprobe{}

	if antwort := mitTauscher(probe).KartenTauschen(probekontext(t), "1", "7", "9"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.eine != 7 || probe.andere != 9 {
		t.Errorf("beim Speicher kamen %d und %d an, erwartet 7 und 9", probe.eine, probe.andere)
	}
}

// **Beide Kennungen werden geprüft**, nicht nur die erste — sonst käme eine
// unlesbare zweite als 0 im Speicher an und träfe dort keine Karte.
func TestKartenTauschenWeistUnlesbareKennungenAb(t *testing.T) {
	for name, paar := range map[string][2]string{
		"die erste":  {"abc", "9"},
		"die zweite": {"7", "abc"},
	} {
		t.Run(name, func(t *testing.T) {
			befund := mitTauscher(&tauschprobe{}).
				KartenTauschen(probekontext(t), "1", paar[0], paar[1]).Fehler
			if befund == nil {
				t.Error("kein Befund auf eine unlesbare Kennung")
			}
		})
	}
}

func TestKartenTauschenMeldetEinenFehler(t *testing.T) {
	befund := mitTauscher(&tauschprobe{fehler: domain.ErrValidation}).
		KartenTauschen(probekontext(t), "1", "7", "9").Fehler
	if befund == nil || befund.Code != "validierung" {
		t.Errorf("%+v, erwartet validierung", befund)
	}
}
