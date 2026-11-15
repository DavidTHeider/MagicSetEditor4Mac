package entdeckung

// **Über welche Schnittstellen die Entdeckung geht** — NET-02.NAT-01
// (#1706).
//
// # Der Befund
//
// **Zwei Anwendungen im selben WLAN fanden einander nicht**, und der Ausweg
// war, `en0` von Hand zu nennen. Der IPv6-Rückfall (#1774) behebt den
// **Abbruch**; hier geht es um die **Wahl**.
//
// `hashicorp/mdns@v1.0.7` ruft `setInterface` nur, wenn eine genannt ist —
// sonst nimmt das Betriebssystem seine „system-assigned multicast
// interface". Beim Ankündigen dasselbe: `net.ListenMulticastUDP("udp4",
// nil, …)` horcht nur auf der vom System zugewiesenen. **Wer eine Frage vom
// WLAN nicht hört, beantwortet sie auch nicht.**
//
// **Gemessen am 2026-09-21**: dieser Rechner führt acht Schnittstellen, und
// **genau eine** taugt — `en0` mit `192.168.178.74`, das WLAN. Genau die,
// die von Hand eingetragen wurde. Auf einem Rechner mit VPN oder Kabel und
// WLAN zugleich ist „die vom System zugewiesene" ein Glücksspiel.
//
// # Was hier geprüft wird und was nicht
//
// **Die Auswahlregel, ohne Netz.** Ob zwei Anwendungen einander wirklich
// finden, steht hinter dem Tag `netzlauf` — das braucht Multicast und die
// Erlaubnis *Lokales Netzwerk*.

import (
	"errors"
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die vier Gründe, aus denen eine Schnittstelle nicht taugt.**
func TestTauglicheSchnittstellenSiebtRichtig(t *testing.T) {
	t.Parallel()

	mitIPv4 := []net.Addr{&net.IPNet{IP: net.IPv4(192, 168, 1, 2)}}
	nurIPv6 := []net.Addr{&net.IPNet{IP: net.ParseIP("fe80::1")}}

	faelle := map[string]struct {
		schnitt  net.Interface
		adressen []net.Addr
		taugt    bool
	}{
		"aktiv, multicastfähig, mit IPv4": {
			net.Interface{Name: "en0", Flags: net.FlagUp | net.FlagMulticast},
			mitIPv4, true,
		},

		// **Eine abgeschaltete Schnittstelle trägt nichts** — und sie ist der
		// Normalfall: dieser Rechner führt acht und benutzt eine.
		"abgeschaltet": {
			net.Interface{Name: "en3", Flags: net.FlagMulticast},
			mitIPv4, false,
		},

		// **Ohne Multicast kein mDNS** — es ist Multicast, und sonst nichts.
		"ohne Multicast": {
			net.Interface{Name: "utun0", Flags: net.FlagUp},
			mitIPv4, false,
		},

		// **Die Rückschleife bleibt draußen.** Sie findet nur diesen
		// Rechner, und `docs/HAERTUNG.md` sagt, dass Multicast über sie
		// nicht Multicast über ein WLAN ist. Wer sie mitnähme, machte aus
		// einer leeren Suche eine, die sich selbst findet.
		"Rückschleife": {
			net.Interface{
				Name:  "lo0",
				Flags: net.FlagUp | net.FlagMulticast | net.FlagLoopback,
			},
			mitIPv4, false,
		},

		// **Ohne IPv4 kann der Rückfall nicht tragen** (#1774): er fragt
		// ausdrücklich ohne IPv6, und dafür braucht es eine IPv4-Adresse.
		"nur IPv6": {
			net.Interface{Name: "en1", Flags: net.FlagUp | net.FlagMulticast},
			nurIPv6, false,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if taugt := taugt(fall.schnitt, fall.adressen); taugt != fall.taugt {
				t.Errorf("taugt = %v, erwartet %v", taugt, fall.taugt)
			}
		})
	}
}

// **Und dieser Rechner hat mindestens eine.**
//
// **Kein Netzlauf**: `net.Interfaces()` liest die Schnittstellen des
// Rechners und schickt kein Paket. Was er belegt, ist die Regel an echten
// Daten statt an gebauten — und er sagt, wenn dieser Rechner gerade gar
// nicht im Netz ist, statt stillschweigend nichts zu prüfen.
func TestDieserRechnerHatEineTauglicheSchnittstelle(t *testing.T) {
	t.Parallel()

	taugliche, fehler := Taugliche()
	if fehler != nil {
		t.Fatalf("die Schnittstellen lesen: %v", fehler)
	}

	if len(taugliche) == 0 {
		t.Skip("dieser Rechner hat gerade keine taugliche Schnittstelle — " +
			"ohne Netz ist hier nichts zu prüfen")
	}

	for _, schnitt := range taugliche {
		if schnitt.Flags&net.FlagLoopback != 0 {
			t.Errorf("%q ist die Rückschleife", schnitt.Name)
		}
	}
}

// **Ein Lesefehler am Betriebssystem ist ein Fehler** — ohne die Liste der
// Schnittstellen gibt es keine Wahl zu treffen.
func TestOhneSchnittstellenlisteKlagtDieWahl(t *testing.T) {
	vorher := schnittstellenLesen
	schnittstellenLesen = func() ([]net.Interface, error) {
		return nil, errors.New("kein Zugriff")
	}

	t.Cleanup(func() { schnittstellenLesen = vorher })

	if _, fehler := Taugliche(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine Schnittstelle, deren Adressen klagen, ist eine weniger.**
//
// **Und nicht der Grund, gar nicht zu suchen**: auf einem Rechner mit acht
// Schnittstellen ist eine unlesbare der Normalfall, kein Ausnahmezustand.
func TestEineKlagendeSchnittstelleFaelltNurSelbstAus(t *testing.T) {
	vorher := schnittstellenLesen
	schnittstellenLesen = func() ([]net.Interface, error) {
		return []net.Interface{
			{Index: 1, Name: "klagt", Flags: net.FlagUp | net.FlagMulticast},
			{Index: 2, Name: "gut", Flags: net.FlagUp | net.FlagMulticast},
		}, nil
	}

	t.Cleanup(func() { schnittstellenLesen = vorher })

	vorherigeAdressen := adresslisteVon
	adresslisteVon = func(schnitt *net.Interface) ([]net.Addr, error) {
		if schnitt.Name == "klagt" {
			return nil, errors.New("keine Adressen")
		}

		return []net.Addr{&net.IPNet{IP: net.IPv4(192, 168, 1, 2)}}, nil
	}

	t.Cleanup(func() { adresslisteVon = vorherigeAdressen })

	taugliche, fehler := Taugliche()
	if fehler != nil {
		t.Fatalf("die Wahl scheiterte: %v", fehler)
	}

	if len(taugliche) != 1 || taugliche[0].Name != "gut" {
		t.Errorf("taugliche = %v, erwartet allein »gut«", taugliche)
	}
}
