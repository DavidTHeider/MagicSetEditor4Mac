package netz

// **Was ein Transportfehler ist** — D02
// (#2233).
//
// Eingeordnet wird **nach Fehlertyp und Phase**, nicht nach Meldungstext
// (§8): ein `*net.DNSError` ist ein Name, der sich nicht auflöst; `ECONNREFUSED`
// ein Port ohne Annahme; eine Zeitüberschreitung beim Wählen ebenso; eine
// danach, solange der Kopf aussteht, die **Kopffrist**; eine beim Lesen des
// Rumpfes die **Stillstandsfrist**. **Ein Abriss** — die Gegenstelle nahm an
// und schloss (`EOF`) oder setzte zurück (`ECONNRESET`), vor dem Kopf oder
// mitten im Rumpf — ist ein eigener Fall (#2487); er hieß bis dahin
// „unerwarteter Fehler".
//
// **Ein Abbruch bleibt ein Abbruch — gefragt wird der Kontext, nicht der
// Fehler.** `net/http` lässt seine Kopffrist auf `errors.Is(…,
// context.DeadlineExceeded)` mit Ja antworten (`timeoutError.Is`, Go 1.27.2,
// `net/http/transport.go`); am Fehler allein wäre die Kopffrist von einer
// abgelaufenen Frist des Aufrufers nicht zu unterscheiden.
//
// **Und ohne Geheimnis**: der Grund geht ins Protokoll, und darin stand die
// ganze Adresse samt Abfrage. Sie wird hier auf Schema, Wirt und Pfad
// gekürzt — ein Schlüssel in der Abfrage oder Zugangsdaten vor dem Wirt
// erreichen weder Meldung noch Protokoll.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// deuteTransport ordnet einen Fehler ein, der vor dem Rumpf entstand.
func deuteTransport(ctx context.Context, ziel *url.URL, fristen Fristen, fehler error) error {
	if ctx.Err() != nil {
		return ohneGeheimnisIm(ziel, fehler)
	}

	redigiert := ohneGeheimnisIm(ziel, fehler)
	grund := nurSatz(redigiert)

	// **Schon eingeordnet** — vom Haken für die Antwort des Proxys.
	var eingeordnet *domain.Gegenstellenfehler
	if errors.As(fehler, &eingeordnet) {
		return eingeordnet
	}

	var name *net.DNSError

	var wahl *net.OpError

	switch {
	case errors.As(fehler, &name):
		return domain.NeuerGegenstellenfehler(domain.ErrNameUnaufloesbar, grund, ziel.Host, 0)
	case errors.Is(fehler, syscall.ECONNREFUSED),
		errors.As(fehler, &wahl) && wahl.Op == "dial" && wahl.Timeout():
		return domain.NeuerGegenstellenfehler(domain.ErrNichtErreichbar, grund, ziel.Host, 0)
	case zeitueberschreitung(fehler):
		return domain.NeuerGegenstellenfehler(domain.ErrKeineAntwort, grund, ziel.Host, griff(fristen, fehler))
	case abgerissen(fehler):
		return domain.NeuerGegenstellenfehler(domain.ErrVerbindungAbgerissen, grund, ziel.Host, 0)
	}

	// **Nicht eingeordnet, und die Kette bleibt** — eine Sperre des
	// Zielbereichs trägt ihre eigene Klasse.
	return fmt.Errorf("%s fragen: %w", ohneGeheimnis(ziel), redigiert)
}

// deuteRumpf ordnet einen Fehler beim Lesen des Rumpfes ein.
func deuteRumpf(ctx context.Context, ziel *url.URL, stillstand time.Duration, fehler error) error {
	// **Ein Abbruch bleibt ein Abbruch** — wie in [deuteTransport].
	switch {
	case ctx.Err() == nil && zeitueberschreitung(fehler):
		return domain.NeuerGegenstellenfehler(domain.ErrAntwortStockt, nurSatz(fehler), ziel.Host, stillstand)
	case ctx.Err() == nil && abgerissen(fehler):
		return domain.NeuerGegenstellenfehler(domain.ErrVerbindungAbgerissen, nurSatz(fehler), ziel.Host, 0)
	}

	return fehler
}

// abgerissen sagt, ob die Gegenstelle eine angenommene Verbindung schloss
// oder zurücksetzte. **Gemessen, nicht vermutet** (#2487, Go 1.27.2): vor
// dem Kopf kommt ein Schließen als `io.EOF` in der Kette an, ein Zurücksetzen
// als `ECONNRESET`, und eines mitten im Rumpf als `io.ErrUnexpectedEOF`.
func abgerissen(fehler error) bool {
	return errors.Is(fehler, io.EOF) || errors.Is(fehler, io.ErrUnexpectedEOF) ||
		errors.Is(fehler, syscall.ECONNRESET)
}

// griff nennt die Frist, die griff, solange der Kopf aussteht: die des
// Kopfes (`ResponseHeaderTimeout`) — oder die Lesefrist, die der Wähler als
// Stillstandsfrist auf die Verbindung legt, wenn sie kürzer ist.
func griff(fristen Fristen, fehler error) time.Duration {
	var lesen *net.OpError
	if errors.As(fehler, &lesen) && lesen.Op == "read" {
		return fristen.Stillstand
	}

	return fristen.Kopf
}

// nurSatz behält vom Grund den Satz und **nicht die Kette**.
//
// **Gemessen, nicht vermutet** (D02): die Kopffrist von `net/http` antwortet
// auf `errors.Is(…, context.DeadlineExceeded)` mit Ja, und [domain.MitGrund]
// hängt den Grund in die Kette. Jede Übersetzung fragt den Abbruch zuerst —
// `internal/app`, `internal/rest` —, und das Fenster blendet einen Abbruch
// aus. Der Grund ist fürs Protokoll; einordnen soll ihn niemand mehr.
func nurSatz(fehler error) error {
	return errors.New(fehler.Error()) //nolint:err113 // der Satz eines eingeordneten Grundes, nur fürs Protokoll
}

func zeitueberschreitung(fehler error) bool {
	var zeit net.Error

	return errors.As(fehler, &zeit) && zeit.Timeout()
}

// ohneGeheimnis gibt Schema, Wirt und Pfad — ohne Zugangsdaten und Abfrage.
func ohneGeheimnis(ziel *url.URL) string {
	gekuerzt := url.URL{Scheme: ziel.Scheme, Host: ziel.Host, Path: ziel.Path}

	return gekuerzt.String()
}

// ohneGeheimnisIm kürzt die Adresse in einem `*url.Error`, das der Klient
// zurückgab — auf die der Anfrage, auch nach einer Weiterleitung, deren Ziel
// ebenso einen Schlüssel tragen kann.
func ohneGeheimnisIm(ziel *url.URL, fehler error) error {
	var adressfehler *url.Error
	if errors.As(fehler, &adressfehler) {
		adressfehler.URL = ohneGeheimnis(ziel)
	}

	return fehler
}

// proxyanmeldung ist der Haken des Transports für die Antwort auf `CONNECT`
// (`http.Transport.OnProxyConnectResponse`, gerufen **vor** der Prüfung auf
// 200, Go 1.27.2): ohne ihn bliebe vom `407` nur der Statustext als Fehler.
func proxyanmeldung(_ context.Context, proxy *url.URL, _ *http.Request, antwort *http.Response) error {
	if antwort.StatusCode != http.StatusProxyAuthRequired {
		return nil
	}

	return domain.NeuerGegenstellenfehler(domain.ErrProxyAnmeldung,
		fmt.Errorf("CONNECT: %s", antwort.Status), proxy.Host, 0)
}

// proxyanmeldungAus ordnet ein `407` ein, das als **Antwort** kam — über einen
// Proxy für `http://`. Der Wirt ist der Proxy, den die Auswahl für diese
// Anfrage nennt; `URL.Host` trägt keine Zugangsdaten.
func (k *Klient) proxyanmeldungAus(anfrage *http.Request, antwort *http.Response) error {
	wirt := anfrage.URL.Host
	if proxy, fehler := proxyAus(k.proxy)(anfrage); fehler == nil && proxy != nil {
		wirt = proxy.Host
	}

	return domain.NeuerGegenstellenfehler(domain.ErrProxyAnmeldung,
		fmt.Errorf("Antwort: %s", antwort.Status), wirt, 0)
}
