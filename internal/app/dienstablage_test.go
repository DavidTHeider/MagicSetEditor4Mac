package app_test

// **Die Ablage des Dienstes** — FMT-03.AK-03 (#1692): übertragene Dateien
// nach ihrem Inhalt benannt und dauerhaft, Ausgaben zum Abholen mit Frist.

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieDienstablageBenenntNachDemInhalt(t *testing.T) {
	ort := t.TempDir()
	ablage := app.NeueDienstablage(ort)

	erster, err := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("PK ein Set"))
	if err != nil {
		t.Fatal(err)
	}

	zweiter, err := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("PK ein Set"))
	if err != nil || zweiter != erster {
		t.Fatalf("zweiter Ort %q, erster %q (%v)", zweiter, erster, err)
	}

	anderer, _ := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("PK ein anderes")) //nolint:errcheck // unten geprüft
	if anderer == erster || filepath.Dir(erster) != filepath.Join(ort, "eingang") || !strings.HasSuffix(erster, ".mse-set") {
		t.Errorf("Orte %q, %q", erster, anderer)
	}

	if roh, _ := os.ReadFile(erster); string(roh) != "PK ein Set" { //nolint:errcheck,gosec // aus t.TempDir()
		t.Errorf("Inhalt %q", roh)
	}

	// **Kein halber Rest bleibt liegen** — weder nach dem zweiten Mal noch
	// nach einem Abbruch mitten im Lesen.
	if _, err := ablage.Ablegen(t.Context(), ".mse-set", iotest.ErrReader(errors.New("die Leitung ist tot"))); err == nil {
		t.Error("ein abgerissener Rumpf wurde angenommen")
	}

	eintraege, _ := os.ReadDir(filepath.Join(ort, "eingang")) //nolint:errcheck // unten gezählt
	if len(eintraege) != 2 {
		t.Errorf("im Eingang liegen %d Dateien, erwartet 2", len(eintraege))
	}
}

func TestDieDienstablageGibtAusgabenZumAbholen(t *testing.T) {
	ort := t.TempDir()
	ablage := app.NeueDienstablage(ort)

	name, pfad, err := ablage.Ausgangsort(t.Context(), ".pdf")
	if err != nil || filepath.Dir(pfad) != filepath.Join(ort, "ausgang") || filepath.Base(pfad) != name ||
		!strings.HasSuffix(name, ".pdf") {
		t.Fatalf("%q %q %v", name, pfad, err)
	}

	if anderer, _, _ := ablage.Ausgangsort(t.Context(), ".pdf"); anderer == name { //nolint:errcheck // nur der Name zählt
		t.Error("zweimal derselbe Name")
	}

	if schreibfehler := os.WriteFile(pfad, []byte("%PDF-probe"), 0o600); schreibfehler != nil {
		t.Fatal(schreibfehler)
	}

	inhalt, err := ablage.Abholen(t.Context(), name)
	if err != nil {
		t.Fatal(err)
	}

	roh, _ := io.ReadAll(inhalt) //nolint:errcheck // unten geprüft
	_ = inhalt.Close()           //nolint:errcheck // nur gelesen

	if string(roh) != "%PDF-probe" {
		t.Errorf("abgeholt %q", roh)
	}

	if err := ablage.Vergessen(t.Context(), name); err != nil {
		t.Fatal(err)
	}

	for _, fall := range []string{name, "../eingang/x.mse-set", ".versteckt", ""} {
		if _, err := ablage.Abholen(t.Context(), fall); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("abholen %q: %v", fall, err)
		}

		if err := ablage.Vergessen(t.Context(), fall); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("vergessen %q: %v", fall, err)
		}
	}
}

// **Eine Ausgabe bleibt eine Stunde** — wer sie nicht abholt, bekommt sie
// danach nicht mehr, und die Platte füllt sich nicht mit Bögen.
func TestEineAlteAusgabeWirdWeggeraeumt(t *testing.T) {
	ort := t.TempDir()
	ablage := app.NeueDienstablage(ort)

	name, pfad, _ := ablage.Ausgangsort(t.Context(), ".pdf") //nolint:errcheck // unten geprüft
	if err := os.WriteFile(pfad, []byte("alt"), 0o600); err != nil {
		t.Fatal(err)
	}

	vorher := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(pfad, vorher, vorher); err != nil {
		t.Fatal(err)
	}

	if _, _, err := ablage.Ausgangsort(t.Context(), ".pdf"); err != nil {
		t.Fatal(err)
	}

	if _, err := ablage.Abholen(t.Context(), name); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("die alte Ausgabe liegt noch: %v", err)
	}
}

// **Ein Ort, an dem nichts entstehen kann, ist ein Fehler** — und kein
// stilles Weiter.
func TestEineDienstablageOhneOrtScheitert(t *testing.T) {
	datei := filepath.Join(t.TempDir(), "datei")
	if err := os.WriteFile(datei, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	ablage := app.NeueDienstablage(datei)

	if _, err := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("x")); err == nil {
		t.Error("Ablegen ohne Ort")
	}

	if _, _, err := ablage.Ausgangsort(t.Context(), ".pdf"); err == nil {
		t.Error("Ausgang ohne Ort")
	}

	if _, err := ablage.Abholen(t.Context(), "a.pdf"); err == nil {
		t.Error("Abholen ohne Ort")
	}
}

// **Und die übrigen Wege, auf denen die Platte nein sagt** — ein Eingang
// ohne Schreibrecht, ein Verzeichnis am Ort einer Prüfsumme, eine Ausgabe,
// die sich nicht wegnehmen lässt. Keiner davon ist „gibt es nicht".
func TestDieDienstablageSagtWennDiePlatteNeinSagt(t *testing.T) {
	ort := t.TempDir()
	ablage := app.NeueDienstablage(ort)

	belegt, _ := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("belegt")) //nolint:errcheck // unten
	if err := os.Remove(belegt); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(belegt, "drin"), 0o700); err != nil {
		t.Fatal(err)
	}

	if _, err := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("belegt")); err == nil {
		t.Error("über ein Verzeichnis gelegt")
	}

	_, pfad, _ := ablage.Ausgangsort(t.Context(), ".pdf") //nolint:errcheck // unten
	if err := os.MkdirAll(filepath.Join(pfad, "drin"), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := ablage.Vergessen(t.Context(), filepath.Base(pfad)); err == nil || errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein volles Verzeichnis vergessen: %v", err)
	}

	gesperrt := filepath.Join(ort, "eingang")
	//#nosec G302 -- absichtlich zugesperrt, um den Fehlerweg zu betreten
	if err := os.Chmod(gesperrt, 0o500); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(gesperrt, 0o700) }) //nolint:errcheck,gosec // zurück für t.TempDir

	if _, err := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("neu")); err == nil {
		t.Error("in einen gesperrten Eingang gelegt")
	}
}

// **Neben dem Bestand fragt die Ablage ihn** — eine alte Quelle, auf die kein
// Set verweist, geht beim nächsten Ablegen (#1944).
func TestDieAblageNebenDemBestandFragtIhn(t *testing.T) {
	db := frischeAppdatenbank(t)
	bestand := filepath.Join(t.TempDir(), "bestand.db")
	ablage := app.DienstablageNeben(db, bestand)

	verlassen := filepath.Join(filepath.Dir(bestand), "dienst", "eingang", "verlassen.mse-set")
	if err := os.MkdirAll(filepath.Dir(verlassen), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(verlassen, []byte("alt"), 0o600); err != nil {
		t.Fatal(err)
	}

	alt(t, verlassen)

	if _, err := ablage.Ablegen(t.Context(), ".mse-set", strings.NewReader("neu")); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(verlassen); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("die verlassene Quelle liegt noch: %v", err)
	}
}
