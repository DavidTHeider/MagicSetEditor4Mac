package fenster

// **Ein Rückruf der Hülle, der in Panik endet, beendet die Anwendung nicht**
// — F02 (#2341),
// Befund A-02.
//
// # Warum
//
// Wails fängt eine Panik in einem Ereignisrückruf und gibt sie seinem
// `PanicHandler`; ohne einen beendet es den Prozess (`os.Exit(1)`,
// `application.go`, v3.0.0-beta.27). Einen Haken auf `WindowClosing` ruft es
// ganz ohne Schutz, und eine Schlussaufgabe hängt mit einem Handler, der
// nicht beendet, für immer (`InvokeSync` ruft `wg.Done` nicht im `defer`).
// **Deshalb fängt jeder eigene Rückruf seine Panik selbst**, und Wails sieht
// sie nie.
//
// # Was die Prüfungen halten
//
// Die Panik endet im Riegel, sie steht mit einer Vorgangskennung im
// Protokoll, und der Riegel sagt, dass der Rückruf nicht zu Ende kam. Dazu
// die zwei Handler für das, was Wails selbst fängt.

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// protokolliert lenkt das Standardprotokoll in das echte Protokoll der
// Anwendung — **mit Schwärzer**, denn er setzt die Vorgangskennung aus dem
// Kontext, und ein roher Handler sähe sie nie.
func protokolliert(t *testing.T) fmt.Stringer {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "mse.log")

	schreiber, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll: %v", fehler)
	}

	vorher := slog.Default()
	slog.SetDefault(schreiber)
	t.Cleanup(func() {
		slog.SetDefault(vorher)
		_ = schliessen() //nolint:errcheck // nach dem Lesen
	})

	return datei(pfad)
}

// datei liest ihren Inhalt, wenn sie gefragt wird.
type datei string

func (d datei) String() string {
	roh, _ := os.ReadFile(string(d)) //nolint:errcheck,gosec // ein Pfad aus t.TempDir

	return string(roh)
}

func TestDerRiegelFaengtDiePanikEinesRueckrufs(t *testing.T) {
	gesammelt := protokolliert(t)

	if riegeln("das Öffnen aus dem Finder", func() { panic("Probepanik") }) {
		t.Error("ein Rückruf, der in Panik endete, galt als fertig")
	}

	zeile := gesammelt.String()
	for _, teil := range []string{"art=\"das Öffnen aus dem Finder\"", "panik=Probepanik", "kennung="} {
		if !strings.Contains(zeile, teil) {
			t.Errorf("im Protokoll fehlt %q:\n%s", teil, zeile)
		}
	}
}

func TestDerRiegelLaesstEinenGutenRueckrufDurch(t *testing.T) {
	gelaufen := false

	if !riegeln("der Start", func() { gelaufen = true }) || !gelaufen {
		t.Error("ein Rückruf ohne Panik lief nicht oder galt als gescheitert")
	}
}

func TestWailsPanikenUndFehlerStehenMitKennungImProtokoll(t *testing.T) {
	gesammelt := protokolliert(t)

	panikMelden(&application.PanicDetails{Error: errors.New("Probepanik in Wails"), StackTrace: "erste Zeile\nzweite Zeile"})
	fehlerMelden(errors.New("Probefehler in Wails"))

	zeile := gesammelt.String()
	for _, teil := range []string{"Probepanik in Wails", "Probefehler in Wails"} {
		if !strings.Contains(zeile, teil) {
			t.Errorf("im Protokoll fehlt %q:\n%s", teil, zeile)
		}
	}

	if strings.Count(zeile, "kennung=") != 2 {
		t.Errorf("nicht jede Zeile trägt eine Vorgangskennung:\n%s", zeile)
	}
}

func TestGesichertReichtDenWertDurchUndFaengtDiePanik(t *testing.T) {
	gesammelt := protokolliert(t)

	var gesehen string

	gesichert("das Ablegen", func(wert string) { gesehen = wert })("probe")

	if gesehen != "probe" {
		t.Errorf("der Rückruf bekam %q", gesehen)
	}

	gesichert("das Ablegen", func(string) { panic("Probepanik beim Ablegen") })("probe")

	if !strings.Contains(gesammelt.String(), "Probepanik beim Ablegen") {
		t.Errorf("die Panik steht nicht im Protokoll:\n%s", gesammelt.String())
	}
}
