// **Der Horcher auf der Schleifenadresse** — Werkstück 4 von Phase 4.12
// (#1351).
//
// # Warum ein eigener Server und kein eigenes Schema
//
// **Nachgelesen am 2026-09-05.** Google verlangt für Desktop-Anwendungen einen
// Rückruf auf die Schleifenadresse — „http://127.0.0.1:port or
// http://[::1]:port. Query your platform for the relevant loopback IP address
// and start an HTTP listener on a random available port" — und weist eigene
// Schemata ausdrücklich ab: „Custom URI schemes are no longer supported due to
// the risk of app impersonation."
//
// **Der Port ist frei gewählt und steht erst danach fest.** Deshalb kennt die
// Rückadresse der Horcher und nicht der Anbieter, und deshalb reicht sie
// [Beginne] herein.

package anmeldung

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Horcher nimmt den Rückruf des Browsers entgegen.
type Horcher struct {
	// Adresse ist, was als `redirect_uri` mitgeht.
	Adresse string

	lauscher net.Listener
	server   *http.Server
	werte    chan url.Values
	einmal   sync.Once
	zu       sync.Once

	// erwartet ist der `state` dieses Laufs ([Horcher.Erwarte]).
	schloss  sync.Mutex
	erwartet string

	// schliessen ist die Naht zum Zumachen — dieselbe Begründung wie bei
	// [lauschen]: ein `Close`, das klagt, ist im Betrieb möglich und im Test
	// nicht herbeizuführen.
	schliessen func() error
}

// lauschen ist die Naht zum Betriebssystem.
//
// **Als Feld und nicht als fester Aufruf**, aus demselben Grund wie `oeffnen`
// in `internal/app`: ein Horchen, das scheitert, kommt auf einem gesunden
// Rechner nicht zustande — im Betrieb aber schon, wenn alle Ports belegt sind
// oder eine Richtlinie es verbietet. Ein Fehlerpfad, den kein Test betritt,
// ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
var lauschen = net.Listen

// Horche öffnet einen Horcher auf einem freien Port der Schleifenadresse.
func Horche() (*Horcher, error) {
	lauscher, fehler := lauschen("tcp", "127.0.0.1:0")
	if fehler != nil {
		return nil, fmt.Errorf("auf der Schleifenadresse horchen: %w", fehler)
	}

	horcher := &Horcher{
		Adresse:  "http://" + lauscher.Addr().String() + "/",
		lauscher: lauscher,
		werte:    make(chan url.Values, 1),
	}

	horcher.server = &http.Server{
		Handler:           http.HandlerFunc(horcher.nimm),
		ReadHeaderTimeout: lesefrist,
	}

	horcher.schliessen = horcher.server.Close

	go func() { _ = horcher.server.Serve(lauscher) }() //nolint:errcheck // beim Schließen erwartet

	return horcher, nil
}

// lesefrist ist, wie lange ein Rückruf für seine Kopfzeilen hat.
//
// **Auch auf der Schleifenadresse.** Ein Horcher ohne Frist hängt an einer
// Verbindung, die niemand zu Ende schreibt — und der Port steht offen, solange
// die Anwendung läuft.
const lesefrist = 10 * time.Second

// nimm nimmt den Rückruf entgegen und antwortet dem Browser.
func (h *Horcher) nimm(schreiben http.ResponseWriter, anfrage *http.Request) {
	schreiben.Header().Set("Content-Type", "text/html; charset=utf-8")

	// **Nur der passende Rückruf zählt** (SEC-18, #2152): ein `GET` auf den
	// Weg der Rückadresse, mit dem `state` dieses Laufs und einem Code oder
	// einer Absage des Anbieters. Alles andere — ein `POST`, das `favicon.ico`
	// des Browsers, ein fremder `state` — verbrauchte bis dahin das eine Mal,
	// und der richtige kam nicht mehr an.
	if !h.passt(anfrage) {
		schreiben.WriteHeader(http.StatusBadRequest)

		_, _ = schreiben.Write([]byte(seiteFremd)) //nolint:errcheck // der Browser ist dann weg

		return
	}

	// **Genau einer zählt.** Ein Browser, der die Seite neu lädt, schickte
	// denselben Rückruf zweimal; der zweite ginge gegen einen Lauf, der schon
	// eingelöst ist — und der Anbieter weist einen verbrauchten Code ab.
	h.einmal.Do(func() { h.werte <- anfrage.URL.Query() })

	// **Der Browser bekommt einen Satz** (§6): ein leerer Bildschirm ließe
	// jemanden warten, der fertig ist.
	_, _ = schreiben.Write([]byte(seite)) //nolint:errcheck // der Browser ist dann weg
}

// passt sagt, ob ein Ruf der Rückruf dieses Laufs ist.
func (h *Horcher) passt(anfrage *http.Request) bool {
	h.schloss.Lock()
	erwartet := h.erwartet
	h.schloss.Unlock()

	werte := anfrage.URL.Query()

	return anfrage.Method == http.MethodGet && anfrage.URL.Path == "/" &&
		erwartet != "" && werte.Get("state") == erwartet &&
		(werte.Get("code") != "" || werte.Get("error") != "")
}

// seite ist, was im Browser stehen bleibt.
//
// **Ohne Bild, ohne Schrift, ohne Netz.** Was hier steht, wird von einem
// Programm ausgeliefert, das gleich wieder zumacht; ein Verweis nach draußen
// führte ins Leere, sobald der Port zu ist.
const seite = `<!doctype html><html lang="de"><head><meta charset="utf-8">` +
	`<title>MagicSetEditor4Mac</title></head><body style="font-family:system-ui;` +
	`max-width:32rem;margin:4rem auto;line-height:1.5">` +
	`<h1>Antwort erhalten</h1><p>MagicSetEditor4Mac hat die Antwort bekommen ` +
	`und schließt die Anmeldung jetzt ab. Dieses Fenster kann geschlossen ` +
	`werden.</p></body></html>`

// seiteFremd steht da, wenn ein Ruf nicht der erwartete Rückruf ist.
//
// **„Antwort erhalten" und nicht „Angemeldet"** (SEC-18): vor dem Einlösen
// des Codes ist niemand angemeldet — der Tausch kann noch scheitern.
const seiteFremd = `<!doctype html><html lang="de"><head><meta charset="utf-8">` +
	`<title>MagicSetEditor4Mac</title></head><body style="font-family:system-ui;` +
	`max-width:32rem;margin:4rem auto;line-height:1.5">` +
	`<h1>Kein Rückruf dieser Anmeldung</h1><p>Diese Anfrage gehört nicht zur ` +
	`laufenden Anmeldung von MagicSetEditor4Mac.</p></body></html>`

// Erwarte nennt den `state` des Laufs, dessen Rückruf zählt — SEC-18
// (#2152). Gerufen nach [Beginne] und bevor der Browser die Anmeldeseite
// öffnet; ohne ihn zählt kein Rückruf.
func (h *Horcher) Erwarte(state string) {
	h.schloss.Lock()
	defer h.schloss.Unlock()

	h.erwartet = state
}

// Warte gibt die Werte des Rückrufs — oder bricht ab.
func (h *Horcher) Warte(ctx context.Context) (url.Values, error) {
	select {
	case werte := <-h.werte:
		return werte, nil

	case <-ctx.Done():
		return nil, fmt.Errorf("auf den Rückruf warten: %w", ctx.Err())
	}
}

// Schliesse macht den Horcher zu.
//
// **Zweimal schließen ist kein Fehler**, denn der Aufrufer tut es doppelt:
// einmal im `defer`, einmal nach einem geglückten Rückruf. Eine Meldung über
// etwas, das in Ordnung ist, wäre eine Falschauskunft (§8).
func (h *Horcher) Schliesse() error {
	var fehler error

	h.zu.Do(func() {
		if schliessfehler := h.schliessen(); schliessfehler != nil &&
			!errors.Is(schliessfehler, http.ErrServerClosed) {
			fehler = fmt.Errorf("den Horcher schließen: %w", schliessfehler)
		}
	})

	return fehler
}
