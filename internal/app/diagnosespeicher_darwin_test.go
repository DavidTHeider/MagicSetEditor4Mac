//go:build darwin

package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Speicher je Prozess reist im Diagnosepaket, ohne Pfad** — OPS02
// (#2348), Abnahmekriterium 3: der Fußabdruck steht darin, und weder das
// Zuhause noch der Bestand.
//
// **Was er nicht hält**: die Liste der Webprozesse ist in einem Testlauf
// leer, denn er hält keine Webansicht — ihre Felder (Art und Bytes) belegt
// `TestDerSpeicherStehtJeProzessDa`, und am Fenster die Hüllenprobe.
func TestDerSpeicherImDiagnosepaketIstNurZahl(t *testing.T) {
	anwendung, bestandsort := gestarteteAnwendung(t)
	ctx := t.Context()

	if vorschau := anwendung.Diagnosepaket(ctx); vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "speicher"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(ctx, ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	metriken := archivinhalte(t, ziel)["metriken.json"]

	var gelesen struct {
		Prozess map[string]json.RawMessage `json:"prozess"`
	}

	if err := json.Unmarshal([]byte(metriken), &gelesen); err != nil {
		t.Fatalf("metriken.json: %v", err)
	}

	if string(gelesen.Prozess["fussabdruckBytes"]) == "0" || gelesen.Prozess["web"] == nil {
		t.Errorf("der Speicher fehlt: %s", metriken)
	}

	for _, verboten := range []string{os.Getenv("HOME"), bestandsort} {
		if strings.Contains(metriken, verboten) {
			t.Errorf("metriken.json trägt %q", verboten)
		}
	}
}
