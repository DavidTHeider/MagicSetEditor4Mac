package vorschauweg

// **Die Wache vor dem Servermodus** — QA-09
// (#1740),
// Befund P1 in `docs/PENTEST.md`.
//
// # Der Befund
//
// **`make dev-browser` stellt die Anwendung mit dem Bestand, der gerade gilt,
// unter `http://localhost:34115` bereit** — und Wails' HTTP-Weg nimmt jeden
// Ruf an. Gemessen am 2026-10-02 an `v3.0.0-beta.23`: `transport_http.go`
// prüft weder `Host` noch `Origin`, verlangt keinen Inhaltstyp und nimmt
// einen Ruf sogar als GET mit `?object=0&method=0&args=…` (ein Rückfall für
// WebKitGTK). Damit ruft eine beliebige Seite im Browser desselben Rechners
// **jede Bindung**: über DNS-Rebinding mit Antwort, über einen POST oder ein
// Bild ohne — die Wirkung tritt trotzdem ein.
//
// # Warum hier und nicht in Wails
//
// **Weil Wails die Weiche der Anwendung vor alles setzt**: `AssetOptions.
// Middleware` steht in `application.New` an erster Stelle der Kette und sieht
// damit auch `/wails/runtime`. Den Ereignisstrom über WebSocket
// (`/wails/events`) sieht sie nicht — er steht neben der Assetleitung; was er
// trägt und warum das bleibt, steht in `docs/PENTEST.md`.
//
// # Was sie prüft
//
//  1. **`Host` ist ein Name dieses Rechners** — `localhost`, `127.0.0.1`,
//     `::1`. Das ist die Abwehr gegen DNS-Rebinding: dort trägt die Anfrage
//     den Namen des Angreifers.
//  2. **Ein `Origin` ist genau die eigene Herkunft** — Schema `http` oder
//     `https` und derselbe Name samt Port wie `Host`. `null` aus einem
//     abgeschotteten Rahmen ist keine.
//  3. **`Sec-Fetch-Site` sagt „same-origin" oder „none"** — außer bei einem
//     Seitenwechsel (`Sec-Fetch-Mode: navigate`): ein Verweis von einer
//     fremden Seite darf die Anwendung öffnen, ein Bild auf einer fremden
//     Seite darf nichts laden.
//  4. **Der Rufweg nimmt nur POST** — für Browser ohne `Sec-Fetch-*` die
//     einzige Sperre gegen ein Bild auf `/wails/runtime?…`.
//
// # Was sie nicht prüft
//
// **Welcher lokale Prozess fragt** — dasselbe wie bei `rest.HerkunftTaugt`:
// ein Programm setzt weder `Origin` noch `Sec-Fetch-*` und kommt durch. Der
// Servermodus hat keine Marke; er ist ein Werkzeug der Entwicklung, und kein
// Freigabebau trägt ihn (`scripts/releaseschalter.sh`).

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
)

// Rufweg ist der Weg, über den Wails' HTTP-Transport Bindungen ruft.
const Rufweg = "/wails/runtime"

// eigeneNamen sind die Namen, unter denen der Servermodus sich ansprechen lässt.
var eigeneNamen = map[string]bool{"localhost": true, "127.0.0.1": true, "::1": true}

// NurVonDiesemRechner legt die Wache vor eine Weiche.
//
// **Sie nimmt die Weiche und gibt eine** — so lässt sie sich vor
// [Weg.Vor] setzen, ohne dass die Hülle `net/http` kennt (#54).
func NurVonDiesemRechner(innen func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(weiter http.Handler) http.Handler {
		drinnen := innen(weiter)

		return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if anfrage.URL.Path == Rufweg && anfrage.Method != http.MethodPost {
				schreiben.Header().Set("Allow", http.MethodPost)
				http.Error(schreiben, "der Rufweg nimmt nur POST", http.StatusMethodNotAllowed)

				return
			}

			if grund := fremd(anfrage); grund != "" {
				http.Error(schreiben, grund, http.StatusForbidden)

				return
			}

			drinnen.ServeHTTP(schreiben, anfrage)
		})
	}
}

// fremd nennt, warum eine Anfrage nicht von der eigenen Seite kommt — leer,
// wenn sie es tut.
func fremd(anfrage *http.Request) string {
	if name := rechnername(anfrage.Host); !eigeneNamen[name] {
		return fmt.Sprintf("der Servermodus antwortet nur unter localhost, nicht unter %q", name)
	}

	if herkunft := anfrage.Header.Get("Origin"); herkunft != "" {
		gelesen, fehler := url.Parse(herkunft)
		if fehler != nil || (gelesen.Scheme != "http" && gelesen.Scheme != "https") ||
			gelesen.Host != anfrage.Host {
			return fmt.Sprintf("die Herkunft %q ist nicht diese Seite", herkunft)
		}
	}

	seite := anfrage.Header.Get("Sec-Fetch-Site")
	if seite != "" && seite != "same-origin" && seite != "none" &&
		anfrage.Header.Get("Sec-Fetch-Mode") != "navigate" {
		return fmt.Sprintf("eine Anfrage von %q lädt hier nichts", seite)
	}

	return ""
}

// rechnername nimmt den Port ab — und die Klammern einer IPv6-Adresse.
func rechnername(rechner string) string {
	if name, _, fehler := net.SplitHostPort(rechner); fehler == nil {
		return name
	}

	return rechner
}
