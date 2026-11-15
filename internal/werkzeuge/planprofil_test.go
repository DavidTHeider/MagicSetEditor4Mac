package werkzeuge_test

// **Wer in einem Test den Planbauer ruft, legt ein Profil an** — QA-03
// (#1656).
//
// Ohne Profil liest der Planbauer den Setkopf nicht und fällt **still** auf den
// Behelfsplan zurück — eine Feldliste auf beigem Grund, samt Befund, den der
// Aufrufer wegwerfen kann. Seit es Profile gibt (2026-09-07) maßen so zwei
// Kanonläufe, drei Profilläufe und drei Benches die Feldliste statt der Karte;
// zwei davon wurden rot, die übrigen lieferten Zahlen über etwas, das niemand
// meinte.
//
// # Was er hält
//
// Jede Testdatei unter `internal/app`, die `planbauer.Kartenplan(` oder
// `planbauer.Stilplan(` ruft, legt auch ein Profil an — über `imProfil`,
// `probekontext` oder `storage.MitProfil` —, oder steht mit Grund in der
// Ausnahmeliste.
//
// # Was er nicht hält
//
// **Ob jeder einzelne Ruf in der Datei das Profil bekommt.** Er sieht die
// Datei, nicht den Datenfluss; eine Datei mit einem Ruf im Profil und einem
// ohne ginge durch. Dafür gibt es `planZurKarte`, das einen Befund zum
// Fehlschlag macht.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// planbauerOhneProfil sind die Testdateien, die den Behelfsplan **meinen**.
var planbauerOhneProfil = map[string]string{
	"stilverdrahtung_intern_test.go": "prüft, dass der Behelfsplan begründet kommt — nicht den Plan selbst",
}

func TestWerDenPlanbauerRuftLegtEinProfilAn(t *testing.T) {
	t.Parallel()

	dateien, fehler := filepath.Glob(filepath.Join("..", "app", "*_test.go"))
	if fehler != nil || len(dateien) == 0 {
		t.Fatalf("keine Testdateien unter internal/app: %v", fehler)
	}

	rufer := 0

	for _, datei := range dateien {
		roh, lesefehler := os.ReadFile(datei) //nolint:gosec // Testdateien im Baum
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		quelle := string(roh)
		if !strings.Contains(quelle, "planbauer.Kartenplan(") && !strings.Contains(quelle, "planbauer.Stilplan(") {
			continue
		}

		rufer++

		name := filepath.Base(datei)
		if _, ausnahme := planbauerOhneProfil[name]; ausnahme {
			continue
		}

		if !strings.Contains(quelle, "imProfil(") && !strings.Contains(quelle, "probekontext(") &&
			!strings.Contains(quelle, "MitProfil(") {
			t.Errorf("%s ruft den Planbauer ohne Profil — er fiele still auf den Behelfsplan zurück", name)
		}
	}

	if rufer < 3 {
		t.Fatalf("nur %d Dateien rufen den Planbauer — der Leser greift daneben", rufer)
	}
}
