// **Über welche Schnittstellen die Entdeckung geht** — NET-02.NAT-01
// (#1706).
//
// # Der Anlass
//
// **Zwei Anwendungen im selben WLAN fanden einander nicht**, und der Ausweg
// war, `en0` von Hand zu nennen (`MSE_ENTDECKUNGSSCHNITTSTELLE`). Der
// IPv6-Rückfall
// (#1774)
// behebt den **Abbruch** — hier geht es um die **Wahl**.
//
// **`hashicorp/mdns@v1.0.7` ruft `setInterface` nur, wenn eine genannt ist.**
// Sonst nimmt das Betriebssystem seine „system-assigned multicast
// interface". Beim Ankündigen dasselbe: `net.ListenMulticastUDP("udp4", nil,
// …)` horcht nur auf der vom System zugewiesenen — und **wer eine Frage vom
// WLAN nicht hört, beantwortet sie auch nicht**.
//
// **Gemessen am 2026-09-21**: dieser Rechner führt acht Schnittstellen, und
// **genau eine** taugt — `en0` mit `192.168.178.74`, das WLAN. Genau die,
// die von Hand eingetragen wurde. Auf einem Rechner mit VPN oder mit Kabel
// und WLAN zugleich ist „die vom System zugewiesene" ein Glücksspiel.

package entdeckung

import (
	"net"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schnittstellenLesen ist die Naht zum Betriebssystem — ein Lauf hängt sich
// daran und braucht dafür kein Netz.
var schnittstellenLesen = net.Interfaces

// tauglicheLesen ist die Naht, an der ein Lauf die Wahl ersetzt.
var tauglicheLesen = Taugliche

// Taugliche gibt die Schnittstellen, über die mDNS gehen kann.
//
// **Eine leere Liste ist kein Fehler.** Ein Rechner ohne Netz hat keine, und
// die Entdeckung fällt dann auf die Wahl des Betriebssystems zurück — das
// ist, was vorher immer galt.
func Taugliche() ([]net.Interface, error) {
	alle, fehler := schnittstellenLesen()
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Netzschnittstellen dieses Rechners lesen")
	}

	taugliche := make([]net.Interface, 0, len(alle))

	for _, schnitt := range alle {
		// **Über dieselbe Naht wie die Ankündigung** — `adresslisteVon`
		// ist die Stelle, an der dieses Paket das Betriebssystem nach
		// Adressen fragt, und ein Lauf hängt sich daran.
		//
		// **Ein Lesefehler an einer Schnittstelle nimmt nicht die anderen
		// mit.** Eine, deren Adressen sich nicht lesen lassen, ist eine
		// weniger — kein Grund, gar nicht zu suchen.
		adressen, fehler := adresslisteVon(&schnitt)
		if fehler != nil {
			continue
		}

		if taugt(schnitt, adressen) {
			taugliche = append(taugliche, schnitt)
		}
	}

	return taugliche, nil
}

// taugt sagt, ob über diese Schnittstelle mDNS gehen kann.
//
// **Vier Gründe, und jeder einzeln nachzulesen:**
//
//   - **Abgeschaltet** trägt nichts — und das ist der Normalfall: dieser
//     Rechner führt acht Schnittstellen und benutzt eine.
//   - **Ohne Multicast** gibt es kein mDNS; es *ist* Multicast.
//   - **Die Rückschleife** findet nur diesen Rechner. Wer sie mitnähme,
//     machte aus einer leeren Suche eine, die sich selbst findet.
//   - **Ohne IPv4** kann der Rückfall aus #1774 nicht tragen: er fragt
//     ausdrücklich ohne IPv6, und dafür braucht es eine IPv4-Adresse.
func taugt(schnitt net.Interface, adressen []net.Addr) bool {
	if schnitt.Flags&net.FlagUp == 0 ||
		schnitt.Flags&net.FlagMulticast == 0 ||
		schnitt.Flags&net.FlagLoopback != 0 {
		return false
	}

	for _, adresse := range adressen {
		if netz, taugt := adresse.(*net.IPNet); taugt && netz.IP.To4() != nil {
			return true
		}
	}

	return false
}
