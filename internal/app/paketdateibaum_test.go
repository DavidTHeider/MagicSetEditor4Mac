// **Der Dateibaum der Werkstatt am Fenster** — FMT-02.AK-01
// (#1691).
package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDasBeiwerkKommtAnsFenster(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(&werkstattprobe{dateien: []string{"style"}, beiwerk: []string{"rahmen.png"}})

	antwort := app.NeueBindungen(anwendung).Paketdateien(t.Context(), "probe.mse-style")
	if len(antwort.Beiwerk) != 1 || antwort.Beiwerk[0] != "rahmen.png" {
		t.Errorf("Beiwerk %v", antwort.Beiwerk)
	}

	// **Leer heißt leer** und nicht nil — im Fenster stünde sonst `null`.
	anwendung.MitPaketdateien(&werkstattprobe{})
	if leer := anwendung.Paketdateien(t.Context(), "p"); leer.Beiwerk == nil {
		t.Error("das Beiwerk ist nil statt leer")
	}
}

func TestDerDateibaumGehtDurchDieBindungen(t *testing.T) {
	anwendung := app.NeueApp("")
	probe := &werkstattprobe{stand: "neu"}
	anwendung.MitPaketdateien(probe)

	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.PaketdateiAnlegen(t.Context(), "p", "skripte/neu", "# leer\n"); antwort.Stand != "neu" ||
		probe.datei != "skripte/neu" || probe.wert != "# leer\n" {
		t.Errorf("Anlegen: %+v, %+v", antwort, probe)
	}

	if antwort := bindungen.PaketdateiUmbenennen(t.Context(), "p", "a", "b", "s"); antwort.Fehler != nil ||
		probe.datei != "a" || probe.neu != "b" || probe.erwartet != "s" {
		t.Errorf("Umbenennen: %+v, %+v", antwort, probe)
	}

	if antwort := bindungen.PaketdateiEntfernen(t.Context(), "p", "c", "t"); antwort.Fehler != nil ||
		probe.datei != "c" || probe.erwartet != "t" {
		t.Errorf("Entfernen: %+v, %+v", antwort, probe)
	}
}

// **Einfügen fragt selbst nach der Datei** — das Fenster reicht keinen Pfad
// herein, nur den Ordner im Paket, in den sie soll.
func TestEineDateiKommtUeberDenDialogInsPaket(t *testing.T) {
	anwendung := app.NeueApp("")
	probe := &werkstattprobe{}
	anwendung.MitPaketdateien(probe)
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "/Volumes/Karten/maske.png", nil
	})

	antwort := app.NeueBindungen(anwendung).PaketdateiEinfuegen(t.Context(), "p", "bilder")
	if antwort.Fehler != nil || antwort.Datei != "bilder/maske.png" ||
		probe.quelle != "/Volumes/Karten/maske.png" || probe.datei != "bilder/maske.png" {
		t.Errorf("Einfügen: %+v, %+v", antwort, probe)
	}

	// **Oben im Paket** ohne Ordner.
	if oben := anwendung.PaketdateiEinfuegen(t.Context(), "p", ""); oben.Datei != "maske.png" {
		t.Errorf("oben: %+v", oben)
	}
}

func TestEinAbgebrochenerDialogFuegtNichtsEin(t *testing.T) {
	anwendung := app.NeueApp("")
	probe := &werkstattprobe{}
	anwendung.MitPaketdateien(probe)
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) { return "", nil })

	antwort := anwendung.PaketdateiEinfuegen(t.Context(), "p", "")
	if !antwort.Abgebrochen || antwort.Fehler != nil || probe.quelle != "" {
		t.Errorf("abgebrochen: %+v, %+v", antwort, probe)
	}
}

func TestDieFehlerDesDateibaumsKommenAn(t *testing.T) {
	ohne := app.NeueApp("")

	mitFehler := app.NeueApp("")
	mitFehler.MitPaketdateien(&werkstattprobe{fehler: domain.ErrConflict})
	mitFehler.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) { return "/x.png", nil })

	dialogfehler := app.NeueApp("")
	dialogfehler.MitPaketdateien(&werkstattprobe{})
	dialogfehler.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", domain.ErrValidation
	})

	for name, anwendung := range map[string]*app.App{"ohne Naht": ohne, "mit Fehler": mitFehler} {
		ctx := t.Context()

		if anwendung.PaketdateiAnlegen(ctx, "p", "a", "").Fehler == nil ||
			anwendung.PaketdateiUmbenennen(ctx, "p", "a", "b", "").Fehler == nil ||
			anwendung.PaketdateiEntfernen(ctx, "p", "a", "").Fehler == nil ||
			anwendung.PaketdateiEinfuegen(ctx, "p", "").Fehler == nil {
			t.Errorf("%s: ein Handgriff kam ohne Fehler durch", name)
		}
	}

	if dialogfehler.PaketdateiEinfuegen(t.Context(), "p", "").Fehler == nil {
		t.Error("ein Dialogfehler kam nicht an")
	}

	if code := mitFehler.PaketdateiEntfernen(t.Context(), "p", "a", "").Fehler.Code; code != "konflikt" {
		t.Errorf("Code %q, erwartet konflikt", code)
	}
}
