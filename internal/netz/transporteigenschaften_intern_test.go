package netz

// **Was der Transport kann** — Werkstück A18 von Phase 5.3
// (#1533).
//
// # Warum das ein interner Lauf ist
//
// **Weil der Transport nicht exportiert ist und es nicht sein soll.** Er ist
// eine Eigenschaft von [Klient] und keine Fähigkeit für Aufrufer — dieselbe
// Überlegung wie bei `fristverbindung` in `klient_intern_test.go`. Von außen
// ließe sich nur prüfen, dass **eine** Anfrage durchgeht; hier steht,
// **wie** sie geht.
//
// # Und was hier gemessen wird, nicht behauptet
//
// **Ein echter Proxy und ein echter HTTP/2-Server.** Ein Lauf, der
// `Transport.Proxy != nil` prüfte, sagte nichts darüber, ob eine Anfrage
// wirklich dort ankommt — und `ForceAttemptHTTP2` zu lesen sagte nichts
// darüber, ob die Gegenstelle h2 auch bekommt.

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// derKlientTransport gibt den Transport, den [NeuerKlient] gebaut hat.
func derKlientTransport(t *testing.T, klient *Klient) *http.Transport {
	t.Helper()

	transport, ist := klient.unten.Transport.(*http.Transport)
	if !ist {
		t.Fatalf("der Klient trägt %T und keinen *http.Transport",
			klient.unten.Transport)
	}

	return transport
}

// **Eine Anfrage geht durch den Proxy, den die Umgebung nennt.**
//
// # Der Befund
//
// **`HTTPS_PROXY` wurde nicht gelesen.** `NeuerKlient` baut einen eigenen
// `http.Transport` mit eigenem `DialContext` — und ein handgebauter erbt
// `Proxy: http.ProxyFromEnvironment` **nicht**. Hinter einem Firmenproxy
// bekam jemand „connection refused" und keinen Hinweis, dass es einen Proxy
// gäbe.
//
// **Gemessen und nicht am Feld geprüft**: der Proxy hier ist ein echter
// Server, und er zählt mit.
func TestEineAnfrageGehtDurchDenProxyDerUmgebung(t *testing.T) {
	gesehen := make(chan string, 1)

	proxy := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			select {
			case gesehen <- anfrage.Host:
			default:
			}

			schreiben.WriteHeader(http.StatusOK)
		}))
	defer proxy.Close()

	// **`t.Setenv` und nicht `os.Setenv`** — und damit **kein** `t.Parallel`:
	// die Umgebung gehört dem ganzen Prozess.
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")

	klient := NeuerKlient(UeblicheFristen())

	antwort, fehler := hole(t, klient, "http://ziel.invalid/etwas")
	if fehler != nil {
		t.Fatalf("die Anfrage ging nicht durch: %v", fehler)
	}

	defer antwort.Body.Close() //nolint:errcheck // der Lauf endet ohnehin

	select {
	case host := <-gesehen:
		if host != "ziel.invalid" {
			t.Errorf("der Proxy sah %q", host)
		}
	default:
		t.Fatal("der Proxy hat die Anfrage nie gesehen — die Umgebung wird " +
			"nicht gelesen")
	}
}

// **Und `NO_PROXY` nimmt ein Ziel wieder heraus.**
//
// # Warum hier kein echter Server steht
//
// **Weil jeder Laufserver auf der Schleife liegt** — und die lässt
// `httpproxy` ohnehin aus, mit `NO_PROXY` wie ohne (gemessen am 2026-09-12).
// Der erste Anlauf hier fuhr einen `httptest.Server`, und er war grün, **ohne
// `NO_PROXY` je betreten zu haben**: er ging direkt, weil das Ziel auf
// `127.0.0.1` lag. Gefunden hat es die eingebaute Gegenprobe — ohne sie hätte
// der Lauf eine Zusage gehalten, die er nie geprüft hat (§10).
//
// **Gemessen wird deshalb die Entscheidung selbst**: welche Adresse der
// Transport für ein Ziel **außerhalb** der Schleife wählt. Dass aus dieser
// Entscheidung eine Anfrage wird, hält der Lauf darüber.
//
// # Und der Fall ist keiner aus dem Lehrbuch
//
// **Ollama auf einem anderen Rechner im eigenen Netz.** Dort greift die
// Schleifenregel nicht, und ohne `NO_PROXY` ginge die **lokale** Einbettung
// durch den Firmenproxy — ein Datenabfluss, den niemand angeordnet hat.
func TestNoProxyNimmtEinZielWiederHeraus(t *testing.T) {
	const firmenproxy = "http://firmenproxy.invalid:3128"

	t.Setenv("HTTP_PROXY", firmenproxy)
	t.Setenv("HTTPS_PROXY", firmenproxy)
	t.Setenv("NO_PROXY", "gpu.intern")

	waehlt := proxywahl(t, NeuerKlient(UeblicheFristen()))

	for name, fall := range map[string]struct{ ziel, erwartet string }{
		"das eigene Ollama im Netz": {
			"http://gpu.intern:11434/api/embed", "(direkt)",
		},
		"ein fremder Dienst": {
			"https://api.beispiel.invalid/x", firmenproxy,
		},
		// **Die Schleife bleibt außen vor, auch ohne Eintrag** — der Satz,
		// der die Sorge im Issue entkräftet hat.
		"das Ollama auf diesem Rechner": {
			"http://127.0.0.1:11434/api/embed", "(direkt)",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if gewaehlt := waehlt(fall.ziel); gewaehlt != fall.erwartet {
				t.Errorf("für %s wählt der Transport %q, erwartet %q",
					fall.ziel, gewaehlt, fall.erwartet)
			}
		})
	}
}

// proxywahl gibt, welche Adresse der Transport für ein Ziel wählt.
func proxywahl(t *testing.T, klient *Klient) func(ziel string) string {
	t.Helper()

	waehlen := derKlientTransport(t, klient).Proxy
	if waehlen == nil {
		t.Fatal("der Transport liest keinen Proxy — dann sagt dieser Lauf " +
			"nichts")
	}

	return func(ziel string) string {
		adresse, fehler := url.Parse(ziel)
		if fehler != nil {
			t.Fatalf("%q lesen: %v", ziel, fehler)
		}

		gewaehlt, wahlfehler := waehlen(&http.Request{URL: adresse})
		if wahlfehler != nil {
			t.Fatalf("die Proxywahl für %q: %v", ziel, wahlfehler)
		}

		if gewaehlt == nil {
			return "(direkt)"
		}

		return gewaehlt.String()
	}
}
