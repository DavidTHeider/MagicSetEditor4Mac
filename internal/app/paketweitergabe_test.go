// **Weitergeben und Einlesen am Fenster** — FMT-02.AK-04
// (#1691).
//
// Beide fragen selbst nach dem Ort — das Fenster reicht keinen Pfad herein.
package app_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// weitergabeprobe gibt weiter und liest ein — oder scheitert.
type weitergabeprobe struct {
	paketfuehrer

	namen      []string
	ziel       string
	quelle     string
	uebernahme *service.Uebernahme
	fehler     error
}

func (p *weitergabeprobe) Weitergeben(_ context.Context, namen []string, ziel string) (string, error) {
	p.namen, p.ziel = namen, ziel

	return ziel, p.fehler
}

func (p *weitergabeprobe) Einlesen(_ context.Context, quelle string) (*service.Uebernahme, error) {
	p.quelle = quelle

	return p.uebernahme, p.fehler
}

// mitDialog setzt eine Dateiwahl, die antwortet und sich merkt, was sie
// gefragt wurde.
func mitDialog(anwendung *app.App, antwort string, fehler error) *app.Dateiauswahl {
	gefragt := &app.Dateiauswahl{}

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		*gefragt = auswahl

		return antwort, fehler
	})

	return gefragt
}

func TestEinPaketGehtUeberDenSicherndialogHinaus(t *testing.T) {
	anwendung := app.NeueApp("")
	probe := &weitergabeprobe{}
	anwendung.MitPaketen(probe)
	gefragt := mitDialog(anwendung, "/Volumes/Karten/probe.mse-installer", nil)

	antwort := app.NeueBindungen(anwendung).PaketWeitergeben(t.Context(), "probe.mse-game")
	if antwort.Fehler != nil || antwort.Abgebrochen || antwort.Datei != "/Volumes/Karten/probe.mse-installer" {
		t.Fatalf("Antwort %+v", antwort)
	}

	if !slices.Equal(probe.namen, []string{"probe.mse-game"}) || probe.ziel != antwort.Datei {
		t.Errorf("weitergegeben %v nach %q", probe.namen, probe.ziel)
	}

	// **Der Dialog schlägt den Namen vor, den MSE2 wählte** — das Paket ohne
	// Endung, und nimmt nur Installer.
	if !gefragt.Speichern || gefragt.Vorschlag != "probe.mse-installer" || gefragt.Muster != "*.mse-installer" {
		t.Errorf("gefragt %+v", *gefragt)
	}
}

func TestEinAbgebrochenerSicherndialogGibtNichtsWeiter(t *testing.T) {
	anwendung := app.NeueApp("")
	probe := &weitergabeprobe{}
	anwendung.MitPaketen(probe)
	mitDialog(anwendung, "", nil)

	antwort := anwendung.PaketWeitergeben(t.Context(), "probe.mse-game")
	if !antwort.Abgebrochen || antwort.Fehler != nil || probe.namen != nil {
		t.Errorf("Antwort %+v, weitergegeben %v", antwort, probe.namen)
	}
}

func TestEinInstallerKommtUeberDenOeffnendialogHerein(t *testing.T) {
	anwendung := app.NeueApp("")
	probe := &weitergabeprobe{uebernahme: &service.Uebernahme{
		Uebernommen: []domain.Paket{{Art: "mse-include", Name: "probe-mitgebracht", Pfad: "/pakete/probe-mitgebracht.mse-include"}},
		Uebergangen: []string{"probe.mse-game"},
		Befunde:     []domain.Befund{{Datei: "probe-kaputt.mse-game", Meldung: "das Paket ließ sich nicht auspacken"}},
	}}
	anwendung.MitPaketen(probe)
	gefragt := mitDialog(anwendung, "/Volumes/Karten/probe.mse-installer", nil)

	antwort := app.NeueBindungen(anwendung).InstallerEinlesen(t.Context())
	if antwort.Fehler != nil || len(antwort.Uebernommen) != 1 || antwort.Uebernommen[0].Name != "probe-mitgebracht" ||
		!slices.Equal(antwort.Uebergangen, []string{"probe.mse-game"}) || len(antwort.Befunde) != 1 {
		t.Fatalf("Antwort %+v", antwort)
	}

	if probe.quelle != "/Volumes/Karten/probe.mse-installer" || gefragt.Speichern || gefragt.Muster != "*.mse-installer" {
		t.Errorf("Quelle %q, gefragt %+v", probe.quelle, *gefragt)
	}
}

func TestAbbruchUndFehlerBeimEinlesenUndWeitergeben(t *testing.T) {
	ctx := t.Context()

	abgebrochen := app.NeueApp("")
	abgebrochen.MitPaketen(&weitergabeprobe{})
	mitDialog(abgebrochen, "", nil)

	if antwort := abgebrochen.InstallerEinlesen(ctx); !antwort.Abgebrochen || antwort.Fehler != nil || antwort.Uebernommen == nil {
		t.Errorf("abgebrochen: %+v", antwort)
	}

	dialogfehler := app.NeueApp("")
	dialogfehler.MitPaketen(&weitergabeprobe{})
	mitDialog(dialogfehler, "", domain.ErrValidation)

	if dialogfehler.InstallerEinlesen(ctx).Fehler == nil || dialogfehler.PaketWeitergeben(ctx, "p.mse-game").Fehler == nil {
		t.Error("ein Dialogfehler kam nicht an")
	}

	// **Ein Teilerfolg bleibt sichtbar**: was eingelesen ist, steht da, auch
	// wenn der Abgleich danach scheiterte (§7, Regel 3).
	teilweise := app.NeueApp("")
	teilweise.MitPaketen(&weitergabeprobe{
		uebernahme: &service.Uebernahme{Uebernommen: []domain.Paket{{Art: "mse-game", Name: "a"}}},
		fehler:     errors.New("der Bestand ist gesperrt"),
	})
	mitDialog(teilweise, "/x.mse-installer", nil)

	if antwort := teilweise.InstallerEinlesen(ctx); antwort.Fehler == nil || len(antwort.Uebernommen) != 1 {
		t.Errorf("teilweise: %+v", antwort)
	}

	if antwort := teilweise.PaketWeitergeben(ctx, "a.mse-game"); antwort.Fehler == nil || antwort.Datei != "" {
		t.Errorf("gescheitert: %+v", antwort)
	}
}
