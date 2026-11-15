// **Der Druckbogen ohne Fenster** —
// #1291.
//
// Eine eigene Datei, weil sie eine eigene Frage beantwortet:
// `zeichenwerk_test.go` prüft, dass aus einer Karte ein Bild wird — hier
// steht, dass aus einem Set Papier wird, und zwar über eine Naht, die keine
// Zeichenkette und kein Papierformat entgegennimmt.
package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// bogenwerk fährt das Probeset ein und gibt Werk und Setkennung.
//
// **Mit der Saat des Repositoriums als Wurzel**, aus demselben Grund wie beim
// Zeichnen: [app.Zeichennaht] nähme den Bestand der Nutzenden, und das
// Ergebnis hinge davon ab, was dort installiert ist (§11).
func bogenwerk(t *testing.T) (werk *app.Zeichenwerk, setID int64) {
	t.Helper()

	db := offeneProbedatenbank(t)

	saat, err := filepath.Abs(filepath.Join("..", "..", "saat"))
	if err != nil {
		t.Fatalf("Saatpfad: %v", err)
	}

	ergebnis, fehler := app.Einfuhrnaht(db).Set(probekontext(t),
		filepath.Join("..", "..", "testdaten", "probe.mse-set"))
	if fehler != nil {
		t.Fatalf("einfahren: %v", fehler)
	}

	return app.NeuesZeichenwerk(storage.NeuerSetLeser(db), []string{saat}, t.TempDir()),
		ergebnis.SetID
}

// **Aus einem Set wird eine PDF-Datei.**
func TestLegeBogenSchreibtEinPDF(t *testing.T) {
	werk, setID := bogenwerk(t)

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	gelegt, fehler := werk.LegeBogen(probekontext(t), setID, ziel)
	if fehler != nil {
		t.Fatalf("legen: %v", fehler)
	}

	if gelegt.Seiten < 1 || gelegt.Karten < 1 {
		t.Errorf("%d Seiten, %d Karten", gelegt.Seiten, gelegt.Karten)
	}

	if gelegt.Pfad != ziel {
		t.Errorf("der Pfad ist %q", gelegt.Pfad)
	}

	// **Die Befunde sind eine leere Liste und nicht `nil`** — dieselbe Falle
	// wie überall (§3): eine fehlende Liste wird zu `undefined`, und wer sie
	// liest, sieht nicht, dass nichts aufgefallen ist.
	if gelegt.Befunde == nil {
		t.Error("die Befunde sind nil statt einer leeren Liste")
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Bogendatei fehlt: %v", err)
	}

	if !strings.HasPrefix(string(roh), "%PDF-") {
		t.Errorf("die Datei beginnt nicht mit %%PDF-: %q", roh[:min(20, len(roh))])
	}
}

// **Ein Set, das es nicht gibt, ist ein Nicht-Gefunden — und kein leeres PDF.**
func TestLegeBogenFuerEinSetDasEsNichtGibt(t *testing.T) {
	werk, _ := bogenwerk(t)

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	_, fehler := werk.LegeBogen(probekontext(t), 4711, ziel)
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler: %v, erwartet ErrNotFound", fehler)
	}

	if _, err := os.Stat(ziel); err == nil {
		t.Error("es wurde trotzdem eine Datei angelegt")
	}
}
