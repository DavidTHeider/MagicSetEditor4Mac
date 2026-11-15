package entdeckung

// **Die Schnittstellen dieses Rechners, zum Wählen** — G08
// (#2272).
//
// Wer einen Dienst über die Schleife hinaus öffnet, soll die Adresse
// **wählen** und nicht abtippen. Dazu braucht es jede Schnittstelle mit
// Namen, Art, Zustand und Adressen — auch die, die gerade nicht taugt: eine
// Wahl, die eine ausgeschaltete Schnittstelle nicht zeigt, ließe jemanden
// raten, wo seine geblieben ist.
//
// # Die Art steht da, wo sie bekannt ist
//
// **Nach dem Namen, wie macOS ihn vergibt**, und nur dort: `lo` ist die
// Schleife, `utun`, `ipsec`, `ppp`, `gif` und `stf` sind Tunnel (VPN, 6to4),
// `bridge` eine Brücke (eine VM, Internetfreigabe), `awdl` und `llw` Apples
// Direktfunk. **`en` heißt nur „Netz"** — ob Kabel oder WLAN, sagt der Name
// nicht (`networksetup` wüsste es, und dafür einen Prozess zu starten, wäre
// mehr, als die Wahl braucht). Was keiner Regel folgt, bleibt ohne Art.
//
// # Adressen
//
// **IPv4 zuerst, dann IPv6**, und eine verbindungslokale IPv6-Adresse
// **mit ihrer Zone** (`fe80::1%en0`) — ohne sie lässt sich auf ihr nicht
// lauschen.

import (
	"net"
	"net/netip"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Arten einer Netzschnittstelle — leer, wo der Name keine verrät.
const (
	ArtSchleife = "schleife"
	ArtNetz     = "netz"
	ArtTunnel   = "tunnel"
	ArtBruecke  = "bruecke"
	ArtDirekt   = "direkt"
)

// Schnittstelleneintrag ist eine Netzschnittstelle dieses Rechners.
type Schnittstelleneintrag struct {
	Name     string
	Art      string
	Aktiv    bool
	Adressen []string

	// Befund sagt, warum die Adressen fehlen, wo sie nicht zu lesen waren —
	// **eine** Schnittstelle nimmt nicht die anderen mit.
	Befund string
}

// Schnittstellenliste liest die Schnittstellen dieses Rechners.
func Schnittstellenliste() ([]Schnittstelleneintrag, error) {
	return schnittstellenlisteAus(net.Interfaces, eigeneAdressliste)
}

func schnittstellenlisteAus(
	lesen func() ([]net.Interface, error), adressen func(*net.Interface) ([]net.Addr, error),
) ([]Schnittstelleneintrag, error) {
	alle, fehler := lesen()
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Netzschnittstellen dieses Rechners lesen")
	}

	liste := make([]Schnittstelleneintrag, 0, len(alle))

	for _, schnitt := range alle {
		eintrag := Schnittstelleneintrag{
			Name: schnitt.Name, Art: artVon(schnitt), Aktiv: schnitt.Flags&net.FlagUp != 0, Adressen: []string{},
		}

		rohe, lesefehler := adressen(&schnitt)
		if lesefehler != nil {
			eintrag.Befund = lesefehler.Error()
		}

		eintrag.Adressen = adresstexte(schnitt.Name, rohe)
		liste = append(liste, eintrag)
	}

	return liste, nil
}

func artVon(schnitt net.Interface) string {
	if schnitt.Flags&net.FlagLoopback != 0 {
		return ArtSchleife
	}

	for vorsatz, art := range map[string]string{
		"utun": ArtTunnel, "ipsec": ArtTunnel, "ppp": ArtTunnel, "gif": ArtTunnel, "stf": ArtTunnel,
		"bridge": ArtBruecke, "awdl": ArtDirekt, "llw": ArtDirekt, "en": ArtNetz,
	} {
		if strings.HasPrefix(schnitt.Name, vorsatz) {
			return art
		}
	}

	return ""
}

// adresstexte ordnet IPv4 vor IPv6 und eine verbindungslokale IPv6 mit Zone
// ans Ende.
func adresstexte(name string, rohe []net.Addr) []string {
	var vier, sechs, lokal []string

	for _, roh := range rohe {
		netz, istNetz := roh.(*net.IPNet)
		if !istNetz {
			continue
		}

		adresse, gut := netip.AddrFromSlice(netz.IP)
		if !gut {
			continue
		}

		adresse = adresse.Unmap()

		switch {
		case adresse.Is4():
			vier = append(vier, adresse.String())
		case adresse.IsLinkLocalUnicast():
			lokal = append(lokal, adresse.WithZone(name).String())
		default:
			sechs = append(sechs, adresse.String())
		}
	}

	// **Eine leere Liste und nicht `nil`** — `slices.Concat` gibt bei null
	// Elementen `nil`, und am Fenster käme `null` statt `[]` an (§3).
	if geordnet := slices.Concat(vier, sechs, lokal); geordnet != nil {
		return geordnet
	}

	return []string{}
}
