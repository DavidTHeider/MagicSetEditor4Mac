// **Eine leere Antwort kommt als leere Liste, nicht als `null`** —
// #1164.
//
// Dieselbe Zusage wie am Fenster und über MCP: wer die Antwort liest, soll auf
// **eines** prüfen müssen. Über die Endpunktliste ist dieser Zweig nicht zu
// erreichen — das willige Double gibt drei Namen, und das soll es auch: ein
// Double, das nichts zurückgäbe, prüfte die vier Klassen an einer Antwort, die
// nichts sagt.
package rest_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// leereumgebung kennt keine Namen — und meldet keinen Fehler dabei.
//
// **Dieselbe Lage bei den Paketdateien** (#1164): ein Paket ohne bearbeitbare
// Datei gibt `nil`, und über die Endpunktliste ist der Zweig nicht zu
// erreichen — das willige Double gibt dort eine Datei zurück, und das soll es
// auch.
type leereumgebung struct{ williger }

func (leereumgebung) Namen(context.Context, int64) ([]string, error) {
	return nil, nil
}

func (leereumgebung) DateienUndBeiwerk(context.Context, string) (dateien, beiwerk []string, fehler error) {
	return nil, nil, nil
}

func TestEineLeereAntwortGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{
			Leser:         williger{},
			Schreiber:     williger{},
			Ausfuhr:       williger{},
			Abgleich:      williger{},
			Bilder:        williger{},
			Auskunft:      williger{},
			Sammeldaten:   williger{},
			Anlage:        williger{},
			Saetze:        williger{},
			Konsole:       leereumgebung{},
			Werkstatt:     leereumgebung{},
			Wortwahl:      williger{},
			Woerter:       williger{},
			Pack:          williger{},
			Angaben:       williger{},
			Stile:         williger{},
			Zusatz:        williger{},
			Pakete:        williger{},
			Zahlen:        williger{},
			Verlauf:       williger{},
			Einstellungen: williger{},
			Fassungen:     williger{},
			Anleger:       williger{},
		})

	for name, fall := range map[string]struct{ weg, feld string }{
		"Namen der Umgebung": {"/karten/2/skriptnamen", `"namen":[]`},
		"Dateien eines Pakets": {
			"/pakete/probe.mse-style/dateien", `"dateien":[]`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, dienst, endpunkt{
				methode: http.MethodGet, weg: fall.weg,
			}, gueltig)

			if antwort.Code != http.StatusOK {
				t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
			}

			if !strings.Contains(antwort.Body.String(), fall.feld) {
				t.Errorf("Antwort %s", antwort.Body.String())
			}
		})
	}
}

// Pruefen meldet keinen Befund.
func (leereumgebung) Pruefen(string) (zeile, spalte int, meldung string) {
	return 0, 0, ""
}
