package fernleitung

// **Der Weg zum Relais geht über denselben Proxy wie jeder andere** —
// NET-02.AK-01 (#1706).
//
// # Der Befund
//
// **Er ging über keinen.** tailscale fragt vor dem Wählen eines DERP-Relais
// `feature.HookProxyFromEnvironment` (`derp/derphttp/derphttp_client.go`,
// `dialNode`, Fassung v1.103.0-pre.0.20260904030409-31d8badb3bfb); gesetzt
// wird der Haken von `tailscale.com/feature/useproxy`, und das ist in
// diesem Programm nicht gebunden — `go list -deps .` am 2026-09-30 nennt
// `derp/derphttp`, aber nicht `feature/useproxy`. In einem Netz, das nur
// über einen Proxy hinauslässt, kam der Tunnel deshalb nie zustande.
//
// **Und `useproxy` einzubinden hätte nicht gereicht**: es liest die Umgebung
// und nicht die Einstellungen `netz.proxy.*`. Gesetzt wird der Haken hier,
// mit der Auswahl, die jeder andere Klient dieses Programms bekommt.
//
// **Nicht parallel**: der Haken ist einer je Prozess.

import (
	"net/http"
	"net/url"
	"testing"

	"tailscale.com/feature"
)

func einRelais() *http.Request {
	return &http.Request{URL: &url.URL{Scheme: "https", Host: "derp1.tailscale.com", Path: "/"}}
}

func TestDerWegZumRelaisNimmtDieProxywahl(t *testing.T) {
	gefragt := ""

	MitProxywahl(func(anfrage *http.Request) (*url.URL, error) {
		gefragt = anfrage.URL.Host

		return url.Parse("http://proxy.firma:3128")
	})

	wahl, gesetzt := feature.HookProxyFromEnvironment.GetOk()
	if !gesetzt {
		t.Fatal("der Haken ist nicht gesetzt")
	}

	adresse, fehler := wahl(einRelais())
	if fehler != nil || adresse == nil || adresse.Host != "proxy.firma:3128" {
		t.Fatalf("die Wahl gab %v, %v", adresse, fehler)
	}

	if gefragt != "derp1.tailscale.com" {
		t.Errorf("gefragt wurde nach %q", gefragt)
	}

	// **Ein zweiter Ruf ersetzt die Wahl**, statt tailscale ein zweites Mal
	// zu setzen — das hielte `Set` mit einem panic an.
	MitProxywahl(func(*http.Request) (*url.URL, error) { return nil, nil })

	if adresse, fehler := wahl(einRelais()); adresse != nil || fehler != nil {
		t.Errorf("nach dem Ersetzen gilt noch %v, %v", adresse, fehler)
	}
}

// **Ohne Wahl geht es ohne Proxy** — so wie bisher.
func TestOhneProxywahlGehtEsDirekt(t *testing.T) {
	vorher := proxywahl.Swap(nil)
	t.Cleanup(func() { proxywahl.Store(vorher) })

	if adresse, fehler := relaisproxy(einRelais()); adresse != nil || fehler != nil {
		t.Errorf("ohne Wahl: %v, %v", adresse, fehler)
	}
}

// **Die Anmeldung am Proxy kommt aus seiner Adresse** — dieselbe Regel wie
// `tshttpproxy.GetAuthHeader` in derselben Fassung: Basic, und nur mit Name
// **und** Kennwort. Aus den Einstellungen kommt keines (S13.NAT-02); aus der
// Umgebung schon.
func TestDieAnmeldungAmProxyKommtAusDerAdresse(t *testing.T) {
	t.Parallel()

	for adresse, erwartet := range map[string]string{
		"http://anna:geheim@proxy.firma:3128": "Basic YW5uYTpnZWhlaW0=",
		"http://anna@proxy.firma:3128":        "",
		"http://proxy.firma:3128":             "",
	} {
		zerlegt, fehler := url.Parse(adresse)
		if fehler != nil {
			t.Fatal(fehler)
		}

		if kopf, fehler := proxyanmeldung(zerlegt); kopf != erwartet || fehler != nil {
			t.Errorf("%s gab %q, %v, erwartet %q", adresse, kopf, fehler, erwartet)
		}
	}
}
