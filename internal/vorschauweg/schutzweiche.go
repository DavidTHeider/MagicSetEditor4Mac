package vorschauweg

// **Die Schutzweiche vor der Assetleitung** — SEC-17
// (#2164),
// Vermerk [0179](../../docs/entscheidungen/0179-die-webview-laeuft-unter-einer-gemessenen-schutzrichtlinie.md).
//
// **Hier und nicht in `internal/fenster`**, aus demselben Grund wie der
// Bildweg: ein `http.Handler` ist ohne `net/http` nicht zu schreiben, und wer
// `net/http` einführt, **kann** ins Netz (`internal/schichten/netz_test.go`).
// Dieses Blatt steht dort schon, und sein eigener Wächter
// (`keineverbindung_test.go`) hält es frei von Klienten und Lauschern. Die
// Richtlinie selbst — was sie erlaubt und warum — steht in der Hülle
// (`internal/fenster/schutznaht.go`); hier steht nur, wie sie an jede
// Antwort kommt und was mit einem Bericht geschieht.

import (
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// Berichtsweg nimmt die Verstoßberichte der Richtlinie an — derselbe Ursprung
// wie die Seite, damit kein Bericht das Bündel verlässt.
const Berichtsweg = "/csp-bericht"

// Berichtsgrenze ist, wie viel von einem Bericht gelesen wird. Ein Bericht
// ist ein paar hundert Bytes; was darüber liegt, ist kein Bericht.
const Berichtsgrenze = 16 << 10

// Schutzbau legt die Schutzweiche mit `richtlinie` vor die Assetleitung
// `innen` — `beobachten` meldet nur, statt durchzusetzen. `buendel` ist die
// ausgelieferte Oberfläche: nur eine Datei darin nennt ein Bericht beim Namen.
func Schutzbau(
	innen func(http.Handler) http.Handler, buendel fs.FS, richtlinie string, beobachten bool,
	protokoll *slog.Logger,
) func(http.Handler) http.Handler {
	return func(weiter http.Handler) http.Handler {
		return Schutzweiche(innen(weiter), buendel, richtlinie, beobachten, protokoll)
	}
}

// Schutzweiche setzt die Richtlinie an jede Antwort und nimmt die Berichte an.
//
// **An jede Antwort und nicht nur an das HTML**: WebKit liest sie am Dokument
// und an jedem Worker; ein Bild trägt sie ohne Wirkung mit.
func Schutzweiche(
	weiter http.Handler, buendel fs.FS, richtlinie string, beobachten bool, protokoll *slog.Logger,
) http.Handler {
	kopf := "Content-Security-Policy"
	if beobachten {
		kopf += "-Report-Only"
	}

	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		if anfrage.Method == http.MethodPost && anfrage.URL.Path == Berichtsweg {
			berichten(anfrage.Body, anfrage.Host, buendel, protokoll)
			schreiben.WriteHeader(http.StatusNoContent)

			return
		}

		schreiben.Header().Set(kopf, richtlinie)
		weiter.ServeHTTP(schreiben, anfrage)
	})
}

// verstoss ist, was WebKit über `report-uri` schickt.
type verstoss struct {
	Bericht struct {
		Anweisung  string `json:"effective-directive"`
		Verletzt   string `json:"violated-directive"`
		Abgewiesen string `json:"blocked-uri"`
		Quelle     string `json:"source-file"`
		Zeile      int    `json:"line-number"`
	} `json:"csp-report"`
}

// berichten schreibt einen Verstoß ins Protokoll — **gekappt** an der
// Berichtsgrenze, denn was hier ankommt, schreibt die Seite, und die ist
// genau das, wovor die Richtlinie schützt. `wirt` ist der Wirt der Seite:
// eine Quelle dort ist eine Datei des eigenen Bündels.
//
// **Und aus jedem Feld nur, was als Diagnose taugt** — angestrichen vom
// Zweitblick über Phase H: ein Bericht muss nicht von WebKit stammen, jedes
// Skript der Seite kann hierher schicken, was es will.
func berichten(rumpf io.Reader, wirt string, buendel fs.FS, protokoll *slog.Logger) {
	roh, _ := io.ReadAll(io.LimitReader(rumpf, Berichtsgrenze+1)) //nolint:errcheck // ein abgebrochener Bericht ist unlesbar, unten

	var gelesen verstoss
	if len(roh) > Berichtsgrenze || json.Unmarshal(roh, &gelesen) != nil {
		protokoll.Warn("CSP-Bericht unlesbar", "anzahl", len(roh))

		return
	}

	bericht := gelesen.Bericht
	anweisung := bericht.Anweisung
	if anweisung == "" {
		anweisung = bericht.Verletzt
	}

	protokoll.Warn("CSP-Verstoß", "anweisung", anweisungsname(anweisung), "abgewiesen", ursprung(bericht.Abgewiesen),
		"stelle", stelle(bericht.Quelle, wirt, buendel)+":"+strconv.Itoa(bericht.Zeile))
}

// anweisungsname gibt den Namen der Anweisung — `violated-directive` trägt
// die Quellen dahinter mit, und was keine Anweisung ist, kommt nicht ins
// Protokoll. **Eine Anweisung ist ein Name aus der Liste, nicht ein Wort, das
// wie einer aussieht** (Nachlauf über Phase H).
func anweisungsname(roh string) string {
	name, _, _ := strings.Cut(roh, " ")
	if !anweisungen[name] {
		return "«unbekannt»"
	}

	return name
}

// anweisungen sind die Anweisungen aus CSP Level 3, Abschnitt 6 (W3C, Stand
// vom 2026-09-16, nachgelesen am 2026-10-07).
var anweisungen = map[string]bool{
	"base-uri": true, "child-src": true, "connect-src": true, "default-src": true,
	"font-src": true, "form-action": true, "frame-ancestors": true, "frame-src": true,
	"img-src": true, "manifest-src": true, "media-src": true, "object-src": true,
	"report-to": true, "report-uri": true, "sandbox": true, "script-src": true,
	"script-src-attr": true, "script-src-elem": true, "style-src": true,
	"style-src-attr": true, "style-src-elem": true, "webrtc": true, "worker-src": true,
}

// stelle nennt die Datei, aus der der Verstoß kam — **beim eigenen Bündel
// ihren Namen ohne Abfrage und Anker, sonst nur den Ursprung**.
//
// **Eigen ist, was im Bündel liegt** — nicht, was auf dem Wirt der Seite
// steht (Nachlauf über Phase H): denselben Wirt kann jedes Skript nennen,
// und dahinter stünde dann ein Name seiner Wahl. Das Dokument selbst (`/`)
// ist `index.html`.
func stelle(quelle, wirt string, buendel fs.FS) string {
	if quelle == "" {
		return "«keine»"
	}

	adresse, fehler := url.Parse(quelle)
	if fehler != nil {
		return "«Adresse»"
	}

	if adresse.Host != wirt {
		return ursprung(quelle)
	}

	datei := strings.TrimPrefix(path.Clean("/"+adresse.Path), "/")
	if datei == "" {
		datei = "index.html"
	}

	if eintrag, statfehler := fs.Stat(buendel, datei); statfehler != nil || eintrag.IsDir() {
		return "«Datei»"
	}

	return path.Base(datei)
}

// ursprung kürzt eine abgewiesene Adresse auf Schema und Wirt — **ein Pfad
// kommt nicht ins Protokoll**: was eine Seite laden wollte, kann tragen, was
// ein Mensch getippt hat. Die Schlüsselwörter der CSP für das, was keine
// Adresse hat, bleiben, wie sie sind; ein anderer Wert nicht.
func ursprung(abgewiesen string) string {
	adresse, fehler := url.Parse(abgewiesen)
	if fehler != nil || adresse.Host == "" {
		if strings.Contains(abgewiesen, "/") {
			return "«Adresse»"
		}

		if !schluesselwoerter[abgewiesen] {
			return "«Wert»"
		}

		return abgewiesen
	}

	return adresse.Scheme + "://" + adresse.Host
}

// schluesselwoerter sind die Werte von `blocked-uri`, die keine Adresse
// sind: CSP Level 3, Abschnitt 5.3 (`inline`, `eval`, `wasm-eval`,
// `trusted-types-…`), dazu die Schemata, die ein Bericht allein nennt.
var schluesselwoerter = map[string]bool{
	"inline": true, "eval": true, "wasm-eval": true,
	"trusted-types-policy": true, "trusted-types-sink": true,
	"data": true, "blob": true, "about": true, "filesystem": true,
}
