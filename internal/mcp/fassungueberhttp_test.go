package mcp_test

// **Die Fassung in der Kopfzeile wird geprüft** — Befund 1 des Zweitblicks
// über Phase 5.4.
//
// # Der Anlass
//
// **Der Kopf von `ueberhttp.go` nannte `MCP-Protocol-Version` und eine
// Vorgabe** — und kein Ruf las die Kopfzeile. `Vorgabefassung` stand als
// exportierte Konstante ohne einen einzigen Leser da. Das ist die Gestalt aus
// §10: ein Kopf, der mehr verspricht als sein Rumpf hält, und niemand prüft
// nach, ob der Rumpf ihn einlöst.
//
// **Nachgelesen, nicht erinnert**: die Spezifikation verlangt für eine
// ungültige oder nicht unterstützte Fassung `400 Bad Request` und für eine
// **fehlende** Kopfzeile die Vorgabe `2025-03-26`. Zwei Regeln, und die
// zweite ist keine Erlaubnis, die erste zu lassen.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// mitSpion baut den Behandler über einem Werkzeug, das mitzählt.
//
// **Weil ein Status keine Ausführung ausschließt** — nachrangiger Befund 3
// des Nachlaufs zum Zweitblick über Phase 5.4. Ein Weg, der `400` schreibt
// **und** das Werkzeug ruft, bliebe grün, solange nur die Zahl geprüft wird.
func mitSpion(t *testing.T) (http.Handler, *netzwegprobe) {
	t.Helper()

	spion := &netzwegprobe{}

	return mcp.UeberHTTP(
		mcp.Werkzeuge(mcp.Naehte{Netzweg: spion}), probemarke), spion
}

// mitFassung stellt eine Anfrage mit einer genannten Fassung.
//
// **Leer heißt: die Kopfzeile fehlt** — und das ist ein eigener Fall, nicht
// derselbe wie eine leere Zeichenkette in der Kopfzeile.
func mitFassung(
	t *testing.T, genannt string,
) *httptest.ResponseRecorder {
	t.Helper()

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost,
		mcp.Endpunkt,
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	anfrage.Host = "127.0.0.1:7778" // der eigene Name (SEC-11)

	anfrage.Header.Set("Authorization", "Bearer "+probemarke)

	if genannt != "" {
		anfrage.Header.Set("MCP-Protocol-Version", genannt)
	}

	schreiben := httptest.NewRecorder()
	ueberHTTP(t).ServeHTTP(schreiben, anfrage)

	return schreiben
}

// **Was dieser Server spricht, kommt durch — und alles andere nicht.**
func TestDieGenannteFassungWirdGeprueft(t *testing.T) {
	t.Parallel()

	faelle := map[string]int{
		"":           http.StatusOK,
		"2025-03-26": http.StatusOK,
		"2025-06-18": http.StatusOK,
		"2024-11-05": http.StatusBadRequest,
		"1900-01-01": http.StatusBadRequest,
		"unbekannt":  http.StatusBadRequest,
	}

	for genannt, erwartet := range faelle {
		name := genannt
		if name == "" {
			name = "ohne Kopfzeile"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if antwort := mitFassung(t, genannt); antwort.Code != erwartet {
				t.Errorf("%q gab %d statt %d — %s",
					genannt, antwort.Code, erwartet, antwort.Body.String())
			}
		})
	}
}

// **Die Absage sagt, was gilt** — sonst rät der Klient.
func TestDieAbsageNenntDieFassungenDiesesServers(t *testing.T) {
	t.Parallel()

	antwort := mitFassung(t, "1900-01-01")

	rumpf := antwort.Body.String()

	for _, muss := range append([]string{"1900-01-01"}, mcp.Fassungen()...) {
		if !strings.Contains(rumpf, muss) {
			t.Errorf("die Absage nennt %q nicht: %s", muss, rumpf)
		}
	}
}

// **Und die Vorgabe ist die, die die Spezifikation nennt.**
//
// **Ohne diesen Lauf wäre die Konstante wieder ohne Leser** — sie stünde als
// Zahl da, und ob der Weg sie benutzt, bliebe eine Behauptung.
func TestDieVorgabefassungGiltOhneKopfzeile(t *testing.T) {
	t.Parallel()

	if mcp.Vorgabefassung != "2025-03-26" {
		t.Fatalf("die Vorgabe ist %q", mcp.Vorgabefassung)
	}

	// **Die Gegenprobe steckt im Vergleich**: wäre die Vorgabe eine, die
	// dieser Server nicht spricht, käme eine Anfrage ohne Kopfzeile nicht
	// durch — und genau das misst der Lauf darüber.
	if antwort := mitFassung(t, mcp.Vorgabefassung); antwort.Code != http.StatusOK {
		t.Errorf("die Vorgabe selbst kam nicht durch: %d", antwort.Code)
	}
}

// **Auch die zwei anderen Wege tragen sie** — es ist eine Kopfzeile je
// Anfrage und keine je Nachricht.
func TestDieFassungGiltAuchFuerDieAnderenWege(t *testing.T) {
	t.Parallel()

	for _, methode := range []string{http.MethodGet, http.MethodDelete} {
		t.Run(methode, func(t *testing.T) {
			t.Parallel()

			anfrage := httptest.NewRequestWithContext(
				t.Context(), methode, mcp.Endpunkt, http.NoBody)
			anfrage.Host = "127.0.0.1:7778" // der eigene Name (SEC-11)
			anfrage.Header.Set("Authorization", "Bearer "+probemarke)
			anfrage.Header.Set("MCP-Protocol-Version", "1900-01-01")

			schreiben := httptest.NewRecorder()
			ueberHTTP(t).ServeHTTP(schreiben, anfrage)

			if schreiben.Code != http.StatusBadRequest {
				t.Errorf("%s gab %d", methode, schreiben.Code)
			}
		})
	}
}

// **Eine abgewiesene Fassung ruft kein Werkzeug.**
//
// **Das ist die Zusage, nicht die Zahl.** Wer nur den Status misst, misst
// eine Absage — und nicht, dass davor nichts geschehen ist.
func TestEineAbgewieseneFassungFuehrtKeinWerkzeugAus(t *testing.T) {
	t.Parallel()

	behandler, spion := mitSpion(t)

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost,
		mcp.Endpunkt, strings.NewReader(
			`{"jsonrpc":"2.0","id":1,"method":"tools/call",`+
				`"params":{"name":"lesen.netzweg","arguments":{}}}`))
	anfrage.Host = "127.0.0.1:7778" // der eigene Name (SEC-11)

	anfrage.Header.Set("Authorization", "Bearer "+probemarke)
	anfrage.Header.Set("MCP-Protocol-Version", "1900-01-01")

	schreiben := httptest.NewRecorder()
	behandler.ServeHTTP(schreiben, anfrage)

	if schreiben.Code != http.StatusBadRequest {
		t.Fatalf("es kam %d", schreiben.Code)
	}

	if spion.gefragt {
		t.Error("das Werkzeug lief trotz Absage")
	}

	// **Und die Gegenprobe steht daneben**: mit einer Fassung, die gilt,
	// läuft dasselbe Werkzeug. Ohne sie sagte der Lauf nur, dass der Spion
	// schweigt — und das täte er auch, wenn ihn nie jemand riefe.
	behandler, spion = mitSpion(t)

	gilt := httptest.NewRequestWithContext(t.Context(), http.MethodPost,
		mcp.Endpunkt, strings.NewReader(
			`{"jsonrpc":"2.0","id":1,"method":"tools/call",`+
				`"params":{"name":"lesen.netzweg","arguments":{}}}`))
	gilt.Host = "127.0.0.1:7778" // der eigene Name (SEC-11)

	gilt.Header.Set("Authorization", "Bearer "+probemarke)

	behandler.ServeHTTP(httptest.NewRecorder(), gilt)

	if !spion.gefragt {
		t.Error("das Werkzeug lief auch mit gültiger Fassung nicht")
	}
}
