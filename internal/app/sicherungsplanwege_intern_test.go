package app

// **Die Fehlerwege des Plans und des Stands** — OPS01
// (#2347):
// ein abgezogener Stick, ein toter Bestand, eine belegte Sperre.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// planAnwendung startet eine Anwendung, deren Dialog `ordner` wählt.
func planAnwendung(t *testing.T, ordner string) *App {
	t.Helper()

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) { return ordner, nil })

	bilder := t.TempDir()
	anwendung.MitBilderort(func() (string, error) { return bilder, nil })

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // nur zum Aufräumen

	return anwendung
}

// **Ein Stick, der eben abgezogen wurde, wird abgewiesen** — und nicht
// angelegt: sonst läge die Sicherung unter `/Volumes` auf der eigenen Platte.
func TestEinFehlenderOrdnerRichtetKeinenPlanEin(t *testing.T) {
	t.Parallel()

	abgezogen := filepath.Join(t.TempDir(), "Stick")
	anwendung := planAnwendung(t, abgezogen)

	if antwort := anwendung.SicherungsplanEinrichten(t.Context()); antwort.Fehler == nil {
		t.Fatalf("eingerichtet über einem Ordner, den es nicht gibt: %+v", antwort)
	}

	if _, err := os.Stat(abgezogen); !os.IsNotExist(err) {
		t.Errorf("der Ordner wurde angelegt: %v", err)
	}

	if danach := anwendung.Sicherungsstand(t.Context()); danach.Fehler != nil || danach.Stand.Plan.Ordner != "" {
		t.Errorf("danach: %+v", danach)
	}
}

// **Ein toter Bestand meldet sich** — beim Lesen, nach einer Ordnerwahl und
// beim Beenden des Plans.
func TestEinToterBestandMeldetSichAmPlan(t *testing.T) {
	t.Parallel()

	anwendung := planAnwendung(t, t.TempDir())

	if err := anwendung.db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if antwort := anwendung.Sicherungsstand(t.Context()); antwort.Fehler == nil {
		t.Errorf("gelesen: %+v", antwort)
	}

	if antwort := anwendung.SicherungsplanEinrichten(t.Context()); antwort.Fehler == nil {
		t.Errorf("eingerichtet: %+v", antwort)
	}

	if antwort := anwendung.SicherungsplanBeenden(t.Context()); antwort.Fehler == nil {
		t.Errorf("beendet: %+v", antwort)
	}
}

// **Zwei Pläne über demselben Bestand sichern nicht beide** — der zweite
// findet die Sperre belegt, und ohne Ort gibt es keine.
func TestDiePlansperre(t *testing.T) {
	t.Parallel()

	nehmen := plansperre(filepath.Join(t.TempDir(), "bestand.db"))

	freigeben, belegt, err := nehmen()
	if err != nil || belegt {
		t.Fatalf("der erste: %v, belegt %v", err, belegt)
	}

	if _, zweiter, zweitfehler := nehmen(); zweitfehler != nil || !zweiter {
		t.Errorf("der zweite: %v, belegt %v", zweitfehler, zweiter)
	}

	freigeben()

	if _, _, ohneOrt := plansperre(filepath.Join(t.TempDir(), "fehlt", "bestand.db"))(); ohneOrt == nil {
		t.Error("eine Sperre in einem Ordner, den es nicht gibt")
	}
}

// **Ein Anstoß ohne laufenden Plan tut nichts** — vor dem Start und nach
// dem Schluss.
func TestEinAnstossOhnePlanTutNichts(t *testing.T) {
	t.Parallel()

	var anwendung App

	anwendung.sicherungsplanAnstossen()

	if anwendung.planlauf.anstoss != nil {
		t.Error("ein Anstoß legte einen Plan an")
	}
}

// **Ein zweiter Anstoß wartet nicht** — liegt schon einer, genügt er; der
// Plan, der gerade sichert, hört ihn danach.
func TestEinZweiterAnstossWartetNicht(t *testing.T) {
	t.Parallel()

	var anwendung App

	anwendung.planlauf.anstoss = make(chan struct{}, 1)

	anwendung.sicherungsplanAnstossen()
	anwendung.sicherungsplanAnstossen()

	if liegen := len(anwendung.planlauf.anstoss); liegen != 1 {
		t.Errorf("%d Anstöße liegen — einer genügt", liegen)
	}
}
