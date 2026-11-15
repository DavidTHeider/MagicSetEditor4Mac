package fernleitung

// **Der Weg zum Relais über den eingestellten Proxy** — NET-02.AK-01
// (#1706).
// Der Befund und seine Messung stehen im Kopf von `proxy_intern_test.go`.
//
// # Was hier nicht geschieht
//
// **Die Gegenstelle selbst geht nie über den Proxy.** WireGuard spricht
// UDP, und ein HTTP-Proxy trägt kein UDP; über ihn geht allein der Weg zum
// DERP-Relais, und dann spricht die Verbindung über das Relais. Ein
// direkter Weg entsteht in einem solchen Netz nicht.

import (
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"tailscale.com/feature"
)

// Proxywahl ist die Gestalt, die `http.Transport.Proxy` erwartet.
type Proxywahl = func(*http.Request) (*url.URL, error)

var (
	proxywahl   atomic.Pointer[Proxywahl]
	hakenEinmal sync.Once
)

// MitProxywahl setzt, über welchen Proxy der Weg zu einem Relais geht.
//
// **Der Haken von tailscale ist einer je Prozess** und lässt sich nur einmal
// setzen; ein zweiter Ruf ersetzt deshalb die Wahl dahinter. Gerechnet wird
// je Anfrage — eine geänderte Einstellung wirkt beim nächsten Wählen.
func MitProxywahl(wahl Proxywahl) {
	proxywahl.Store(&wahl)

	hakenEinmal.Do(func() {
		feature.HookProxyFromEnvironment.Set(relaisproxy)
		feature.HookProxyGetAuthHeader.Set(proxyanmeldung)
	})
}

// relaisproxy fragt die gesetzte Wahl — ohne eine geht es direkt.
func relaisproxy(anfrage *http.Request) (*url.URL, error) {
	wahl := proxywahl.Load()
	if wahl == nil {
		return nil, nil //nolint:nilnil // „kein Proxy" ist die Antwort, die tailscale erwartet
	}

	return (*wahl)(anfrage)
}

// proxyanmeldung ist der Kopf `Proxy-Authorization` für eine Adresse.
//
// **Dieselbe Regel wie `tshttpproxy.GetAuthHeader`** in der gepinnten
// Fassung: Basic, und nur mit Name **und** Kennwort.
func proxyanmeldung(adresse *url.URL) (string, error) {
	kennwort, mitKennwort := adresse.User.Password()
	if adresse.User.Username() == "" || !mitKennwort {
		return "", nil
	}

	anfrage := &http.Request{Header: make(http.Header)}
	anfrage.SetBasicAuth(adresse.User.Username(), kennwort)

	return anfrage.Header.Get("Authorization"), nil
}
