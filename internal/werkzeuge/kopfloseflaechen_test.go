package werkzeuge

// **REST und MCP bauen ihre Nutzerfläche ohne Sensor** — SEC-07
// (#2148).
//
// Vor diesen zwei Flächen sitzt niemand am Rechner; die Wurzel mit Sensor
// (`app.NutzerflaecheUeber`) öffnete dort auf einen leeren Nachweis hin einen
// Touch-ID-Dialog. Ein neuer Aufruf hier nähme sie wieder — also zählt der
// Wächter, statt es dem Gedächtnis zu überlassen.
//
// # Was er nicht hält
//
// **Die Wurzeln in `internal/app` selbst** — `kostenflaechen.go` und ihre
// Geschwister bauen über `NutzerflaecheUeber`, und keine ihrer Bindungen
// erreicht den Sensor. Wer dort eine Anmeldung, einen Profilwechsel oder ein
// Einrichten ergänzt, prüft das selbst.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKopfloseFlaechenNehmenKeinenSensor(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")

	for _, ort := range []string{"cmd/mse-rest", "cmd/mse-mcp", "internal/mcpnaehte", "internal/rest", "internal/mcp"} {
		dateien, fehler := filepath.Glob(filepath.Join(wurzel, ort, "*.go"))
		if fehler != nil || len(dateien) == 0 {
			t.Fatalf("%s: %v (%d Dateien)", ort, fehler, len(dateien))
		}

		for _, datei := range dateien {
			if strings.HasSuffix(datei, "_test.go") {
				continue
			}

			roh, lesefehler := os.ReadFile(datei) //#nosec G304 -- Werkzeug: Glob über das eigene Repositorium
			if lesefehler != nil {
				t.Fatal(lesefehler)
			}

			if strings.Contains(string(roh), "NutzerflaecheUeber(") {
				t.Errorf("%s baut die Nutzerfläche mit Sensor — `app.NutzerflaecheKopflos` nehmen (SEC-07)", datei)
			}
		}
	}
}
