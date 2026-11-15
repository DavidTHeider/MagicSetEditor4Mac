package entdeckung

// **Die Schnittstellen dieses Rechners, zum Wählen** — G08
// (#2272).
//
// Über eingesetzte Schnittstellen und Adressen, nicht über die dieses
// Rechners (C06, #2231): welche es hier gibt, hängt an WLAN, VPN und einer
// laufenden VM. Gegen den Rechner läuft nur die Zusage, dass es die Schleife
// gibt.

import (
	"errors"
	"net"
	"reflect"
	"testing"
)

func adressenAus(t *testing.T, texte ...string) []net.Addr {
	t.Helper()

	adressen := make([]net.Addr, 0, len(texte))
	for _, text := range texte {
		ip, netz, fehler := net.ParseCIDR(text)
		if fehler != nil {
			t.Fatalf("%q: %v", text, fehler)
		}

		netz.IP = ip
		adressen = append(adressen, netz)
	}

	return adressen
}

func TestDieSchnittstellenlisteNenntArtZustandUndAdressen(t *testing.T) {
	t.Parallel()

	rechner := []net.Interface{
		{Index: 1, Name: "lo0", Flags: net.FlagUp | net.FlagLoopback},
		{Index: 2, Name: "en0", Flags: net.FlagUp | net.FlagMulticast},
		{Index: 3, Name: "en5", Flags: 0},
		{Index: 4, Name: "utun3", Flags: net.FlagUp},
		{Index: 5, Name: "bridge100", Flags: net.FlagUp},
		{Index: 6, Name: "awdl0", Flags: net.FlagUp},
		{Index: 7, Name: "xyz9", Flags: net.FlagUp},
	}
	adressen := map[string][]net.Addr{
		"lo0":       adressenAus(t, "::1/128", "127.0.0.1/8"),
		"en0":       adressenAus(t, "fe80::1/64", "192.168.178.74/24", "2001:db8::5/64"),
		"en5":       nil,
		"utun3":     adressenAus(t, "10.8.0.2/24"),
		"bridge100": adressenAus(t, "192.168.64.1/24"),
		"awdl0":     adressenAus(t, "fe80::2/64"),
	}

	liste, fehler := schnittstellenlisteAus(
		func() ([]net.Interface, error) { return rechner, nil },
		func(schnitt *net.Interface) ([]net.Addr, error) {
			if schnitt.Name == "xyz9" {
				return nil, errors.New("nicht lesbar")
			}

			return adressen[schnitt.Name], nil
		})
	if fehler != nil {
		t.Fatal(fehler)
	}

	erwartet := []Schnittstelleneintrag{
		{Name: "lo0", Art: ArtSchleife, Aktiv: true, Adressen: []string{"127.0.0.1", "::1"}},
		{Name: "en0", Art: ArtNetz, Aktiv: true, Adressen: []string{"192.168.178.74", "2001:db8::5", "fe80::1%en0"}},
		{Name: "en5", Art: ArtNetz, Aktiv: false, Adressen: []string{}},
		{Name: "utun3", Art: ArtTunnel, Aktiv: true, Adressen: []string{"10.8.0.2"}},
		{Name: "bridge100", Art: ArtBruecke, Aktiv: true, Adressen: []string{"192.168.64.1"}},
		{Name: "awdl0", Art: ArtDirekt, Aktiv: true, Adressen: []string{"fe80::2%awdl0"}},
		{Name: "xyz9", Art: "", Aktiv: true, Adressen: []string{}, Befund: "nicht lesbar"},
	}

	if !reflect.DeepEqual(liste, erwartet) {
		t.Errorf("die Liste ist\n%+v\nerwartet\n%+v", liste, erwartet)
	}
}

func TestOhneLesbareSchnittstellenGibtEsEinenFehler(t *testing.T) {
	t.Parallel()

	_, fehler := schnittstellenlisteAus(
		func() ([]net.Interface, error) { return nil, errors.New("verweigert") },
		func(*net.Interface) ([]net.Addr, error) { return nil, nil })
	if fehler == nil {
		t.Fatal("ohne Schnittstellen kam kein Fehler")
	}
}

// **Gegen den Rechner nur die Schleife** — sie gibt es auf jedem.
func TestDieSchnittstellenlisteDiesesRechnersKenntDieSchleife(t *testing.T) {
	t.Parallel()

	liste, fehler := Schnittstellenliste()
	if fehler != nil {
		t.Fatal(fehler)
	}

	for _, eintrag := range liste {
		if eintrag.Art == ArtSchleife && len(eintrag.Adressen) > 0 {
			return
		}
	}

	t.Errorf("keine Schleife mit Adresse in %+v", liste)
}

// **Was keine Netzadresse ist, kommt nicht in die Liste** — eine Adresse
// ohne Netz (`*net.IPAddr`) und eine mit krummer Länge; das Betriebssystem
// liefert beides nicht, aber `net.Addr` erlaubt es.
func TestDieSchnittstellenlisteUebergehtKeineNetzadressen(t *testing.T) {
	t.Parallel()

	rohe := append([]net.Addr{
		&net.IPAddr{IP: net.ParseIP("192.0.2.1")},
		&net.IPNet{IP: net.IP{192, 0, 2}, Mask: net.CIDRMask(24, 32)},
	}, adressenAus(t, "192.0.2.10/24")...)

	if adressen := adresstexte("en0", rohe); !reflect.DeepEqual(adressen, []string{"192.0.2.10"}) {
		t.Errorf("die Adressen sind %v — erwartet nur 192.0.2.10", adressen)
	}
}
