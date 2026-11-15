package app_test

// **Der freiwillige Plan am Fenster** — OPS01
// (#2347):
// erst nach einer Ordnerwahl, dann sofort und danach wöchentlich.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// planfenster startet eine Anwendung, deren Dialog `ordner` wählt.
func planfenster(t *testing.T, ordner string) *app.App {
	t.Helper()

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) { return ordner, nil })

	bilder := t.TempDir()
	anwendung.MitBilderort(func() (string, error) { return bilder, nil })

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // nur zum Aufräumen

	return anwendung
}

// bisZumStand fragt den Stand, bis `gilt` stimmt — auf die Wirkung gewartet.
func bisZumStand(t *testing.T, anwendung *app.App, gilt func(app.SicherungsstandAntwort) bool) app.SicherungsstandAntwort {
	t.Helper()

	var antwort app.SicherungsstandAntwort

	for range 400 {
		if antwort = anwendung.Sicherungsstand(t.Context()); gilt(antwort) {
			return antwort
		}

		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("der Stand kam nicht: %+v", antwort)

	return antwort
}

func TestDerPlanSichertGleichNachDerOrdnerwahl(t *testing.T) {
	ordner := t.TempDir()
	anwendung := planfenster(t, ordner)

	if vorher := anwendung.Sicherungsstand(t.Context()); vorher.Stand.Stufe != "unbekannt" || vorher.Stand.Plan.Ordner != "" {
		t.Fatalf("vorher: %+v — der Plan ist freiwillig", vorher)
	}

	if eingerichtet := anwendung.SicherungsplanEinrichten(t.Context()); eingerichtet.Fehler != nil || eingerichtet.Stand.Plan.Ordner != ordner {
		t.Fatalf("eingerichtet: %+v", eingerichtet)
	}

	nachher := bisZumStand(t, anwendung, func(a app.SicherungsstandAntwort) bool { return a.Stand.Plan.Letzter != "" })
	if nachher.Stand.Stufe != "frisch" || nachher.Stand.Anlass != "plan" || nachher.Stand.Flaeche != "fenster" ||
		nachher.Stand.Plan.Faellig || nachher.Planfehler != nil {
		t.Errorf("nach dem ersten Lauf: %+v", nachher)
	}

	eintraege, err := os.ReadDir(ordner)
	if err != nil || len(eintraege) != 1 || !strings.HasPrefix(eintraege[0].Name(), "bestand-plan-") {
		t.Errorf("im Ordner %v (%v)", eintraege, err)
	}

	if beendet := anwendung.SicherungsplanBeenden(t.Context()); beendet.Fehler != nil || beendet.Stand.Plan.Ordner != "" {
		t.Errorf("beendet: %+v", beendet)
	}
}

// **Ein abgebrochener Dialog richtet nichts ein** — „nur nach Ordnerwahl".
func TestEinAbgebrochenerDialogRichtetKeinenPlanEin(t *testing.T) {
	anwendung := planfenster(t, "")

	if antwort := anwendung.SicherungsplanEinrichten(t.Context()); antwort.Fehler != nil || antwort.Stand.Plan.Ordner != "" {
		t.Errorf("abgebrochen: %+v", antwort)
	}
}
