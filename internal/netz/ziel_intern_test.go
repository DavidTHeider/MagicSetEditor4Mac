package netz

// **Die Bereiche der Zielsperre, ohne zu wählen** — SEC-P4
// (#2000).
//
// **Eine Tafel und kein Lauf gegen eine Gegenstelle**, weil die meisten
// dieser Adressen diesen Rechner verlassen würden: ein roter Lauf gegen
// `169.254.169.254` schickte ein Paket an den Metadatendienst des Netzes, in
// dem er gerade steht (`docs/PENTEST.md` §1, „nur lokal"). Dass die Tafel
// an der Verbindung gilt, belegt `ziel_test.go` mit den Schreibweisen, die
// hier bleiben.

import (
	"errors"
	"net/netip"
	"strings"
	"testing"
)

func TestDieZielsperreTeiltAdressenEin(t *testing.T) {
	t.Parallel()

	type fall struct {
		adresse  string
		bereich  Zielbereich
		schleife bool
		gesperrt bool
	}

	faelle := map[string]fall{
		"öffentlich v4 überall":           {"93.184.216.34", ZielUeberall, false, false},
		"öffentlich v4 öffentlich":        {"93.184.216.34", ZielOeffentlich, false, false},
		"öffentlich v6 öffentlich":        {"2606:2800:220:1::1", ZielOeffentlich, false, false},
		"Schleife überall":                {"127.0.0.1", ZielUeberall, false, false},
		"Schleife öffentlich":             {"127.0.0.1", ZielOeffentlich, false, true},
		"Schleife öffentlich, Schalter":   {"127.0.0.1", ZielOeffentlich, true, false},
		"Schleife 127/8 öffentlich":       {"127.255.0.9", ZielOeffentlich, false, true},
		"Schleife v6 öffentlich":          {"::1", ZielOeffentlich, false, true},
		"Schleife abgebildet öffentlich":  {"::ffff:127.0.0.1", ZielOeffentlich, false, true},
		"unbestimmt überall":              {"0.0.0.0", ZielUeberall, false, true},
		"unbestimmt 0/8 überall":          {"0.1.2.3", ZielUeberall, false, true},
		"unbestimmt v6 überall":           {"::", ZielUeberall, false, true},
		"unbestimmt, Schalter":            {"0.0.0.0", ZielOeffentlich, true, true},
		"Metadatendienst überall":         {"169.254.169.254", ZielUeberall, false, true},
		"Metadatendienst abgebildet":      {"::ffff:169.254.169.254", ZielUeberall, false, true},
		"verbindungslokal v6 überall":     {"fe80::1", ZielUeberall, false, true},
		"Gruppe überall":                  {"224.0.0.1", ZielUeberall, false, true},
		"Gruppe v6 überall":               {"ff02::1", ZielUeberall, false, true},
		"Rundruf überall":                 {"255.255.255.255", ZielUeberall, false, true},
		"reserviert überall":              {"240.0.0.1", ZielUeberall, false, true},
		"privat 10/8 überall":             {"10.0.0.1", ZielUeberall, false, false},
		"privat 10/8 öffentlich":          {"10.0.0.1", ZielOeffentlich, false, true},
		"privat 172.16/12 öffentlich":     {"172.31.255.1", ZielOeffentlich, false, true},
		"außerhalb 172.16/12 öffentlich":  {"172.32.0.1", ZielOeffentlich, false, false},
		"privat 192.168/16 öffentlich":    {"192.168.1.10", ZielOeffentlich, false, true},
		"privat, Schalter":                {"192.168.1.10", ZielOeffentlich, true, true},
		"geteilt 100.64/10 öffentlich":    {"100.64.0.1", ZielOeffentlich, false, true},
		"ULA fc00::/7 öffentlich":         {"fd12:3456::1", ZielOeffentlich, false, true},
		"standortlokal fec0::/10":         {"fec0::1", ZielOeffentlich, false, true},
		"Prüfnetz 198.18/15 öffentlich":   {"198.18.0.1", ZielOeffentlich, false, true},
		"NAT64 auf Metadaten überall":     {"64:ff9b::a9fe:a9fe", ZielUeberall, false, true},
		"NAT64 auf Schleife öffentlich":   {"64:ff9b::7f00:1", ZielOeffentlich, false, true},
		"NAT64 auf öffentlich öffentlich": {"64:ff9b::5db8:d822", ZielOeffentlich, false, false},
	}

	for name, f := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			grund := sperrgrund(netip.MustParseAddr(f.adresse), f.bereich, f.schleife)
			if (grund != "") != f.gesperrt {
				t.Errorf("%s: gesperrt %v (%q), erwartet %v", f.adresse, grund != "", grund, f.gesperrt)
			}
		})
	}
}

// **Die Leitung zum Proxy wird wie vom Transport gerechnet** — ohne Schema
// heißt `http://`, ohne Port der des Schemas. Eine Angabe, die sich nicht
// lesen lässt, ist keine Leitung: dann gilt für die Verbindung der Bereich
// des Klienten, und das ist die engere Wahl.
func TestDieLeitungZumProxyWirdWieVomTransportGerechnet(t *testing.T) {
	t.Parallel()

	for wert, erwartet := range map[string]string{
		"proxy.firma:3128":         "proxy.firma:3128",
		"127.0.0.1:3128":           "127.0.0.1:3128",
		"https://Proxy.Firma":      "proxy.firma:443",
		"http://proxy.firma":       "proxy.firma:80",
		"socks5://tor":             "tor:1080",
		"http://[::1]:8080":        "[::1]:8080",
		"ftp://ohne-port":          "",
		"":                         "",
		"%zz":                      "",
		"http://nutzer:geheim@p:1": "p:1",
	} {
		leitung, gelesen := proxyleitung(wert)
		if leitung != erwartet || gelesen != (erwartet != "") {
			t.Errorf("proxyleitung(%q) = %q, %v — erwartet %q", wert, leitung, gelesen, erwartet)
		}
	}
}

// **Eine Adresse, die sich nicht lesen lässt, wird abgewiesen** — der
// Wähler übergibt immer eine; fehlt sie doch, öffnet die Sperre nicht.
func TestEineUnlesbareVerbindungsadresseWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if fehler := kontrolleFuer(ZielUeberall)("tcp", "kein-ziel", nil); !errors.Is(fehler, ErrZielGesperrt) {
		t.Errorf("%v statt einer Sperre", fehler)
	}
}

// **Eine Adresse, die sich nicht zerlegen lässt, geht an der Vorprüfung
// vorbei** — der Bau der Anfrage meldet sie mit seinem eigenen Satz, und
// gewählt wird nichts.
func TestEineUnzerlegbareAdresseScheitertBeimBau(t *testing.T) {
	t.Parallel()

	klient := NeuerKlientNach(UeblicheFristen(), nil, ZielOeffentlich)

	if fehler := klient.darfFragen(t.Context(), "%zz"); fehler != nil {
		t.Errorf("die Vorprüfung: %v", fehler)
	}

	if _, fehler := Hole(t.Context(), klient, nil, Kenner{}, "%zz", 1); fehler == nil ||
		!strings.Contains(fehler.Error(), "bauen") {
		t.Errorf("%v statt eines Baufehlers", fehler)
	}
}
