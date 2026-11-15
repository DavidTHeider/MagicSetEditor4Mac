package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Übernahme aus einer MSE2-Installation (#55).

// **Vorgeschlagen wird nur, was es gibt.**
//
// Ein Vorschlag auf ein Verzeichnis, das nicht existiert, wäre eine
// Aufforderung, es zu suchen — und genau das soll die Anwendung abnehmen.
func TestDieVorschlaegeNennenNurVorhandeneOrte(t *testing.T) {
	anwendung := &app.App{}
	anwendung.MitMSE2Orten(func() []string { return []string{"/mse2/data"} })

	if antwort := anwendung.MSE2Vorschlaege(); len(antwort.Orte) != 1 {
		t.Errorf("Orte: %v", antwort.Orte)
	}
}

// **`MSE2Orte` statet und liest nicht.**
//
// Der Weg auf diesem Rechner: ein erfundenes Heimatverzeichnis, ein
// vorhandener und ein fehlender Ort. Ohne den Test bliebe offen, ob die Liste
// überhaupt prüft — eine, die alle fünf Pfade zurückgäbe, sähe von außen
// genauso aus.
func TestMSE2OrteMeldetNurWasDasteht(t *testing.T) {
	heim := t.TempDir()
	t.Setenv("HOME", heim)

	if orte := app.MSE2Orte(); len(orte) != 0 {
		t.Fatalf("ohne Installation: %v", orte)
	}

	echter := filepath.Join(heim, ".wine", "drive_c", "Program Files (x86)",
		"Magic Set Editor 2", "data")
	if err := os.MkdirAll(echter, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	orte := app.MSE2Orte()
	if len(orte) != 1 || orte[0] != echter {
		t.Errorf("Orte: %v", orte)
	}
}

// **Ohne Heimatverzeichnis keine Vermutung.**
//
// Dann bleibt der Weg „Verzeichnis wählen …" — ehrlicher als eine Liste
// relativer Pfade, die auf das Arbeitsverzeichnis des Prozesses zeigten.
func TestOhneHeimatverzeichnisKeineVorschlaege(t *testing.T) {
	t.Setenv("HOME", "")

	if orte := app.MSE2Orte(); len(orte) != 0 {
		t.Errorf("Orte: %v", orte)
	}
}

// **Ein leerer Ort heißt „abgebrochen" und nicht „Fehler".**
func TestUebernehmenOhneOrtIstEinAbbruch(t *testing.T) {
	fuehrer := &paketfuehrer{}
	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	antwort := anwendung.VorlagenUebernehmen(t.Context(), "  ")

	if !antwort.Abgebrochen || antwort.Fehler != nil {
		t.Errorf("Antwort: %+v", antwort)
	}

	if fuehrer.uebernommen != "" {
		t.Errorf("trotzdem übernommen: %q", fuehrer.uebernommen)
	}
}

// **Was übernommen wurde, steht in der Antwort — samt Übergangenem.**
func TestUebernehmenGibtDasErgebnis(t *testing.T) {
	fuehrer := &paketfuehrer{uebernahme: &service.Uebernahme{
		Uebernommen: []domain.Paket{{Art: "mse-game", Name: "magic", Pfad: "/p/magic.mse-game"}},
		Uebergangen: []string{"magic-m15.mse-style"},
		Befunde:     []domain.Befund{{Datei: "kaputt.mse-style", Meldung: "nicht lesbar"}},
	}}
	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	antwort := anwendung.VorlagenUebernehmen(t.Context(), "/mse2/data")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if fuehrer.uebernommen != "/mse2/data" {
		t.Errorf("übernommen aus %q", fuehrer.uebernommen)
	}

	if len(antwort.Uebernommen) != 1 || len(antwort.Uebergangen) != 1 || len(antwort.Befunde) != 1 {
		t.Errorf("Antwort: %+v", antwort)
	}
}

// **Ein Fehler wird übersetzt — und Listen bleiben Listen.**
func TestEinFehlerBeiDerUebernahmeWirdUebersetzt(t *testing.T) {
	anwendung := &app.App{}
	anwendung.MitPaketen(&paketfuehrer{fehler: errors.New("kein Verzeichnis")})

	antwort := anwendung.VorlagenUebernehmen(t.Context(), "/nirgends")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.Uebernommen == nil || antwort.Uebergangen == nil || antwort.Befunde == nil {
		t.Errorf("nil statt leerer Liste: %+v", antwort)
	}
}

// **Die Bindungen reichen durch** — und der Dialog kommt nur, wenn kein
// Vorschlag genannt ist.
func TestDieUebernahmebindungenReichenDurch(t *testing.T) {
	fuehrer := &paketfuehrer{}
	anwendung := app.NeueApp("")
	anwendung.MitPaketen(fuehrer)
	anwendung.MitMSE2Orten(func() []string { return []string{"/mse2/data"} })

	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.MSE2Vorschlaege(t.Context()); len(antwort.Orte) != 1 {
		t.Errorf("Vorschläge: %+v", antwort)
	}

	// **Mit Vorschlag: kein Dialog.** Wer einen der genannten Orte wählt, soll
	// nicht noch einmal gefragt werden.
	gefragt := false

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		gefragt = true

		return "", nil
	})

	if antwort := bindungen.VorlagenUebernehmen(t.Context(), "/mse2/data"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if gefragt {
		t.Error("trotz Vorschlag gefragt")
	}

	if fuehrer.uebernommen != "/mse2/data" {
		t.Errorf("übernommen aus %q", fuehrer.uebernommen)
	}

	// **Ohne Vorschlag: der Ordnerdialog.**
	var gewaehlt app.Dateiauswahl

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		gewaehlt = auswahl

		return "/anderswo/data", nil
	})

	if antwort := bindungen.VorlagenUebernehmen(t.Context(), ""); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !gewaehlt.Verzeichnis {
		t.Errorf("gefragt wurde nach %+v — erwartet war der Ordnerdialog", gewaehlt)
	}

	if fuehrer.uebernommen != "/anderswo/data" {
		t.Errorf("übernommen aus %q", fuehrer.uebernommen)
	}
}

// **Ein abgebrochener Dialog kommt als Fehler an und nicht als Absturz.**
func TestEinAbgebrochenerUebernahmedialog(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketen(&paketfuehrer{})
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", errors.New("der Dialog ging nicht auf")
	})

	antwort := app.NeueBindungen(anwendung).VorlagenUebernehmen(t.Context(), "")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler ging verloren")
	}

	if antwort.Uebernommen == nil {
		t.Error("nil statt leerer Liste")
	}
}
