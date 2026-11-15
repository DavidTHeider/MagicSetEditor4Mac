// **Der Transport dieses Pakets** — Werkstück 4 von Phase 4.12
// (#1351).
//
// # Warum ein eigener Typ und kein `*http.Client`
//
// **Damit `internal/app` ihn benennen kann, ohne `net/http` einzuführen.**
// Dieselbe Lage wie bei `internal/vorschauweg` (#1270): die Netzgrenze
// (`internal/schichten/netz_test.go`) prüft **strukturell** — wer `net/http`
// importiert, **kann** ins Netz, und die ganze Bindungsfassade hätte damit den
// Weg nach draußen, für einen Typnamen.
//
// # Warum nicht `internal/netz`
//
// **Weil dessen Fläche nicht passt.** `netz.Hole` ist ein GET, `netz.SendeJSON`
// ein POST mit JSON-Rumpf; ein OAuth-Tausch ist ein POST mit
// `application/x-www-form-urlencoded` und der zweite Ruf ein GET mit
// `Authorization`. Die Netzgrenze sagt zu ihrer eigenen Liste: „Ein Eintrag,
// der die Liste kürzer machen soll, nicht länger."
//
// **Was das kostet, gehört gesagt**: dieser Transport hat keine Drossel und
// keinen Zwischenspeicher. Beides braucht er nicht — ein Anmeldefluss stellt
// **zwei** Anfragen, und zwar dann, wenn ein Mensch gerade darauf wartet.

package anmeldung

import (
	"net/http"
	"net/url"
	"time"
)

// Rufer ist der Transport dieses Pakets.
type Rufer struct{ unten *http.Client }

// Frist ist, wie lange ein einzelner Ruf höchstens dauert.
//
// **Eine Frist am Klienten und nicht nur im Kontext** (§16, Regel 3):
// `http.DefaultClient` hat keine, und ein hängender Anbieter hielte den
// Anmeldefluss beliebig lange fest.
//
// **Dreißig Sekunden**, weil ein Tausch zwei Rufe braucht und zwischen ihnen
// ein Mensch nichts tut — wer länger wartet, wartet auf etwas, das nicht mehr
// kommt.
const Frist = 30 * time.Second

// Proxywahl ist die Auswahlfunktion eines `http.Transport`.
//
// **Als Funktionstyp aus der Standardbibliothek und nicht als `netz`-Typ** —
// dieses Paket ist ein Blatt und kennt `internal/netz` nicht (§2). Was es
// braucht, ist die Gestalt, die `http.Transport.Proxy` erwartet; wer sie
// füllt, entscheidet der Composition Root.
type Proxywahl func(*http.Request) (*url.URL, error)

// NeuerRufer baut den Transport für den Betrieb.
//
// **Ohne Proxywahl** — dann gilt `http.DefaultTransport` und damit Gos
// zwischengespeicherte Umgebung. Für Tests und für Aufrufer ohne Bestand.
func NeuerRufer() *Rufer {
	return &Rufer{unten: &http.Client{Timeout: Frist}}
}

// NeuerRuferMit baut ihn mit derselben Proxywahl wie der Rest der Anwendung.
//
// # Warum es das gibt
//
// **Weil dieses Paket der zweite HTTP-Klient dieser Anwendung ist** — und bis
// N5 (#1571)
// mit **anderen** Eigenschaften: `http.DefaultTransport` liest die Umgebung
// **einmal** und hält sie danach fest, und eine Einstellung erreichte ihn gar
// nicht. Wer hinter einem Firmenproxy sitzt, konnte damit Karten laden und
// sich nicht anmelden.
//
// **Die Frist bleibt eine andere, und das ist Absicht**: ein Anmeldefluss
// sind zwei Rufe, zwischen denen ein Mensch wartet — drei getrennte Fristen
// wie in `internal/netz` beschreiben einen Bezug über eine langsame Leitung
// und nicht diesen Fall. Der Unterschied steht in
// `docs/NETZEIGENSCHAFTEN.md`.
func NeuerRuferMit(proxy Proxywahl) *Rufer {
	if proxy == nil {
		return NeuerRufer()
	}

	return &Rufer{unten: &http.Client{
		Timeout: Frist,
		Transport: &http.Transport{
			Proxy: proxy,

			// **HTTP/2 ausdrücklich** — aus demselben Grund wie in
			// `internal/netz`: ein Transport mit eigenen Feldern bekommt h2
			// sonst nicht, und zwar still.
			ForceAttemptHTTP2: true,
		},
	}}
}

// RuferUeber baut einen über einem fertigen Klienten.
//
// **Für Tests, und nur für sie** — dieselbe Begründung wie bei
// `netz.KlientUeber`: `httptest.Server` bringt einen eigenen Klienten mit, der
// sein Zertifikat kennt; ohne diesen Weg bräuchte jeder Test ein echtes.
func RuferUeber(unten *http.Client) *Rufer { return &Rufer{unten: unten} }
