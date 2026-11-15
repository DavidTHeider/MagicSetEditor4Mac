package mcp

// **Derselbe Server über HTTP** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Was hier gebaut ist, und wonach
//
// **Der Transport „Streamable HTTP" der MCP-Spezifikation, Fassung
// `2025-11-25`** — nachgelesen am 2026-09-15 über Context7 (§1: „eine
// Auskunft aus der Erinnerung ist kein Prüfergebnis"), Abschnitt
// `basic/transports`. Was davon gilt:
//
//   - **Ein Endpunkt** für alles. Hier `/mcp`.
//   - **POST** trägt eine JSON-RPC-Nachricht. Eine **Anfrage** bekommt eine
//     Antwort, eine **Benachrichtigung** bekommt `202 Accepted` ohne Rumpf.
//   - **GET** eröffnet einen Ereignisstrom vom Server zum Klienten. Wer ihn
//     nicht anbietet, antwortet `405` — und genau das tut dieser Server.
//   - **`Mcp-Session-Id`** darf der Server bei der Initialisierung vergeben;
//     tut er es, trägt ihn jede weitere Anfrage.
//   - **`MCP-Protocol-Version`** trägt jede Anfrage nach der
//     Initialisierung; fehlt er, gilt `2025-03-26`. **Und was dieser Server
//     nicht spricht, bekommt `400`** — beides steht in demselben Absatz, und
//     die Vorgabe ist keine Erlaubnis, die Prüfung zu lassen.
//   - **Die `Origin`-Kopfzeile wird geprüft.** Die Spezifikation nennt es
//     ausdrücklich als Schutz gegen DNS-Rebinding.
//
// # Was dieser Server **nicht** tut, und warum
//
//   - **Kein eigener Ereignisstrom über GET.** Dieser Werkzeugserver schickt
//     von sich aus nichts: jedes Werkzeug ist ein Ruf mit einer Antwort. Ein
//     Ereignisstrom ohne Ereignisse wäre eine offene Verbindung ohne Inhalt —
//     `405` ist die Antwort, die die Spezifikation dafür vorsieht. **Die
//     Antwort auf einen `POST` wird seit FMT-03 zum Strom**, wenn der Klient
//     um Fortschritt bittet (`ereignisstrom.go`).
//   - **Keine Wiederaufnahme** (`Last-Event-ID`). Sie gehört zum
//     Ereignisstrom, den es hier nicht gibt.
//   - **Keine Sitzungskennung.** Sie ist ausdrücklich freigestellt („MAY"),
//     und dieser Server hält zwischen zwei Rufen nichts, was eine bräuchte.
//     **Ein Zugang tritt an ihre Stelle**, und der ist Pflicht.
//
// # Und der Zugang ist keine Zugabe
//
// **Jede Anfrage trägt die Marke**, auf der Schleife wie jenseits davon
// ([0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)).
// Die Spezifikation verlangt „robust authentication for all connections"; hier
// ist es dieselbe Marke, die `mse-rest` kennt — ein zweiter Zugangsbaukasten
// wäre eine zweite Stelle, an der jemand ihn falsch prüft.

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Endpunkt ist der eine Weg, über den dieser Transport spricht.
const Endpunkt = "/mcp"

// Vorgabefassung gilt, wenn eine Anfrage keine nennt.
//
// **Nachgelesen und nicht gewählt**: „If the server does not receive this
// header […] it should default to version 2025-03-26."
const Vorgabefassung = "2025-03-26"

// Fassungen sind die Protokollfassungen, die dieser Server annimmt.
//
// **Zwei, und beide sind begründet**: [fassung] ist die, die er in seiner
// Vorstellung nennt, und [Vorgabefassung] die, die die Spezifikation einer
// Anfrage **ohne** Kopfzeile zuschreibt. Wer die zweite ablehnte, wiese
// jeden Klienten ab, der die Kopfzeile gar nicht kennt — und genau für den
// steht die Vorgabe da.
//
// **Eine Kopie und kein Ausschnitt**: der Aufrufer bekommt eine eigene Liste,
// damit ein Leser die Fassungen dieses Servers nicht ändern kann.
func Fassungen() []string { return []string{Vorgabefassung, fassung} }

// Hoechstenachricht ist, wie groß eine hereinkommende Nachricht sein darf.
//
// **Dieselbe Zahl wie beim Lesen über stdio** — dort steht sie als
// Puffergrenze im `bufio.Scanner`, und zwei verschiedene Grenzen für denselben
// Inhalt wären zwei Antworten auf eine Frage.
const Hoechstenachricht = 8 * 1024 * 1024

// UeberHTTP gibt den Behandler, der diesen Server über HTTP bedient.
//
// **Die Marke kommt herein und wird nicht hier erzeugt** — sie gehört dem
// Posten, der den Prozess gestartet hat, und der sagt sie einem Menschen an.
func UeberHTTP(werkzeuge []Werkzeug, marke string) http.Handler {
	return behandlerFuer(NeuerServer(werkzeuge), marke)
}

// UeberHTTPMit ist [UeberHTTP] über einem fertigen Server.
//
// **Für den Composition Root**, der seinen Server ohnehin baut — Fassung,
// Ausstattung, Kontingent und Wache hängen daran, und ein zweiter Server
// daneben trüge nichts davon.
//
// **eigeneNamen** sind die Namen, unter denen der Server **zusätzlich** zur
// Schleife erreichbar ist — eine Bindung im Netz und die Namen ihres
// Zeugnisses (SEC-11, [herkunftTaugt]).
func UeberHTTPMit(server *Server, marke string, eigeneNamen ...string) http.Handler {
	return behandlerFuer(server, marke, eigeneNamen...)
}

// behandlerFuer baut den Behandler über einem fertigen Server.
func behandlerFuer(server *Server, marke string, eigeneNamen ...string) http.Handler {
	weg := http.NewServeMux()

	uhr := server.uhr
	if uhr == nil {
		uhr = time.Now
	}

	falsche := &fehlversuche{jetzt: uhr, seit: uhr()}

	weg.HandleFunc(Endpunkt, func(schreiben http.ResponseWriter, anfrage *http.Request) {
		if !herkunftTaugt(schreiben, anfrage, eigeneNamen) {
			return
		}

		if !markeTaugt(schreiben, anfrage, marke, falsche) {
			return
		}

		if !fassungTaugt(schreiben, anfrage) {
			return
		}

		// **Die Freigabe vor dem ersten Blick in den Bestand**, für jede
		// Methode und bis zur Antwort (Nachlauf über Phase 2 von 0.9.0,
		// Befund 3).
		anfrage, frei, gut := server.unterFreigabe(schreiben, anfrage)
		if !gut {
			return
		}

		defer frei()

		if !prozessTaugt(schreiben, anfrage, server) {
			return
		}

		switch anfrage.Method {
		case http.MethodPost:
			beantworte(schreiben, anfrage, server)

		case http.MethodGet:
			// **Kein Ereignisstrom** — der Grund steht im Kopf, und `405`
			// ist die Antwort, die die Spezifikation dafür vorsieht.
			schreiben.Header().Set("Allow", "POST, DELETE")
			schreiben.WriteHeader(http.StatusMethodNotAllowed)

		case http.MethodDelete:
			// **Eine Sitzung, die es nicht gibt, lässt sich beenden.** Das
			// ist kein Nichtstun, sondern die richtige Antwort: der Klient
			// wollte, dass danach nichts mehr steht, und danach steht nichts
			// mehr.
			schreiben.WriteHeader(http.StatusNoContent)

		default:
			schreiben.Header().Set("Allow", "POST, DELETE")
			schreiben.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return weg
}

// fassungTaugt weist ab, was dieser Server nicht spricht.
//
// # Warum das hier steht und nicht im Protokoll darunter
//
// **Weil die Kopfzeile zum Transport gehört und nicht zur Nachricht.** Über
// stdio gibt es sie nicht; dort sagt `initialize`, was gilt. Eine Prüfung im
// Server wäre eine Prüfung an der falschen Stelle — und über stdio eine, die
// nie zutrifft.
//
// **Und sie gilt für jede Methode**, nicht nur für `POST`: die Spezifikation
// knüpft sie an die Anfrage, nicht an die Nachricht darin.
func fassungTaugt(schreiben http.ResponseWriter, anfrage *http.Request) bool {
	// **Fehlt sie, gilt die Vorgabe** — und die steht in der Liste, also
	// kommt eine Anfrage ohne Kopfzeile durch. Das ist die zweite Hälfte
	// desselben Absatzes und keine Ausnahme.
	genannt := anfrage.Header.Get("MCP-Protocol-Version")
	if genannt == "" {
		genannt = Vorgabefassung
	}

	if slices.Contains(Fassungen(), genannt) {
		return true
	}

	schreiben.Header().Set("Content-Type", "application/json")
	schreiben.WriteHeader(http.StatusBadRequest)

	// **Die Absage nennt, was gilt.** Ein Klient, der nur „nein" hört, rät
	// beim nächsten Versuch — und rät falsch.
	gesagt, _ := json.Marshal(map[string]any{ //nolint:errcheck // Zeichenketten und Listen
		"error": map[string]any{
			"message":   "diese Protokollfassung spricht dieser Server nicht",
			"requested": genannt,
			"supported": Fassungen(),
		},
	})

	_, _ = schreiben.Write(gesagt) //nolint:errcheck // die Verbindung ist fort

	return false
}

// beantworte liest eine Nachricht und schreibt, was darauf folgt.
func beantworte(schreiben http.ResponseWriter, anfrage *http.Request, server *Server) {
	roh, fehler := io.ReadAll(http.MaxBytesReader(
		schreiben, anfrage.Body, Hoechstenachricht))
	if fehler != nil {
		schreiben.WriteHeader(http.StatusBadRequest)

		return
	}

	// **Mit Bitte um Meldungen als Ereignisstrom** (`ereignisstrom.go`).
	if stromGewuenscht(anfrage, roh) {
		strom := beginneStrom(schreiben)

		if antwort, antworten := server.behandle(mitSender(anfrage.Context(), strom.sende), roh); antworten {
			_ = strom.sende(antwort) //nolint:errcheck // die Verbindung ist fort
		}

		return
	}

	antwort, antworten := server.behandle(anfrage.Context(), roh)

	// **Eine Benachrichtigung bekommt `202` ohne Rumpf.** Das steht so in der
	// Spezifikation, und es ist mehr als eine Formsache: ein Klient, der auf
	// eine Antwort wartet, die es nie geben wird, hängt.
	if !antworten {
		schreiben.WriteHeader(http.StatusAccepted)

		return
	}

	schreiben.Header().Set("Content-Type", "application/json")

	// **Ohne Fehlerweg für das Serialisieren** — derselbe Satz wie bei
	// [Server.schreibe]: eine Antwort besteht aus Zeichenketten, Zahlen,
	// Listen und Abbildungen.
	geschrieben, _ := json.Marshal(antwort) //nolint:errcheck // siehe darüber

	_, _ = schreiben.Write(geschrieben) //nolint:errcheck // die Verbindung ist fort
}

// herkunftTaugt weist ab, was nach einem Browser aussieht oder einen fremden
// Namen trägt.
//
// # Warum das hier steht
//
// **Weil die Spezifikation es ausdrücklich verlangt**: „Servers MUST validate
// the Origin header on all incoming connections to prevent DNS rebinding
// attacks."
//
// # Was sie prüft — dieselbe Regel wie `rest.HerkunftTaugtUnter`
//
//   - **Kein `Origin`, kein `Referer`.** Beide setzt ein Browser, kein Skript.
//   - **`Host` ist einer der eigenen Namen** — `127.0.0.1`, `localhost` und,
//     bei einer Bindung im Netz, deren Adressen und die Namen ihres
//     Zeugnisses.
//
// **Hier stand: „Keine Herkunft ist in Ordnung — ein Browser schickt immer
// eine."** Das stimmt nicht (SEC-11,
// #2140):
// bei DNS-Rebinding hält der Browser die Seite für gleichherkünftig und
// schickt **kein** `Origin` — wohl aber den Namen des Angreifers als `Host`.
// Mit der eigenen Marke gab ein solches `DELETE` ein 204. Ein leeres `Origin`
// belegt keinen Aufrufer ohne Browser.
//
// **Sie ersetzt die Marke nicht**: ein lokales Programm setzt weder `Origin`
// noch `Referer` und schreibt den richtigen `Host`. Dagegen steht die Marke.
//
// **Das paketfremde `rest` wird nicht gerufen** — dieses Paket ist ein Blatt
// (§2); die Regel ist zehn Zeilen und steht hier ein zweites Mal, gehalten
// von derselben Tafel in `herkunft_test.go`.
func herkunftTaugt(schreiben http.ResponseWriter, anfrage *http.Request, eigeneNamen []string) bool {
	if anfrage.Header.Get("Origin") == "" && anfrage.Header.Get("Referer") == "" &&
		eigenerName(anfrage.Host, eigeneNamen) {
		return true
	}

	schreiben.WriteHeader(http.StatusForbidden)

	_, _ = fmt.Fprint(schreiben, //nolint:errcheck // die Verbindung ist fort
		`{"error":"dieser Server bedient nur Anfragen an seine eigenen Namen und keine aus einem Browser"}`)

	return false
}

// eigenerName sagt, ob `host` — mit oder ohne Port — einer der eigenen Namen
// ist. `::1` gehört nicht dazu: gebunden wird `127.0.0.1`.
func eigenerName(host string, weitere []string) bool {
	rechner, _, fehler := net.SplitHostPort(host)
	if fehler != nil {
		rechner = host
	}

	return rechner == "127.0.0.1" || rechner == "localhost" || slices.Contains(weitere, rechner)
}

// prozessTaugt fragt den eigenen Schalter dieses Prozesses, bevor irgendeine
// Anfrage beantwortet wird.
//
// **Zweitblick über Phase 6.12, Z7**: gefragt wurde er nur vor einem
// Werkzeug, und `ping`, `initialize` und `tools/list` gingen weiter durch —
// der Prozess lief, obwohl sein Schalter aus war. **Nach der Marke**: wer
// keine hat, erfährt nichts über den Schalter. Der Weg ist leer, und damit
// zählt nur der eigene Schalter, nicht der eines Werkzeugs.
//
// **Mit dem Kontext, den auch ein Werkzeug bekommt** ([Server.ausgestattet]):
// erst er trägt das Profil, und ohne Profil lesen sich die Einstellungen
// nicht — der Schalter galt dann als an, und `ping` ging durch. Gemessen am
// echten `mse-mcp` (`internal/flaechen/eigenerschalter_test.go`). Lässt sich
// der Kontext nicht ausstatten, geht die Anfrage weiter und sagt es selbst.
func prozessTaugt(schreiben http.ResponseWriter, anfrage *http.Request, server *Server) bool {
	gerichtet, ausstattungsfehler := server.ausgestattet(anfrage.Context())
	if ausstattungsfehler != nil {
		return true
	}

	fehler := server.abgeschaltet(gerichtet, Werkzeug{})
	if fehler == nil {
		return true
	}

	schreiben.Header().Set("Content-Type", "application/json")
	schreiben.WriteHeader(http.StatusForbidden)

	gesagt, _ := json.Marshal(map[string]any{ //nolint:errcheck // Zeichenketten
		"error": map[string]any{"code": "abgeschaltet", "message": fehler.Error()},
	})
	_, _ = schreiben.Write(gesagt) //nolint:errcheck // die Verbindung ist ohnehin am Ende

	return false
}

// markeTaugt prüft den Zugang.
//
// **Mit `subtle.ConstantTimeCompare`** und nicht mit `==` — dieselbe
// Überlegung wie in `internal/rest`: ein Vergleich, der früh abbricht, verrät
// über die Zeit, wie weit jemand richtig geraten hat.
func markeTaugt(schreiben http.ResponseWriter, anfrage *http.Request, marke string, falsche *fehlversuche) bool {
	gegeben, gefunden := strings.CutPrefix(
		anfrage.Header.Get("Authorization"), "Bearer ")

	if gefunden &&
		subtle.ConstantTimeCompare([]byte(gegeben), []byte(marke)) == 1 {
		return true
	}

	// **Über der Grenze ein 429 und kein 401** (SEC-12): wer rät, erfährt,
	// wann er wieder raten darf. **Verglichen wird trotzdem, und zwar
	// davor** — sonst sperrte ein Ratender die gültige Marke mit aus. Das
	// Budget begrenzt die Antworten, nicht die Vergleiche (Zweitblick über E,
	// Befund 5).
	if !falsche.zaehle() {
		schreiben.Header().Set("Retry-After", strconv.Itoa(int(fehlversuchsfenster.Seconds())))
		schreiben.WriteHeader(http.StatusTooManyRequests)

		return false
	}

	schreiben.Header().Set("WWW-Authenticate", `Bearer realm="mse-mcp"`)
	schreiben.WriteHeader(http.StatusUnauthorized)

	return false
}
