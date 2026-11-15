package mcp_test

// **Fortschritt und Abbruch über stdio** — FMT-03.AK-02
// (#1692).
//
// Nachgelesen am 2026-09-29 über Context7 in der MCP-Spezifikation, Fassung
// 2026-07-28, `basic/patterns/progress` und `basic/patterns/cancellation`:
// ein Aufrufer bittet mit `_meta.progressToken` um Meldungen, der Server
// schickt `notifications/progress` mit steigendem `progress`; über stdio
// bricht `notifications/cancelled` mit `requestId` einen Ruf ab, und der
// Server antwortet darauf **nicht**.

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type melderschluessel struct{}

// meldenderServer baut einen Server mit einem Werkzeug, das `schritte`
// meldet und danach fertig ist, und einem, das bis zum Abbruch wartet.
func meldenderServer(schritte int) *mcp.Server {
	melden := mcp.Werkzeug{
		Name: "schreiben.melden", Art: mcp.Schreibend,
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			melder, gehoert := ctx.Value(melderschluessel{}).(domain.Fortschrittsmelder)

			for erledigt := range schritte + 1 {
				if gehoert {
					melder(domain.Fortschritt{Erledigt: erledigt, Gesamt: schritte, Einheit: domain.EinheitKarte})
				}
			}

			return "fertig", nil
		},
	}

	warten := mcp.Werkzeug{
		Name: "schreiben.warten", Art: mcp.Schreibend,
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			<-ctx.Done()

			return "", ctx.Err()
		},
	}

	return mcp.NeuerServer([]mcp.Werkzeug{melden, warten}).MitFortschrittsnaht(
		func(ctx context.Context, melder domain.Fortschrittsmelder) context.Context {
			return context.WithValue(ctx, melderschluessel{}, melder)
		})
}

// gespraech schickt Zeilen und gibt die Antwortzeilen.
func gespraech(t *testing.T, server *mcp.Server, zeilen ...string) []string {
	t.Helper()

	var aus bytes.Buffer
	if err := server.Bediene(t.Context(), strings.NewReader(strings.Join(zeilen, "\n")+"\n"), &aus); err != nil {
		t.Fatalf("bedienen: %v", err)
	}

	return strings.FieldsFunc(aus.String(), func(r rune) bool { return r == '\n' })
}

func TestEinRufMitMarkeMeldetSeinenFortschritt(t *testing.T) {
	zeilen := gespraech(t, meldenderServer(3),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"schreiben.melden","arguments":{},"_meta":{"progressToken":"marke-1"}}}`)

	if len(zeilen) < 3 {
		t.Fatalf("Zeilen: %v", zeilen)
	}

	// **Die erste und die letzte Meldung kommen immer** — dazwischen darf
	// gedrosselt werden, und die Zahl steigt.
	erste, letzte := entpacke(t, zeilen[0]), entpacke(t, zeilen[len(zeilen)-2])
	for _, meldung := range []map[string]any{erste, letzte} {
		if meldung["method"] != "notifications/progress" || meldung["id"] != nil {
			t.Fatalf("keine Meldung: %v", meldung)
		}

		if parameter := abbildung(t, meldung["params"]); parameter["progressToken"] != "marke-1" ||
			parameter["total"] != float64(3) {
			t.Errorf("Parameter: %v", parameter)
		}
	}

	if parameter := abbildung(t, erste["params"]); parameter["progress"] != float64(0) ||
		parameter["message"] != "0 von 3 Karten" {
		t.Errorf("erste: %v", parameter)
	}

	if parameter := abbildung(t, letzte["params"]); parameter["progress"] != float64(3) {
		t.Errorf("letzte: %v", parameter)
	}

	vorher := -1.0

	for _, zeile := range zeilen[:len(zeilen)-1] {
		fortschritt, zahl := abbildung(t, entpacke(t, zeile)["params"])["progress"].(float64)
		if !zahl || fortschritt <= vorher {
			t.Errorf("nicht steigend: %v nach %v", fortschritt, vorher)
		}

		vorher = fortschritt
	}

	// **Die Antwort zuletzt** — nach ihr kommt keine Meldung mehr.
	if antwort := entpacke(t, zeilen[len(zeilen)-1]); antwort["id"] != float64(1) || antwort["result"] == nil {
		t.Errorf("Antwort: %v", antwort)
	}
}

func TestOhneMarkeKommtKeineMeldung(t *testing.T) {
	zeilen := gespraech(t, meldenderServer(3),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"schreiben.melden","arguments":{}}}`)

	if len(zeilen) != 1 || entpacke(t, zeilen[0])["id"] != float64(1) {
		t.Errorf("Zeilen: %v", zeilen)
	}
}

// **Ein abgebrochener Ruf bekommt keine Antwort** — und der nächste wird
// danach bedient, in der Reihenfolge, in der er kam.
func TestEinAbgebrochenerRufBekommtKeineAntwort(t *testing.T) {
	zeilen := gespraech(t, meldenderServer(1),
		`{"jsonrpc":"2.0","id":"lang","method":"tools/call","params":{"name":"schreiben.warten","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"schreiben.melden","arguments":{}}}`,
		`{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":"lang","reason":"genug"}}`)

	if len(zeilen) != 1 || entpacke(t, zeilen[0])["id"] != float64(2) {
		t.Errorf("Zeilen: %v", zeilen)
	}
}

// **Während ein Ruf läuft, bleibt der Server erreichbar** — ein `ping`
// bekommt seine Antwort vor der des laufenden Rufs.
func TestEinPingWaehrendEinesRufsWirdGleichBeantwortet(t *testing.T) {
	zeilen := gespraech(t, meldenderServer(1),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"schreiben.warten","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
		`{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":1}}`)

	if len(zeilen) != 1 || entpacke(t, zeilen[0])["id"] != float64(2) {
		t.Errorf("Zeilen: %v", zeilen)
	}
}

// **Ein Abbruch für einen unbekannten Ruf wird übergangen** — die
// Spezifikation erlaubt es ausdrücklich, und er ist auch sonst folgenlos.
func TestEinAbbruchFuerNichtsWirdUebergangen(t *testing.T) {
	zeilen := gespraech(t, meldenderServer(1),
		`{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":99}}`,
		`{"jsonrpc":"2.0","method":"notifications/cancelled","params":"kaputt"}`,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"schreiben.melden","arguments":{}}}`)

	if len(zeilen) != 1 || entpacke(t, zeilen[0])["id"] != float64(1) {
		t.Errorf("Zeilen: %v", zeilen)
	}
}

// **Endet der Strom, während ein Ruf läuft, wird er zu Ende bedient** —
// `printf … | mse-mcp` schickt alles auf einmal und schließt dann.
func TestEinEndenderStromWartetAufDenLaufendenRuf(t *testing.T) {
	zeilen := gespraech(t, meldenderServer(2),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"schreiben.melden","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"schreiben.melden","arguments":{}}}`)

	if len(zeilen) != 2 || entpacke(t, zeilen[0])["id"] != float64(1) ||
		entpacke(t, zeilen[1])["id"] != float64(2) {
		t.Errorf("Zeilen: %v", zeilen)
	}
}

// **Ein toter Strom wird auch während eines Rufs gemeldet** — beim `ping`,
// der nebenher beantwortet wird, und bei der Antwort des Rufs selbst.
func TestEinToterStromWirdAuchNebenherGemeldet(t *testing.T) {
	for _, fall := range []struct {
		name   string
		zeilen []string
	}{
		{"beim ping", []string{
			`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"schreiben.warten","arguments":{}}}`,
			`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
		}},
		{"bei der Antwort", []string{
			`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"schreiben.melden","arguments":{}}}`,
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			ein := strings.NewReader(strings.Join(fall.zeilen, "\n") + "\n")

			if err := meldenderServer(1).Bediene(t.Context(), ein, tauberSchreiber{}); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
