//go:build !server

package fenster

// **Die Schutzrichtlinie der WebView** — SEC-17
// (#2164).
//
// # Warum es sie gibt
//
// Bis dahin setzte weder `frontend/index.html` noch die Hülle eine Content
// Security Policy: ein eingeschleustes Skript hätte alles gedurft, was die
// Seite darf — jede Bindung rufen, jede Datei lesen, die der Bildweg kennt.
// **Eine belegte Lücke, durch die es käme, gibt es nicht**; die Richtlinie ist
// die zweite Schicht, die greift, wenn die erste einmal nicht hält.
//
// # Warum als Kopfzeile und nicht im HTML
//
// **Weil nur die Kopfzeile beobachten kann** — `Content-Security-Policy-
// Report-Only` gibt es nicht als `<meta>` —, und weil `vite dev` sein
// Vorspiel inline einsetzt: eine Richtlinie in `index.html` legte den
// Entwicklungsweg lahm. Wails reicht die Kopfzeilen seiner Assetleitung als
// `NSHTTPURLResponse` an WebKit (`internal/assetserver/webview/
// responsewriter_darwin.go`, Zeilen 61–64, v3.0.0-beta.23); ob WebKit sie dort
// durchsetzt, belegt die Hüllenprobe `sec17`.
//
// # Wo die Weiche steht
//
// **In `internal/vorschauweg`** ([vorschauweg.Schutzbau]) — ein Handler
// braucht `net/http`, und dieses Paket soll es nicht kennen
// (`internal/schichten/netz_test.go`). Hier steht die Richtlinie und was sie
// erlaubt.
//
// # Was sie nicht abdeckt
//
// **Den Servermodus** (`schutznaht_server.go`): er liefert dieselbe Seite an
// einen Browser, und dort entscheidet die Wache, wer überhaupt fragen darf.

import (
	"io/fs"
	"log/slog"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// fremdeBildherkuenfte sind die Ursprünge der Vorschaubilder in der
// Kartensuche — **die Adresse nennt der Dienst**, und die WebView holt das
// Bild selbst (`Scryfallsuche.tsx`). Angestrichen vom Zweitblick über Phase
// H: die erste Fassung der Richtlinie nahm der Suche ihre Bilder, und kein
// gefahrenes Drehbuch betrat sie mit Bildern.
//
// **Gemessen am 2026-10-07** an den echten Diensten, je eine Suche: Scryfall
// nennt 24 von 24 Drucken auf `https://cards.scryfall.io`;
// magicthegathering.io nennt `http://gatherer.wizards.com` (aus `http` wird
// `https` in `mtgapi`), und das leitet mit 308 auf
// `https://gatherer-static.wizards.com` weiter — eine Weiterleitung prüft die
// Richtlinie an ihrem Ziel. Am nativen Fenster: Hüllenprobe `sec17bilder`.
// Von MTGJSON und Scrydex übernimmt die Suche kein Bild
// (`internal/app/dienstsuchadapter.go`).
var fremdeBildherkuenfte = []string{
	"https://cards.scryfall.io",
	"https://gatherer.wizards.com",
	"https://gatherer-static.wizards.com",
}

// schutzrichtlinie gibt die Richtlinie; `zusatz` ist eine weitere Quelle für
// `connect-src` — im Abnahmebau der Rückkanal der Hüllenprobe ([cspZusatz]),
// sonst leer.
//
// **Jede Quelle außer `'self'` ist gemessen**, und der Grund steht daneben:
// was hier steht, hat der beobachtende Lauf der Hüllenprobe `sec17`
// angestrichen (`docs/belege/huellenprobe/sec17.md`).
func schutzrichtlinie(zusatz string) string {
	verbinden := "'self'"
	if zusatz != "" {
		verbinden += " " + zusatz
	}

	return strings.Join([]string{
		"default-src 'self'",
		// **Skripte allein aus dem Bündel** — kein Inline, kein `eval`. Der
		// Tisch kommt ohne `eval` aus, seit er `pixi.js/unsafe-eval` lädt.
		"script-src 'self'",
		// **Stile auch inline, und das ist eine Entscheidung.** Mantine und
		// `react-style-singleton` (die Scroll-Sperre der Dialoge) setzen zur
		// Laufzeit `<style>`-Elemente — gemessen: `style-src-elem inline` an
		// zwei Stellen des Bündels. Ein eingeschleustes Stück CSS erreicht
		// nach draußen **nur die drei Bildursprünge** unten — als Bildanfrage
		// mit einem Pfad seiner Wahl; Schriften und Verbindungen bleiben auf
		// `'self'`. Der strengere Weg wäre eine Nonce je Dokument
		// (Mantines `getStyleNonce`, `__webpack_nonce__`).
		"style-src 'self' 'unsafe-inline'",
		// **`data:` für Bilder** — PixiJS prüft die Bildformate mit
		// eingebetteten AVIF- und WebP-Bildern (`loadTextures`, 8.20.1),
		// gemessen als `img-src data` im Lauf des Tisches. Dazu die
		// [fremdeBildherkuenfte].
		"img-src 'self' data: " + strings.Join(fremdeBildherkuenfte, " "),
		"font-src 'self'",
		"connect-src " + verbinden,
		// **`blob:` für Worker** — PixiJS lädt Texturen in einem Worker aus
		// einer `blob:`-Adresse (`WorkerManager`), gemessen als
		// `worker-src blob`.
		"worker-src 'self' blob:",
		"media-src 'self'",
		"object-src 'none'",
		"base-uri 'none'",
		"form-action 'none'",
		"report-uri " + vorschauweg.Berichtsweg,
	}, "; ") + ";"
}

// schutzbau legt die Schutzweiche vor die Assetleitung.
func schutzbau(innen application.Middleware, oberflaeche fs.FS, protokoll *slog.Logger) application.Middleware {
	return vorschauweg.Schutzbau(innen, oberflaeche, schutzrichtlinie(cspZusatz()), cspBeobachten(), protokoll)
}
