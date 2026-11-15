package entdeckung

// **Der Rückfall auf IPv4 gilt je Schnittstelle** — C06
// (#2231).
//
// Die Fälle in `ipv4rueckfall_intern_test.go` laufen ohne Schnittstelle, die
// in `ueberalle_intern_test.go` ohne Rückfall. Ob eine Schnittstelle ohne
// IPv6-Route auf IPv4 zurückfällt, **während** die andere bei IPv6 bleibt,
// hielt keiner — und eine Suche, die beim ersten Rückfall für alle auf IPv4
// umschaltete, fiele durch keinen von beiden auf.

import (
	"context"
	"net"
	"testing"

	"github.com/hashicorp/mdns"
)

// einerSchnittstelle hängt genau eine taugliche an die Naht.
func einerSchnittstelle(t *testing.T) {
	t.Helper()

	zweiSchnittstellen(t)

	vorher := tauglicheLesen
	tauglicheLesen = func() ([]net.Interface, error) {
		alle, fehler := vorher()

		return alle[:1], fehler
	}

	t.Cleanup(func() { tauglicheLesen = vorher })
}

// keinerSchnittstelle lässt die Naht keine taugliche finden.
func keinerSchnittstelle(t *testing.T) {
	t.Helper()

	vorher := tauglicheLesen
	tauglicheLesen = func() ([]net.Interface, error) { return nil, nil }

	t.Cleanup(func() { tauglicheLesen = vorher })
}

// nameVon ist der Name einer Schnittstelle — oder leer, wenn das
// Betriebssystem wählt.
func nameVon(param *mdns.QueryParam) string {
	if param.Interface == nil {
		return ""
	}

	return param.Interface.Name
}

// rufeJeSchnittstelle zählt die Rufe je Name, getrennt nach IPv6 und IPv4.
func rufeJeSchnittstelle(rufe []mdns.QueryParam) map[string][2]int {
	zahl := map[string][2]int{}

	for _, ruf := range rufe {
		bisher := zahl[nameVon(&ruf)]
		if ruf.DisableIPv6 {
			bisher[1]++
		} else {
			bisher[0]++
		}

		zahl[nameVon(&ruf)] = bisher
	}

	return zahl
}

// en0 hat keine IPv6-Route, en5 schon — und wo das Betriebssystem wählt,
// auch keine.
func ohneIPv6AufEn0(_ int, param *mdns.QueryParam) error {
	if (nameVon(param) == "en0" || nameVon(param) == "") && !param.DisableIPv6 {
		return unerreichbar("udp6")
	}

	return nil
}

func TestDerRueckfallAufIPv4GiltJeSchnittstelle(t *testing.T) {
	for _, fall := range []struct {
		name string
		naht func(*testing.T)
		will map[string][2]int
	}{
		// **Keine taugliche**: es bleibt beim Betriebssystem, und auch dort
		// fällt die Frage ohne IPv6-Route auf IPv4 zurück.
		{"keine taugliche Schnittstelle", keinerSchnittstelle, map[string][2]int{"": {1, 1}}},
		{"eine Schnittstelle", einerSchnittstelle, map[string][2]int{"en0": {1, 1}}},
		{"zwei Schnittstellen, nur eine ohne IPv6", zweiSchnittstellen, map[string][2]int{"en0": {1, 1}, "en5": {1, 0}}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			fall.naht(t)
			rufe := mitRufen(t, ohneIPv6AufEn0)

			if _, fehler := Suche(context.Background(), Frage{}); fehler != nil {
				t.Fatalf("die Suche scheiterte: %v", fehler)
			}

			zahl := rufeJeSchnittstelle(*rufe)
			if len(zahl) != len(fall.will) {
				t.Fatalf("gefragt über %v, erwartet %v", zahl, fall.will)
			}

			for name, will := range fall.will {
				if zahl[name] != will {
					t.Errorf("%s: %v Rufe (IPv6, IPv4), erwartet %v", name, zahl[name], will)
				}
			}
		})
	}
}
