package mcp_test

// **Fortschritt über HTTP kommt im Ereignisstrom der Antwort** —
// FMT-03.AK-02 und AK-03
// (#1692).
//
// Nachgelesen am 2026-09-29 über Context7, MCP-Spezifikation Fassung
// 2026-07-28, `basic/transports/streamable-http`: auf eine Anfrage antwortet
// der Server mit `application/json` **oder** mit `text/event-stream`; im
// Strom dürfen Benachrichtigungen zu dieser Anfrage vor der Antwort stehen,
// und die Antwort beendet ihn. Der Klient muss beides annehmen — dieser
// Server wählt den Strom nur, wenn er darum gebeten wurde.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

const meldenMitMarke = `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"schreiben.melden","arguments":{},"_meta":{"progressToken":7}}}`

func frageMitAnnahme(t *testing.T, rumpf, annahme string) *httptest.ResponseRecorder {
	t.Helper()

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost, mcp.Endpunkt, strings.NewReader(rumpf))
	anfrage.Host = "127.0.0.1:7778" // der eigene Name (SEC-11)
	anfrage.Header.Set("Authorization", "Bearer "+probemarke)
	anfrage.Header.Set("Accept", annahme)

	schreiben := httptest.NewRecorder()
	mcp.UeberHTTPMit(meldenderServer(3), probemarke).ServeHTTP(schreiben, anfrage)

	return schreiben
}

func TestFortschrittUeberHTTPKommtImEreignisstrom(t *testing.T) {
	antwort := frageMitAnnahme(t, meldenMitMarke, "application/json, text/event-stream")

	if art := antwort.Header().Get("Content-Type"); art != "text/event-stream" {
		t.Fatalf("Art %q, Rumpf %s", art, antwort.Body.String())
	}

	var daten []string

	for zeile := range strings.SplitSeq(antwort.Body.String(), "\n") {
		if wert, ist := strings.CutPrefix(zeile, "data: "); ist {
			daten = append(daten, wert)
		}
	}

	if len(daten) < 3 {
		t.Fatalf("Ereignisse: %v", daten)
	}

	erste := entpacke(t, daten[0])
	if erste["method"] != "notifications/progress" || abbildung(t, erste["params"])["progressToken"] != float64(7) {
		t.Errorf("erstes Ereignis: %v", erste)
	}

	if letzte := entpacke(t, daten[len(daten)-1]); letzte["id"] != float64(5) || letzte["result"] == nil {
		t.Errorf("letztes Ereignis: %v", letzte)
	}
}

// **Ohne Bitte um den Strom bleibt es bei JSON** — und damit ohne Meldungen:
// wer nur `application/json` annimmt, bekäme sie nirgends hin.
func TestOhneAnnahmeDesStromsBleibtEsBeiJSON(t *testing.T) {
	for _, annahme := range []string{"application/json", ""} {
		antwort := frageMitAnnahme(t, meldenMitMarke, annahme)

		if art := antwort.Header().Get("Content-Type"); art != "application/json" {
			t.Errorf("%q: Art %q", annahme, art)
		}

		if antwort := entpacke(t, antwort.Body.String()); antwort["id"] != float64(5) {
			t.Errorf("%q: Antwort %v", annahme, antwort)
		}
	}
}

// **Und ohne Marke auch mit Annahme** — ein Strom für eine Antwort ohne
// Meldungen wäre eine Verpackung ohne Inhalt.
func TestOhneFortschrittsmarkeKeinStrom(t *testing.T) {
	antwort := frageMitAnnahme(t,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"schreiben.melden","arguments":{}}}`,
		"application/json, text/event-stream")

	if art := antwort.Header().Get("Content-Type"); art != "application/json" {
		t.Errorf("Art %q", art)
	}
}

// **Eine kaputte Nachricht bekommt ihre Antwort als JSON** — auch mit Bitte
// um den Strom: ohne lesbare Anfrage gibt es keine Marke.
func TestEineKaputteNachrichtBleibtBeiJSON(t *testing.T) {
	antwort := frageMitAnnahme(t, `{kaputt`, "text/event-stream")

	if art := antwort.Header().Get("Content-Type"); art != "application/json" {
		t.Errorf("Art %q", art)
	}
}

// abgerissenerStrom nimmt die Kopfzeilen an und nichts danach.
type abgerissenerStrom struct{ httptest.ResponseRecorder }

func (*abgerissenerStrom) Write([]byte) (int, error) { return 0, http.ErrHandlerTimeout }

// **Ein abgerissener Strom hält den Server nicht auf** — der Ruf läuft zu
// Ende, und nichts hängt.
func TestEinAbgerissenerStromHaeltNichtsAuf(t *testing.T) {
	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost, mcp.Endpunkt, strings.NewReader(meldenMitMarke))
	anfrage.Host = "127.0.0.1:7778" // der eigene Name (SEC-11)
	anfrage.Header.Set("Authorization", "Bearer "+probemarke)
	anfrage.Header.Set("Accept", "text/event-stream")

	schreiben := &abgerissenerStrom{ResponseRecorder: *httptest.NewRecorder()}
	mcp.UeberHTTPMit(meldenderServer(3), probemarke).ServeHTTP(schreiben, anfrage)

	if schreiben.Code != http.StatusOK {
		t.Errorf("Status %d", schreiben.Code)
	}
}
