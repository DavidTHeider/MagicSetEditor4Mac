package netz

// **Wie dieses Programm ins Netz geht — als Vorgabe statt aus der Umgebung**
// — Werkstück N5 von Phase 5.4
// (#1571).
//
// # Die Regel, und sie ist eine Antwort auf eine offene Frage
//
// **Die Einstellung schlägt die Umgebung.** A18
// (#1533)
// hat sie wörtlich verlangt und offengelassen; hier steht sie, und ein Lauf
// hält sie fest.
//
// **Je Feld und nicht als Ganzes.** Wer nur `netz.proxy.https` setzt, behält
// `HTTP_PROXY` aus der Umgebung — sonst schaltete eine einzelne Einstellung
// die andere still ab, und niemand sähe es.
//
// # Und warum die Herkunft mitreist
//
// **Weil „es steht ein Proxy" ohne „woher" keine Auskunft ist.** Wer im
// Fenster `proxy.firma:3128` liest, muss wissen, ob er ihn ändern kann oder
// ob er aus der Umgebung dieses Prozesses kommt — dort ändert ihn kein
// Fenster.

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/http/httpproxy"
)

// Herkunft sagt, woher ein wirksamer Wert stammt.
type Herkunft string

const (
	// AusDerEinstellung heißt: jemand hat ihn im Bestand gesetzt.
	AusDerEinstellung Herkunft = "einstellung"

	// AusDerUmgebung heißt: er kommt aus `HTTP_PROXY` und Geschwistern.
	AusDerUmgebung Herkunft = "umgebung"

	// OhneWert heißt: keiner von beiden nennt einen.
	OhneWert Herkunft = "ohne"
)

// Proxyquelle gibt die Vorgabe — **je Anfrage gefragt**.
//
// # Warum eine Funktion und kein Wert
//
// **Weil ein Klient einmal gebaut wird und eine Einstellung sich ändert.**
// Das Fenster baut seine Nähte beim Start; wer danach einen Proxy einträgt,
// erwartet nicht, die Anwendung neu starten zu müssen. `http.Transport` ruft
// `Proxy` **je Anfrage** — damit ist das der Ort, an dem eine Einstellung
// ohne Neustart wirkt.
//
// **`nil` heißt: nur die Umgebung.** Für `internal/bezug`, das den Bestand
// nicht kennt, und für Tests.
//
// **Was dabei nicht mitwandert, sind die Fristen**: sie sitzen im Wähler und
// im Transport und werden beim Bauen festgelegt. Eine geänderte Frist wirkt
// deshalb erst am nächsten Klienten — das steht so in
// `docs/NETZEIGENSCHAFTEN.md`.
//
// **Sie bekommt den Kontext der Anfrage** (D06, #2237): die Einstellungen
// stehen je Profil, und das Profil reist im Kontext. Ohne ihn fragte eine
// Quelle über den Speicher ins Leere, und ein Lesefehler heißt hier „wie
// ohne Einstellung" — die Anfrage ging am eingestellten Proxy vorbei.
type Proxyquelle func(ctx context.Context) Proxyvorgabe

// Proxyvorgabe ist, was eingestellt wurde — leer heißt „die Umgebung gilt".
type Proxyvorgabe struct {
	HTTP  string
	HTTPS string
	Aus   string
}

// Proxystand ist der **wirksame** Stand samt Herkunft je Feld.
type Proxystand struct {
	HTTP  Proxyfeld
	HTTPS Proxyfeld
	Aus   Proxyfeld
}

// Proxyfeld ist ein wirksamer Wert und woher er kommt.
type Proxyfeld struct {
	Wert     string
	Herkunft Herkunft
}

// Benutzt sagt, ob überhaupt ein Proxy im Spiel ist.
//
// **Für die Meldung eines Fehlschlags** (§8): „connection refused" ohne den
// Zusatz, dass ein Proxy dazwischenstand, ist die stille Form — und #1533
// führt sie als offenen Punkt.
func (p Proxystand) Benutzt() bool {
	return p.HTTP.Wert != "" || p.HTTPS.Wert != ""
}

// StandMit rechnet aus Vorgabe und Umgebung den wirksamen Stand.
//
// **Die Umgebung wird bei jedem Ruf gelesen** und nicht einmal: `httpproxy`
// liest sie beim Ruf von `FromEnvironment`, und genau das ist der Grund,
// warum dieses Paket nicht `http.ProxyFromEnvironment` nimmt — jenes
// speichert sie zwischen (gemessen am 2026-09-12).
func StandMit(ctx context.Context, quelle Proxyquelle) Proxystand {
	var vorgabe Proxyvorgabe
	if quelle != nil {
		vorgabe = quelle(ctx)
	}

	umgebung := httpproxy.FromEnvironment()

	return Proxystand{
		HTTP:  feld(vorgabe.HTTP, umgebung.HTTPProxy),
		HTTPS: feld(vorgabe.HTTPS, umgebung.HTTPSProxy),
		Aus:   feld(vorgabe.Aus, umgebung.NoProxy),
	}
}

// feld entscheidet ein einzelnes Feld.
func feld(eingestellt, ausDerUmgebung string) Proxyfeld {
	if strings.TrimSpace(eingestellt) != "" {
		return Proxyfeld{Wert: eingestellt, Herkunft: AusDerEinstellung}
	}

	if strings.TrimSpace(ausDerUmgebung) != "" {
		return Proxyfeld{Wert: ausDerUmgebung, Herkunft: AusDerUmgebung}
	}

	return Proxyfeld{Herkunft: OhneWert}
}

// Proxyauswahl gibt die Auswahlfunktion für einen **fremden** Transport.
//
// **Für `internal/anmeldung`**, das ein Blatt ist und dieses Paket nicht
// kennen darf (N5, #1571). Es bekommt die Gestalt, die `http.Transport.Proxy`
// erwartet, und nichts darüber hinaus — damit gilt dort dieselbe Einstellung
// wie überall, ohne dass eine Schichtenkante entsteht.
func Proxyauswahl(quelle Proxyquelle) func(*http.Request) (*url.URL, error) {
	return proxyAus(quelle)
}

// proxyAus baut die Auswahlfunktion für einen Transport.
//
// **Je Anfrage gerechnet und nicht einmal.** Das kostet eine Übersetzung von
// drei Feldern je Anfrage — gemessen an einer Netzrunde ist das nichts, und
// dafür wirkt eine Einstellung ohne Neustart. Die frühere Fassung baute die
// Auswahl beim Klienten und trug den Satz „je Klient und nicht je Anfrage"
// als Einschränkung; sie ist hier eingelöst.
func proxyAus(quelle Proxyquelle) func(*http.Request) (*url.URL, error) {
	return func(anfrage *http.Request) (*url.URL, error) {
		stand := StandMit(anfrage.Context(), quelle)

		auswahl := (&httpproxy.Config{
			HTTPProxy:  stand.HTTP.Wert,
			HTTPSProxy: stand.HTTPS.Wert,
			NoProxy:    stand.Aus.Wert,

			// **`CGIMode` bleibt aus.** Es ignoriert `HTTP_PROXY` in einer
			// CGI-Umgebung, weil dort ein Angreifer die Variable über einen
			// Kopfsatz setzen könnte. Dieses Programm ist kein CGI-Skript,
			// und eingeschaltet würde `HTTP_PROXY` still wirkungslos, sobald
			// jemand `REQUEST_METHOD` in der Umgebung hat.
		}).ProxyFunc()

		return auswahl(anfrage.URL) //nolint:wrapcheck // der Transport erwartet den Fehler der Auswahl
	}
}
