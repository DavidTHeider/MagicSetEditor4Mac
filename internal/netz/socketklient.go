package netz

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"syscall"
)

// NeuerKlientUeberSocket gibt einen Klienten, der **jede** Anfrage über einen
// Unix-Socket schickt — der Name in der Adresse zählt nicht. AI-05
// (#2042):
// `fm serve --socket` ist der Weg zu Apples Gerätemodell, der keinen Port
// öffnet (Vermerk 0166).
//
// # Was ihn von [NeuerKlient] unterscheidet
//
//   - **Kein Vermittlungsrechner**, auch wenn die Umgebung einen nennt: ein
//     Proxy für einen Socket auf diesem Rechner schickte die Anfrage hinaus.
//   - **Keine Weiterleitung**: eine Antwort 3xx kommt als solche zurück — ein
//     Dienst am Socket hat niemanden, an den er weiterleiten dürfte.
//   - **Eine engere Zielprüfung** als [Zielbereich]: sie fragt nach Adressen
//     im Netz, und ein Socket hat keine. Gewählt wird **genau dieser Pfad**
//     ([socketkontrolle]) — derselbe Grund wie dort, ein Wähler ohne
//     `Control` wäre ein Weg an der Prüfung vorbei (#2000).
func NeuerKlientUeberSocket(fristen Fristen, pfad string) *Klient {
	return &Klient{unten: &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				verbindung, err := (&net.Dialer{
					Timeout: fristen.Verbindung, Control: socketkontrolle(pfad),
				}).DialContext(ctx, "unix", pfad)
				if err != nil {
					return nil, err //nolint:wrapcheck // der Aufrufer erwartet den Fehler des Wählers
				}

				return &fristverbindung{Conn: verbindung, frist: fristen.Stillstand}, nil
			},
			ResponseHeaderTimeout: fristen.Kopf,
		},
	}}
}

// socketkontrolle lässt nur den einen Socket zu — jedes andere Ziel ist
// gesperrt, auch ein anderer Socket.
func socketkontrolle(pfad string) func(netzart, adresse string, _ syscall.RawConn) error {
	return func(netzart, adresse string, _ syscall.RawConn) error {
		if netzart != "unix" || adresse != pfad {
			return fmt.Errorf("%w: der Socketklient wählt nur %s, nicht %s %s",
				ErrZielGesperrt, pfad, netzart, adresse)
		}

		return nil
	}
}
