package werkzeuge

// **`npx` installiert nichts** — Phase 5 von 0.9.0
// (#2326).
//
// Fehlt `frontend/node_modules`, holt `npx <werkzeug>` das Werkzeug aus der
// Registry, und zwar in der **neuesten** Fassung, nicht in der aus der
// Sperrdatei (§1). Am 2026-09-23, am 2026-10-15 und am 2026-10-16 lief so in
// einem frischen Arbeitsbaum ESLint 10 statt 9.39.5, und `npx tsc` startete
// das fremde Paket `tsc` statt TypeScript. `scripts/ci.sh` rief schon
// `npx --no-install`; `scripts/vorpruefung.sh` nicht.
//
// # Was er prüft
//
// Jeder Aufruf von `npx` in `scripts/*.sh` und im `Makefile` trägt
// `--no-install` — ein Kommentar und ein Hinweistext in Anführung nicht.
//
// # Was er nicht prüft
//
//   - **`npm exec`** und **`npm install`** — beides ruft kein Skript, und
//     ein neuer Aufruf wäre ein Befund für die Durchsicht.
//   - **Die `.mjs` unter `scripts/`** — sie rufen `npx` nicht.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var npxAufruf = regexp.MustCompile(`(^|[\s(&|;])npx\s`)

func TestNpxInstalliertNichts(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")

	skripte, fehler := filepath.Glob(filepath.Join(wurzel, "scripts", "*.sh"))
	if fehler != nil || len(skripte) < 20 {
		t.Fatalf("die Skripte: %v (%d)", fehler, len(skripte))
	}

	skripte = append(skripte, filepath.Join(wurzel, "Makefile"))
	gesehen := 0

	for _, pfad := range skripte {
		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Datei im Repositorium
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		for nummer, zeile := range strings.Split(string(roh), "\n") {
			code := strings.TrimSpace(zeile)
			if strings.HasPrefix(code, "#") || !npxAufruf.MatchString(code) {
				continue
			}

			// **Ein Hinweistext ist kein Aufruf**: `fehl "… 'npx …' hilft"`.
			if strings.HasPrefix(code, "fehl ") || strings.HasPrefix(code, "printf ") {
				continue
			}

			gesehen++

			if !strings.Contains(code, "npx --no-install ") {
				t.Errorf("%s:%d: %s — npx ohne --no-install holt fehlende Werkzeuge aus der Registry",
					filepath.Base(pfad), nummer+1, code)
			}
		}
	}

	if gesehen < 3 {
		t.Errorf("nur %d Aufrufe von npx gesehen — der Weg trifft nicht", gesehen)
	}
}
