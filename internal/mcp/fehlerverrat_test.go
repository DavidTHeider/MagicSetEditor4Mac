package mcp_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein unbekannter Fehler verrät nichts** — QA-09
// (#1740),
// Befund P3 in `docs/PENTEST.md`.
//
// # Der Befund
//
// **Der Dienst verbirgt ihn, der Werkzeugserver reichte ihn durch**:
// `internal/rest/antwort.go` antwortet auf einen Fehler ohne Fachklasse
// „der Aufruf ist gescheitert", `internal/mcp` gab `err.Error()` zurück —
// samt SQLite-Meldung, Pfad unter dem Zuhause und der Adresse samt
// Transportfehler einer Gegenstelle. **Und was das Modell bekommt, geht bei
// einem Anbieter im Netz hinaus.** #1812 hatte den Fall der vollen Platte
// behoben; hier ist die Klasse dahinter (§10).
func TestEinUnbekannterFehlerVerraetDemModellNichts(t *testing.T) {
	t.Parallel()

	for name, fehler := range map[string]error{
		"SQLite samt Pfad": fmt.Errorf("die Sets lesen: %w",
			errors.New("SQL logic error: no such table: sets_alt (1) in /Users/jemand/Library/bestand.db")),
		"Transportfehler einer Gegenstelle": fmt.Errorf("das Paket holen: %w",
			errors.New(`Get "http://169.254.169.254/latest/meta-data/": dial tcp 169.254.169.254:80: connect: connection refused`)),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bestand := probenbestand()
			bestand.fehler = fehler

			text, istFehler := sprichEinWerkzeug(t, bestand)
			if !istFehler {
				t.Fatalf("kein isError: %s", text)
			}

			for _, verraeterisch := range []string{"/Users/jemand", "SQL", "169.254", "dial tcp"} {
				if strings.Contains(text, verraeterisch) {
					t.Errorf("die Antwort verrät %q: %s", verraeterisch, text)
				}
			}

			if !strings.Contains(text, "gescheitert") {
				t.Errorf("die Antwort sagt nicht, dass es scheiterte: %s", text)
			}
		})
	}
}

// **Und ein Fachfehler bleibt lesbar** — die Gegenprobe (§10): ohne sie stünde
// die Wache auch dann grün da, wenn sie jede Meldung verschluckte, und das
// Modell verlöre den Satz, aus dem es lernt.
func TestEinFachfehlerErreichtDasModellWeiter(t *testing.T) {
	t.Parallel()

	for name, fehler := range map[string]error{
		"Validierung":    fmt.Errorf("%w: der Name ist leer", domain.ErrValidation),
		"nicht gefunden": fmt.Errorf("%w: das Set 7 gibt es nicht", domain.ErrNotFound),
		"Konflikt":       fmt.Errorf("%w: das Set ist überholt", domain.ErrConflict),
		"Erlaubnis":      fmt.Errorf("%w: dafür braucht es die Rolle", domain.ErrErlaubnis),
		"nur lesend":     fmt.Errorf("%w: in diesem Profil wird nur gelesen", domain.ErrNurLesend),
		"abgeschaltet":   fmt.Errorf("%w: der Bereich ist aus", domain.ErrAbgeschaltet),
		"belegt":         fmt.Errorf("%w: der Bestand ist belegt", domain.ErrBelegt),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bestand := probenbestand()
			bestand.fehler = fehler

			text, istFehler := sprichEinWerkzeug(t, bestand)
			if !istFehler || !strings.Contains(text, strings.SplitN(fehler.Error(), ": ", 2)[1]) {
				t.Errorf("isError %v, Antwort: %s — erwartet war %q", istFehler, text, fehler)
			}
		})
	}
}

// sprichEinWerkzeug ruft `lesen.sets` und gibt Text und `isError`.
func sprichEinWerkzeug(t *testing.T, bestand *bestandsprobe) (string, bool) {
	t.Helper()

	zeilen := sprich(t, bestand,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lesen.sets","arguments":{}}}`)

	ergebnis := abbildung(t, entpacke(t, zeilen[0])["result"])

	return fmt.Sprint(ergebnis["content"]), ergebnis["isError"] == true
}
