// **Wie lange das Protokoll bleibt** — Werkstück 6 von Phase 4.13
// (#1352).
//
// # Der Anlass, gemessen
//
// **Am 2026-09-07 auf dem Rechner des Auftraggebers**: `mse.log` trug
// 1.347.580 Byte in 9.220 Zeilen, der älteste Eintrag vom 2026-08-22 — sechzehn
// Tage, rund 84 kB am Tag. `os.OpenFile` steht dort mit `O_APPEND`, und
// **nichts** begrenzte das.
//
// Das ist dieselbe Klasse wie die dreißig Sicherungen aus Werkstück 4: etwas
// wächst von selbst, und niemand sagt es. Die Frage aus Punkt 5 des Issues
// lautet „wohin, **wie lange**, wie zu finden" — das Wohin und das Finden
// standen; das Wielange nicht.
package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ist die Datei zu groß, wird sie beiseitegelegt — und nicht gelöscht.**
//
// # Warum eine Vorfassung und kein Abschneiden
//
// **Weil das Interessante am Anfang eines Fehlers steht.** Wer eine Datei bei
// ihrer Grenze abschneidet, verliert genau die Zeilen vor dem Absturz; wer sie
// beiseitelegt, behält sie eine Runde lang.
func TestEinVollesProtokollWirdBeiseitegelegt(t *testing.T) {
	ort := t.TempDir()
	pfad := filepath.Join(ort, app.Protokollname)

	alt := strings.Repeat("a", int(app.Protokollgrenze)+1)

	if fehler := os.WriteFile(pfad, []byte(alt), 0o600); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	schliessen := angelegtesProtokoll(t, pfad)

	if fehler := schliessen(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	// **Die neue ist frisch** — und nicht die alte mit einer Zeile mehr.
	frisch, fehler := os.Stat(pfad)
	if fehler != nil {
		t.Fatalf("die neue Datei fehlt: %v", fehler)
	}

	if frisch.Size() > app.Protokollgrenze {
		t.Errorf("die neue Datei wiegt %d Byte", frisch.Size())
	}

	// **Und die alte steht daneben, ganz.**
	vorher, lesefehler := os.ReadFile(filepath.Join(ort, app.Vorprotokollname)) //nolint:gosec // Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("die Vorfassung fehlt: %v", lesefehler)
	}

	if len(vorher) != len(alt) {
		t.Errorf("die Vorfassung wiegt %d statt %d Byte", len(vorher), len(alt))
	}
}

// **Und eine Datei unter der Grenze bleibt, wo sie ist.**
//
// **Die Gegenprobe zur Zusage** (§10): eine Drehung bei jedem Start wäre
// dasselbe wie keine Grenze — nach zwei Starts wäre alles fort.
func TestEinKleinesProtokollBleibtStehen(t *testing.T) {
	ort := t.TempDir()
	pfad := filepath.Join(ort, app.Protokollname)

	if fehler := os.WriteFile(pfad, []byte("eine Zeile\n"), 0o600); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	schliessen := angelegtesProtokoll(t, pfad)

	if fehler := schliessen(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if !strings.HasPrefix(string(roh), "eine Zeile\n") {
		t.Errorf("die bisherige Zeile ist fort:\n%s", roh)
	}

	if _, fehler := os.Stat(filepath.Join(ort, app.Vorprotokollname)); fehler == nil {
		t.Error("es entstand eine Vorfassung, obwohl nichts zu drehen war")
	}
}

// **Zweimal drehen lässt genau zwei Dateien übrig.**
//
// **Das ist die eigentliche Zusage**: der Platz ist begrenzt, nicht nur
// aufgeteilt. Eine Kette `mse.log.1`, `.2`, `.3` wüchse weiter — langsamer.
func TestZweimalDrehenLaesstZweiDateien(t *testing.T) {
	ort := t.TempDir()
	pfad := filepath.Join(ort, app.Protokollname)

	for _, inhalt := range []string{"erster Lauf", "zweiter Lauf"} {
		voll := inhalt + strings.Repeat("x", int(app.Protokollgrenze))

		if fehler := os.WriteFile(pfad, []byte(voll), 0o600); fehler != nil {
			t.Fatalf("vorbereiten: %v", fehler)
		}

		if fehler := angelegtesProtokoll(t, pfad)(); fehler != nil {
			t.Fatalf("schließen: %v", fehler)
		}
	}

	dabei, fehler := os.ReadDir(ort)
	if fehler != nil {
		t.Fatalf("nachsehen: %v", fehler)
	}

	// **Gezählt werden die Protokolle, nicht die Dateien im Ordner.**
	//
	// Daneben liegt seit Befund 5 des Zweitblicks die Sperrdatei, über die
	// sich zwei Prozesse verständigen, wer drehen darf. Sie ist kein
	// Protokoll — und ein Lauf, der stumpf „zwei Dateien" zählte, wäre an ihr
	// rot geworden, ohne dass sich an der Zusage etwas geändert hätte.
	protokolle := make([]string, 0, len(dabei))

	for _, eintrag := range dabei {
		if strings.HasPrefix(eintrag.Name(), app.Protokollname) &&
			eintrag.Name() != app.Protokollsperre {
			protokolle = append(protokolle, eintrag.Name())
		}
	}

	if len(protokolle) != 2 {
		t.Fatalf("es liegen %d Protokolle da: %v", len(protokolle), protokolle)
	}

	// **Und die Vorfassung ist die vom zweiten Lauf.** Die erste ist fort,
	// und das ist der Preis, den die Grenze kostet — er gehört genannt.
	vorher, lesefehler := os.ReadFile(filepath.Join(ort, app.Vorprotokollname)) //nolint:gosec // Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if !strings.HasPrefix(string(vorher), "zweiter Lauf") {
		t.Errorf("die Vorfassung ist die falsche: %.20s", vorher)
	}
}

// angelegtesProtokoll legt eines an und gibt seine Schließfunktion.
func angelegtesProtokoll(t *testing.T, pfad string) func() error {
	t.Helper()

	_, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll anlegen: %v", fehler)
	}

	return schliessen
}
