package app_test

// **Der Eingang wächst nicht ohne Ende** — #1944, Folge aus FMT-03.
//
// Eine Vorschau legt ihre Datei nur **zwischen** und nimmt sie danach wieder
// weg; eine Quelle, auf die kein Set mehr verweist, geht nach einer Stunde.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// alt macht eine Datei zwei Stunden alt.
func alt(t *testing.T, pfad string) {
	t.Helper()

	vorher := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(pfad, vorher, vorher); err != nil {
		t.Fatal(err)
	}
}

func TestEineVorschauLegtIhreDateiNurZwischen(t *testing.T) {
	ort := t.TempDir()
	ablage := app.NeueDienstablage(ort)

	// **Dieselbe Datei, schon eingefahren** — die Vorschau darf sie nicht
	// wegnehmen, auch wenn der Inhalt derselbe ist.
	eingefahren, err := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("PK ein Set"))
	if err != nil {
		t.Fatal(err)
	}

	pfad, weg, err := ablage.Zwischenlegen(t.Context(), ".mse-set", strings.NewReader("PK ein Set"))
	if err != nil || pfad == eingefahren || !strings.HasSuffix(pfad, ".mse-set") ||
		filepath.Dir(pfad) != filepath.Join(ort, "vorschau") {
		t.Fatalf("%q %v — eingefahren liegt %q", pfad, err, eingefahren)
	}

	if roh, _ := os.ReadFile(pfad); string(roh) != "PK ein Set" { //nolint:errcheck,gosec // aus t.TempDir()
		t.Errorf("Inhalt %q", roh)
	}

	weg()

	if _, err := os.Stat(pfad); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("die Vorschau liegt noch: %v", err)
	}

	if _, err := os.Stat(eingefahren); err != nil {
		t.Errorf("die eingefahrene Datei ist fort: %v", err)
	}

	// **Ein abgerissener Rumpf hinterlässt nichts** — und ein Ort, an dem
	// nichts entstehen kann, ist ein Fehler.
	if _, _, err := ablage.Zwischenlegen(t.Context(), ".mse-set", lesefehler{}); err == nil {
		t.Error("ein abgerissener Rumpf wurde zwischengelegt")
	}

	if eintraege, _ := os.ReadDir(filepath.Join(ort, "vorschau")); len(eintraege) != 0 { //nolint:errcheck // gezählt
		t.Errorf("in der Vorschau liegen %d Dateien", len(eintraege))
	}

	gesperrt := filepath.Join(ort, "vorschau")
	//#nosec G302 -- absichtlich zugesperrt, um den Fehlerweg zu betreten
	if err := os.Chmod(gesperrt, 0o500); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(gesperrt, 0o700) }) //nolint:errcheck,gosec // zurück für t.TempDir

	if _, _, err := ablage.Zwischenlegen(t.Context(), ".mse-set", strings.NewReader("x")); err == nil {
		t.Error("in eine gesperrte Vorschau gelegt")
	}

	datei := filepath.Join(t.TempDir(), "datei")
	if err := os.WriteFile(datei, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := app.NeueDienstablage(datei).Zwischenlegen(t.Context(), ".mse-set", strings.NewReader("x")); err == nil {
		t.Error("zwischengelegt ohne Ort")
	}
}

type lesefehler struct{}

func (lesefehler) Read([]byte) (int, error) { return 0, errors.New("die Leitung ist tot") }

func TestDerEingangRaeumtAbWoraufNiemandVerweist(t *testing.T) {
	ort := t.TempDir()
	verwiesen := map[string]bool{}
	gefragt := 0
	ablage := app.NeueDienstablage(ort).MitVerweis(func(_ context.Context, pfad string) (bool, error) {
		gefragt++

		if strings.Contains(pfad, "kaputt") {
			return false, errors.New("die Datenbank antwortet nicht")
		}

		return verwiesen[pfad], nil
	})

	eingang := filepath.Join(ort, "eingang")
	if err := os.MkdirAll(eingang, 0o700); err != nil {
		t.Fatal(err)
	}

	liegt := func(name string) string {
		pfad := filepath.Join(eingang, name)
		if err := os.WriteFile(pfad, []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}

		return pfad
	}

	verlassen, gebraucht, frisch, kaputt := liegt("verlassen.mse-set"), liegt("gebraucht.mse-set"),
		liegt("frisch.mse-set"), liegt("kaputt.mse-set")
	halb := liegt(".eingang-halb")

	for _, pfad := range []string{verlassen, gebraucht, kaputt, halb} {
		alt(t, pfad)
	}

	verwiesen[gebraucht] = true

	if _, err := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("neu")); err != nil {
		t.Fatal(err)
	}

	for pfad, bleibt := range map[string]bool{verlassen: false, gebraucht: true, frisch: true, kaputt: true, halb: true} {
		if _, err := os.Stat(pfad); (err == nil) != bleibt {
			t.Errorf("%s: bleibt %v, Stat %v", filepath.Base(pfad), bleibt, err)
		}
	}

	// **Höchstens einmal je Stunde** — eine zweite Übertragung gleich danach
	// fragt den Bestand nicht noch einmal.
	vorher := gefragt
	spaeter := liegt("spaeter.mse-set")
	alt(t, spaeter)

	if _, err := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("noch eins")); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(spaeter); err != nil || gefragt != vorher {
		t.Errorf("gleich danach aufgeräumt: %v, %d Fragen statt %d", err, gefragt, vorher)
	}
}
