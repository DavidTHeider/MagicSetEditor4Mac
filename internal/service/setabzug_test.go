package service_test

// **Ein Set als Archiv, ohne den Arbeitsort zu verschieben** — NET-04
// (#1708).
//
// Wer ein Set teilt, gibt eine **Kopie** weiter. `Schreibe` setzt den
// Arbeitsort auf das Ziel (#226) — ein geteiltes Set, das danach mit ⌘S in
// einen Ordner für Mailanhänge schriebe, wäre der Fehler, den dieser Lauf
// verhindert.

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinAbzugIstEinArchivUndLaesstDenOrt(t *testing.T) {
	t.Parallel()

	orte := &ortsschreiber{}
	archiv := filepath.Join(t.TempDir(), "Mein Probeset.mse-set")

	if fehler := service.NeuerSetschreiber(mitQuelle(), orte).Abzug(t.Context(), 7, archiv); fehler != nil {
		t.Fatalf("Abzug: %v", fehler)
	}

	if orte.ort != "" {
		t.Errorf("der Arbeitsort wurde auf %q gesetzt", orte.ort)
	}

	gelesen, fehler := zip.OpenReader(archiv)
	if fehler != nil {
		t.Fatalf("das Archiv: %v", fehler)
	}

	t.Cleanup(func() { _ = gelesen.Close() }) //nolint:errcheck // nur ein Lauf

	if len(gelesen.File) != 1 || gelesen.File[0].Name != "set" {
		t.Fatalf("im Archiv: %v", gelesen.File)
	}

	datei, fehler := gelesen.File[0].Open()
	if fehler != nil {
		t.Fatal(fehler)
	}

	inhalt, _ := io.ReadAll(datei) //nolint:errcheck // der Vergleich darunter zeigt, was fehlt

	if string(inhalt) != probenquelle {
		t.Errorf("die Hauptdatei weicht ab:\n%q", inhalt)
	}
}

// **Ein Set, das sich nicht lesen lässt, gibt keinen Abzug** — und keine
// halbe Datei.
func TestEinAbzugOhneQuelleHinterlaesstNichts(t *testing.T) {
	t.Parallel()

	archiv := filepath.Join(t.TempDir(), "kaputt.mse-set")
	kaputt := mitQuelle()
	kaputt.quellenfehler = errProbe

	if fehler := service.NeuerSetschreiber(kaputt, &ortsschreiber{}).Abzug(t.Context(), 7, archiv); fehler == nil {
		t.Fatal("ein Abzug ohne Quelle kam durch")
	}

	if _, fehler := zip.OpenReader(archiv); fehler == nil {
		t.Error("es liegt ein Archiv da")
	}
}

// **Jeder Fehlerweg hinterlässt kein Archiv** — und sagt, woran es lag.
func TestEinAbzugScheitertOhneRest(t *testing.T) {
	t.Parallel()

	ohneKarten := mitQuelle()
	ohneKarten.kartenfehler = errProbe

	mitKaputterHauptdatei := mitQuelle()
	mitKaputterHauptdatei.quelle.Datei = "fehlt/set"

	for name, fall := range map[string]struct {
		leser  quellenDouble
		archiv func(t *testing.T) string
	}{
		"die Karten lassen sich nicht lesen": {ohneKarten, func(t *testing.T) string {
			return filepath.Join(t.TempDir(), "a.mse-set")
		}},
		"der Bauort lässt sich nicht anlegen": {mitQuelle(), func(t *testing.T) string {
			return filepath.Join(t.TempDir(), "fehlt", "a.mse-set")
		}},
		"die Hauptdatei lässt sich nicht ablegen": {mitKaputterHauptdatei, func(t *testing.T) string {
			return filepath.Join(t.TempDir(), "a.mse-set")
		}},
		"am Ziel steht ein Verzeichnis": {mitQuelle(), func(t *testing.T) string {
			ziel := filepath.Join(t.TempDir(), "a.mse-set")
			if fehler := os.MkdirAll(filepath.Join(ziel, "drin"), 0o700); fehler != nil {
				t.Fatal(fehler)
			}

			return ziel
		}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			archiv := fall.archiv(t)

			if fehler := service.NeuerSetschreiber(fall.leser, &ortsschreiber{}).Abzug(t.Context(), 7, archiv); fehler == nil {
				t.Fatal("der Abzug kam durch")
			}

			if gelesen, fehler := zip.OpenReader(archiv); fehler == nil {
				_ = gelesen.Close() //nolint:errcheck // nur ein Lauf
				t.Error("es liegt ein Archiv da")
			}
		})
	}
}

// **Das Archiv heißt wie das Set** — und der Name kommt aus derselben Zeile
// des Bestands, die der Abzug liest.
func TestDerSetnameKommtAusDemBestand(t *testing.T) {
	t.Parallel()

	name, fehler := service.NeuerSetschreiber(mitQuelle(), &ortsschreiber{}).Setname(t.Context(), 7)
	if fehler != nil || name != "Mein Probeset" {
		t.Errorf("der Name: %q, %v", name, fehler)
	}

	kaputt := mitQuelle()
	kaputt.fehler = errProbe

	if _, fehler := service.NeuerSetschreiber(kaputt, &ortsschreiber{}).Setname(t.Context(), 7); fehler == nil {
		t.Error("ein unlesbares Set hatte einen Namen")
	}
}
