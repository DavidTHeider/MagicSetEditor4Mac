package rest_test

// **Der Dienst gegen einen Rufenden, der nichts Gutes will** — Werkstück O11
// von Phase 5.2 (#1435).
//
// # Warum
//
// Der Dienst lauscht auf der Schleifenadresse, und wer auf diesem Rechner
// einen Prozess starten kann, kann ihn anrufen. Er ist die einzige Fläche, an
// der **beides** zugleich fremd ist: die Marke und der Rumpf.
//
// **Gefahren wird der ganze Weg** — durch `Behandler()`, also durch Marke,
// Herkunftsprüfung, Drossel, Entzifferung und Antwortschreiber. Ein Lauf gegen
// die Entzifferung allein prüfte die Naht und nicht die Fläche; genau das
// verlangt O11 („negative Fälle laufen durch echte Serialisierung").
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **eine fremde Marke, die durchkommt** — jede Antwort unter 400 auf einen
//     Ruf ohne die richtige Marke ist ein Loch, kein Schönheitsfehler
//   - **ein 5xx auf einen kaputten Rumpf**: ein Aufrufer, der Unsinn schickt,
//     bekommt 400 oder 413; ein 500 wäre die Aussage, dass **dieser** Prozess
//     kaputt ist
//   - **eine Fehlerantwort ohne `code` oder ohne `meldung`** — ein Skript
//     liest das eine, ein Mensch das andere, und beide sollen immer da sein
//     (§8)
//   - **eine Antwort, die kein JSON ist**
//   - **ein 401 ohne `WWW-Authenticate`** — er sagte sonst nicht, womit man es
//     versuchen soll
//
//	go test ./internal/rest -run Fuzz -fuzz FuzzDienstruf -fuzztime 2m

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// dienstsaat sind Paare aus Marke und Rumpf.
var dienstsaat = []struct {
	marke string
	rumpf string
}{
	{probemarke, `{"printing":"abc","menge":1}`},
	{probemarke, `{"menge":`},
	{probemarke, ``},
	{probemarke, `null`},
	{probemarke, `[]`},
	{probemarke, `{"printing":"abc","menge":1,"unbekannt":true}`},
	{probemarke, `{"menge":1e309}`},
	{"", `{"printing":"abc","menge":1}`},
	{"fremd", `{"printing":"abc","menge":1}`},
	{probemarke + "x", `{"printing":"abc","menge":1}`},
	{strings.ToUpper(probemarke), `{"printing":"abc","menge":1}`},
	{probemarke[:len(probemarke)-1], `{"printing":"abc","menge":1}`},
	{"Bearer " + probemarke, `{"printing":"abc","menge":1}`},
	{probemarke, strings.Repeat(`{"a":`, 200)},
}

// **Keine fremde Marke kommt durch, und kein Rumpf macht ein 500.**
func FuzzDienstruf(f *testing.F) {
	for _, saat := range dienstsaat {
		f.Add(saat.marke, saat.rumpf)
	}

	f.Fuzz(func(t *testing.T, marke, rumpf string) {
		// **Ein eigener Dienst je Ausführung.** Die Drossel zählt je Dienst,
		// und ein geteilter gäbe nach ein paar hundert Rufen nur noch 429 —
		// der Lauf prüfte dann die Drossel und nichts sonst.
		dienst := neuerDienst(t)

		anfrage := httptest.NewRequestWithContext(t.Context(),
			http.MethodPost, "/sammlung", strings.NewReader(rumpf))
		anfrage.Host = "127.0.0.1:7777"
		anfrage.Header.Set("Authorization", "Bearer "+marke)

		antwort := httptest.NewRecorder()
		dienst.Behandler().ServeHTTP(antwort, anfrage)

		if antwort.Code >= http.StatusInternalServerError {
			t.Fatalf("Status %d auf Marke %q und Rumpf %q: %s",
				antwort.Code, marke, rumpf, antwort.Body.String())
		}

		if marke != probemarke && antwort.Code < http.StatusBadRequest {
			t.Fatalf("die Marke %q kam mit Status %d durch", marke, antwort.Code)
		}

		if antwort.Code == http.StatusUnauthorized &&
			antwort.Header().Get("WWW-Authenticate") == "" {
			t.Error("ein 401 ohne WWW-Authenticate")
		}

		var gelesen map[string]any
		if fehler := json.Unmarshal(antwort.Body.Bytes(), &gelesen); fehler != nil {
			t.Fatalf("die Antwort ist kein JSON (%d): %q", antwort.Code,
				antwort.Body.String())
		}

		if antwort.Code < http.StatusBadRequest {
			return
		}

		for _, feld := range []string{"code", "meldung"} {
			wert, da := gelesen[feld].(string)
			if !da || strings.TrimSpace(wert) == "" {
				t.Errorf("die Fehlerantwort (%d) trägt kein %q: %s",
					antwort.Code, feld, antwort.Body.String())
			}
		}
	})
}
