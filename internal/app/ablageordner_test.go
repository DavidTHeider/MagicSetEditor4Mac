package app_test

// **Jeder Dialog nimmt seinen Ordner, und ein fehlender hält an** — OPS-02
// (#1719).
//
// Vier Ordner: Hereinholen, Hinausgeben, Sicherungen, Diagnose. **Beim
// Setzen** fragt die Anwendung die Platte (AK-01), **beim Sichern** noch
// einmal — ein Laufwerk, das seitdem fehlt, gibt eine Meldung statt eines
// Dialogs an einem anderen Ort (AK-03). Geprüft an der Übergabe an die
// Dateiwahl; was der native Dialog damit tut, ist Apples Sache.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// setzeOrdner setzt einen Ordner und bricht ab, wenn es nicht geht.
func setzeOrdner(t *testing.T, anwendung *app.App, schluessel, ordner string) {
	t.Helper()

	if antwort := anwendung.EinstellungSetzen(t.Context(), schluessel, ordner); antwort.Fehler != nil {
		t.Fatalf("%s: %+v", schluessel, antwort.Fehler)
	}
}

// mitgeschrieben setzt eine Dateiwahl, die sich merkt, womit sie kam.
func mitgeschrieben(anwendung *app.App) *[]app.Dateiauswahl {
	gesehen := []app.Dateiauswahl{}

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		gesehen = append(gesehen, auswahl)

		return "", nil
	})

	return &gesehen
}

func TestEinFehlenderOrdnerWirdBeimSetzenAbgewiesen(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	fehlt := filepath.Join(t.TempDir(), "fehlt")

	antwort := anwendung.EinstellungSetzen(t.Context(), domain.VorgabeSicherungspfad, fehlt)
	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "nicht erreichbar") {
		t.Fatalf("%+v", antwort.Fehler)
	}
}

func TestJederSicherndialogNimmtSeinenOrdner(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	ausfuhr, sicherung, diagnose := t.TempDir(), t.TempDir(), t.TempDir()

	setzeOrdner(t, anwendung, domain.VorgabeAusfuhrpfad, ausfuhr)
	setzeOrdner(t, anwendung, domain.VorgabeSicherungspfad, sicherung)
	setzeOrdner(t, anwendung, domain.VorgabeDiagnosepfad, diagnose)

	gesehen := mitgeschrieben(anwendung)
	ctx := t.Context()

	anwendung.PersonendatenSichern(ctx, "1")
	anwendung.BestandSichern(ctx, "", false)
	anwendung.Diagnose(ctx)
	anwendung.DiagnoseSichern(ctx)

	orte := make([]string, 0, len(*gesehen))
	for _, auswahl := range *gesehen {
		orte = append(orte, auswahl.Startverzeichnis)
	}

	if strings.Join(orte, "|") != strings.Join([]string{ausfuhr, sicherung, diagnose}, "|") {
		t.Errorf("die Dialoge begannen bei %v", orte)
	}
}

// **Ein Laufwerk, das seit dem Setzen fehlt** (AK-03): Sichern bekommt eine
// Meldung und keinen Dialog an einem anderen Ort.
func TestEinVerschwundenerOrdnerHaeltDasSichernAn(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	ordner := filepath.Join(t.TempDir(), "platte")

	if fehler := os.Mkdir(ordner, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	setzeOrdner(t, anwendung, domain.VorgabeSicherungspfad, ordner)

	if fehler := os.Remove(ordner); fehler != nil {
		t.Fatal(fehler)
	}

	gesehen := mitgeschrieben(anwendung)

	antwort := anwendung.BestandSichern(t.Context(), "", false)
	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "nicht erreichbar") {
		t.Errorf("%+v", antwort.Fehler)
	}

	if len(*gesehen) != 0 {
		t.Errorf("der Dialog kam trotzdem: %+v", *gesehen)
	}
}

// **Beim Hereinholen geht der Dialog auf, wo das System ihn öffnet** — dort
// wird nichts gespeichert, und wer ein Laufwerk abgezogen hat, holt seine
// Datei vielleicht von woanders.
func TestEinVerschwundenerOrdnerLaesstDasOeffnenZu(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	ordner := filepath.Join(t.TempDir(), "platte")

	if fehler := os.Mkdir(ordner, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	setzeOrdner(t, anwendung, domain.VorgabeHochladepfad, ordner)

	if fehler := os.Remove(ordner); fehler != nil {
		t.Fatal(fehler)
	}

	gesehen := mitgeschrieben(anwendung)
	anwendung.NutzerbildWaehlen(t.Context(), "1", "1")

	if len(*gesehen) != 1 || (*gesehen)[0].Startverzeichnis != "" {
		t.Errorf("%+v", *gesehen)
	}
}

// **`~` wird zum Zuhause** — dem dessen, der die Anwendung ausführt.
func TestDieTildeWirdZumZuhause(t *testing.T) {
	heim := t.TempDir()
	t.Setenv("HOME", heim)

	if fehler := os.Mkdir(filepath.Join(heim, "Ausfuhr"), 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	anwendung := angelaufeneAnwendung(t)
	setzeOrdner(t, anwendung, domain.VorgabeAusfuhrpfad, "~/Ausfuhr")

	gesehen := mitgeschrieben(anwendung)
	anwendung.PersonendatenSichern(t.Context(), "1")

	if len(*gesehen) != 1 || (*gesehen)[0].Startverzeichnis != filepath.Join(heim, "Ausfuhr") {
		t.Errorf("%+v", *gesehen)
	}
}
