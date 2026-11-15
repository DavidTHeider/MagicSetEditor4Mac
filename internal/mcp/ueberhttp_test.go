package mcp_test

// **Derselbe Server über HTTP** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Wonach hier gefragt wird
//
// Nach dem, was die Spezifikation `2025-11-25` für „Streamable HTTP"
// verlangt — **ein Endpunkt**, `POST` für Nachrichten, `202` für eine
// Benachrichtigung, `405` für den Ereignisstrom, den dieser Server nicht
// anbietet —, und nach dem, was 0115 dazu verlangt: **kein Ruf ohne Zugang**.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

const probemarke = "die-marke"

// ueberHTTP baut den Behandler über einer kleinen Werkzeugliste.
func ueberHTTP(t *testing.T) http.Handler {
	t.Helper()

	return mcp.UeberHTTP(mcp.Werkzeuge(mcp.Naehte{Netzweg: &netzwegprobe{}}), probemarke)
}

// frage stellt eine Anfrage an den Endpunkt.
func frage(
	t *testing.T, behandler http.Handler, methode, rumpf, marke string,
) *httptest.ResponseRecorder {
	t.Helper()

	var leser *strings.Reader
	if rumpf != "" {
		leser = strings.NewReader(rumpf)
	}

	var anfrage *http.Request
	if leser != nil {
		anfrage = httptest.NewRequestWithContext(t.Context(), methode, mcp.Endpunkt, leser)
		anfrage.Host = "127.0.0.1:7778" // der eigene Name (SEC-11)
	} else {
		anfrage = httptest.NewRequestWithContext(t.Context(), methode, mcp.Endpunkt, http.NoBody)
		anfrage.Host = "127.0.0.1:7778" // der eigene Name (SEC-11)
	}

	if marke != "" {
		anfrage.Header.Set("Authorization", "Bearer "+marke)
	}

	schreiben := httptest.NewRecorder()
	behandler.ServeHTTP(schreiben, anfrage)

	return schreiben
}

// **Eine Anfrage bekommt ihre Antwort — als JSON.**
func TestEineAnfrageUeberHTTPBekommtIhreAntwort(t *testing.T) {
	t.Parallel()

	antwort := frage(t, ueberHTTP(t), http.MethodPost,
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`, probemarke)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	if art := antwort.Header().Get("Content-Type"); art != "application/json" {
		t.Errorf("der Typ ist %q", art)
	}

	var gelesen struct {
		Ergebnis struct {
			Werkzeuge []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}

	if fehler := json.Unmarshal(antwort.Body.Bytes(), &gelesen); fehler != nil {
		t.Fatalf("kein JSON: %v — %s", fehler, antwort.Body.String())
	}

	if len(gelesen.Ergebnis.Werkzeuge) == 0 {
		t.Errorf("die Liste ist leer: %s", antwort.Body.String())
	}
}

// **Eine Benachrichtigung bekommt `202` ohne Rumpf.**
//
// **Das steht so in der Spezifikation**, und es ist mehr als eine Formsache:
// ein Klient, der auf eine Antwort wartet, die es nie geben wird, hängt.
func TestEineBenachrichtigungBekommtZweihundertzwei(t *testing.T) {
	t.Parallel()

	antwort := frage(t, ueberHTTP(t), http.MethodPost,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`, probemarke)

	if antwort.Code != http.StatusAccepted {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	if antwort.Body.Len() != 0 {
		t.Errorf("die Antwort trägt einen Rumpf: %s", antwort.Body.String())
	}
}

// **Ohne Marke kommt nichts durch — auch keine Liste.**
func TestOhneMarkeKommtNichtsDurch(t *testing.T) {
	t.Parallel()

	for name, marke := range map[string]string{
		"gar keine":    "",
		"eine falsche": "die-falsche",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := frage(t, ueberHTTP(t), http.MethodPost,
				`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`, marke)

			if antwort.Code != http.StatusUnauthorized {
				t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
			}

			if antwort.Header().Get("WWW-Authenticate") == "" {
				t.Error("die Absage sagt nicht, was verlangt wird")
			}
		})
	}
}

// **Ein Ereignisstrom wird abgelehnt — mit `405` und einem `Allow`.**
//
// **Die Spezifikation sieht genau das vor** für einen Server, der keinen
// anbietet. Ein `404` wäre eine Aussage über den Endpunkt, und den gibt es.
func TestEinEreignisstromWirdAbgelehnt(t *testing.T) {
	t.Parallel()

	antwort := frage(t, ueberHTTP(t), http.MethodGet, "", probemarke)

	if antwort.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Status %d", antwort.Code)
	}

	if erlaubt := antwort.Header().Get("Allow"); !strings.Contains(erlaubt, "POST") {
		t.Errorf("Allow ist %q", erlaubt)
	}
}

// **Eine Sitzung lässt sich beenden, auch wenn es keine gibt.**
//
// **Das ist kein Nichtstun**: der Klient wollte, dass danach nichts mehr
// steht, und danach steht nichts mehr.
func TestEineSitzungLaesstSichBeenden(t *testing.T) {
	t.Parallel()

	if antwort := frage(t, ueberHTTP(t), http.MethodDelete, "", probemarke); antwort.Code !=
		http.StatusNoContent {
		t.Fatalf("Status %d", antwort.Code)
	}
}

// **Eine Anfrage aus einem Browser wird abgewiesen — vor der Marke.**
//
// **Die Spezifikation verlangt es ausdrücklich**: „Servers MUST validate the
// Origin header on all incoming connections to prevent DNS rebinding
// attacks." Und der Angriff ist echt: eine Seite im Netz kann einen Namen auf
// die Schleife zeigen lassen.
func TestEineAnfrageAusEinemBrowserWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost, mcp.Endpunkt,
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	anfrage.Host = "127.0.0.1:7778" // der eigene Name (SEC-11)
	anfrage.Header.Set("Authorization", "Bearer "+probemarke)
	anfrage.Header.Set("Origin", "https://boese.example")

	schreiben := httptest.NewRecorder()
	ueberHTTP(t).ServeHTTP(schreiben, anfrage)

	if schreiben.Code != http.StatusForbidden {
		t.Fatalf("Status %d: %s", schreiben.Code, schreiben.Body.String())
	}
}

// **Ein anderer Weg als der Endpunkt gibt `404`.**
//
// **Ein Endpunkt und nicht mehrere** — das steht so in der Spezifikation, und
// dieser Lauf hält, dass es dabei bleibt.
func TestNurDerEineEndpunktAntwortet(t *testing.T) {
	t.Parallel()

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/werkzeuge", http.NoBody)
	anfrage.Header.Set("Authorization", "Bearer "+probemarke)

	schreiben := httptest.NewRecorder()
	ueberHTTP(t).ServeHTTP(schreiben, anfrage)

	if schreiben.Code != http.StatusNotFound {
		t.Errorf("Status %d", schreiben.Code)
	}
}

// **Eine kaputte Nachricht endet mit einer Antwort und nicht mit einem
// Absturz.**
func TestEineKaputteNachrichtEndetMitEinerAntwort(t *testing.T) {
	t.Parallel()

	antwort := frage(t, ueberHTTP(t), http.MethodPost, `kein JSON`, probemarke)

	// **`200` mit einem JSON-RPC-Fehler und nicht `400`**: die Spezifikation
	// kennt für eine unlesbare Nachricht den Fehlerrumpf, und `Server.behandle`
	// baut ihn. Ein HTTP-Status daneben wäre eine zweite Aussage über
	// dasselbe.
	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	if !strings.Contains(antwort.Body.String(), "error") {
		t.Errorf("die Antwort nennt keinen Fehler: %s", antwort.Body.String())
	}
}

// **Ein anderes Verfahren als POST, GET oder DELETE gibt `405`.**
func TestEinFremdesVerfahrenGibtFuenfhundertfuenf(t *testing.T) {
	t.Parallel()

	if antwort := frage(t, ueberHTTP(t), http.MethodPut, "", probemarke); antwort.Code !=
		http.StatusMethodNotAllowed {
		t.Errorf("Status %d", antwort.Code)
	}
}

// **Eine Nachricht über der Grenze wird abgewiesen, nicht gelesen.**
//
// **Acht Megabyte sind die Grenze**, und sie ist dieselbe wie über stdio. Ohne
// diesen Lauf stünde nur, dass die Zahl dasteht — nicht, dass sie greift.
func TestEineZuGrosseNachrichtKommtNichtDurch(t *testing.T) {
	t.Parallel()

	zuviel := strings.Repeat("x", mcp.Hoechstenachricht+1)

	antwort := frage(t, ueberHTTP(t), http.MethodPost,
		`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{"x":"`+
			zuviel+`"}}`, probemarke)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("eine zu große Nachricht gab %d", antwort.Code)
	}

	// **Und zwar ohne Rumpf** — wer nicht lesen konnte, hat nichts zu sagen.
	if antwort.Body.Len() != 0 {
		t.Errorf("es kam ein Rumpf zurück: %q", antwort.Body.String())
	}
}

// **Derselbe Behandler über einem fertigen Server.**
//
// **Den Weg geht `cmd/mse-mcp`**: dort hängen Fassung, Ausstattung,
// Kontingent und Wache am Server, und ein zweiter Server daneben trüge nichts
// davon. Ohne diesen Lauf wäre der Weg, den der Betrieb geht, ungemessen.
func TestDerBehandlerUeberEinemFertigenServer(t *testing.T) {
	t.Parallel()

	server := mcp.NeuerServer(mcp.Werkzeuge(mcp.Naehte{Netzweg: &netzwegprobe{}}))

	antwort := frage(t, mcp.UeberHTTPMit(server, probemarke), http.MethodPost,
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`, probemarke)

	if antwort.Code != http.StatusOK {
		t.Fatalf("tools/list gab %d — %s", antwort.Code, antwort.Body.String())
	}

	if !strings.Contains(antwort.Body.String(), "lesen.netzweg") {
		t.Errorf("die Werkzeugliste fehlt: %s", antwort.Body.String())
	}

	// **Und ohne Zugang kommt auch dieser Weg nicht durch** — die Marke hängt
	// am Behandler und nicht am Server.
	if ohne := frage(t, mcp.UeberHTTPMit(server, probemarke), http.MethodPost,
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`, ""); ohne.Code !=
		http.StatusUnauthorized {
		t.Errorf("ohne Marke kam %d zurück", ohne.Code)
	}
}
