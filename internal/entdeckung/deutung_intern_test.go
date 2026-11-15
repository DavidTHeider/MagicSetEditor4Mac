package entdeckung

// **Was aus einer Antwort wird — und was nicht** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum diese Läufe **im** Paket stehen
//
// **Weil sie Antworten prüfen, die kein echtes Netz liefert.** Auf
// `_mse4mac._tcp` antwortet dieses Programm — eine Antwort ohne Kennung, mit
// krummer Kennung oder ohne Adresse käme nur von jemandem, der es darauf
// anlegt. Sie hier zusammenzusetzen ist der einzige Weg, den Fehlerweg zu
// betreten, ohne ein fremdes Gerät zu brauchen.

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/hashicorp/mdns"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// eineAntwort baut, was die Bibliothek liefert.
func eineAntwort(felder []string, anders func(*mdns.ServiceEntry)) *mdns.ServiceEntry {
	antwort := &mdns.ServiceEntry{
		Name: "probe", Host: "probe.local.", Port: 41234,
		AddrV4: net.IPv4(127, 0, 0, 1), InfoFields: felder,
	}

	if anders != nil {
		anders(antwort)
	}

	return antwort
}

// **Was keine Peerkennung trägt, wird verworfen und gezählt.**
func TestWasKeinePeerkennungTraegtWirdVerworfen(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	gut := kennungsmarke + string(ich.Kennung())

	faelle := map[string]*mdns.ServiceEntry{
		"gar keine Felder": eineAntwort(nil, nil),
		"nur ein Name":     eineAntwort([]string{namensmarke + "Wer"}, nil),
		"krumme Kennung":   eineAntwort([]string{kennungsmarke + "unsinn"}, nil),
		"leere Kennung":    eineAntwort([]string{kennungsmarke}, nil),
		"ohne Port": eineAntwort([]string{gut}, func(a *mdns.ServiceEntry) {
			a.Port = 0
		}),
		"ohne jede Adresse": eineAntwort([]string{gut}, func(a *mdns.ServiceEntry) {
			a.AddrV4, a.AddrV6, a.Host = nil, nil, ""
		}),
	}

	for name, antwort := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fund, gilt := deute(antwort); gilt {
				t.Fatalf("es kam durch: %+v", fund)
			}
		})
	}
}

// **Und die verworfenen werden gezählt, nicht verschwiegen.**
func TestDieVerworfenenWerdenGezaehlt(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	antworten := make(chan *mdns.ServiceEntry, 4)
	antworten <- eineAntwort(nil, nil)
	antworten <- eineAntwort([]string{kennungsmarke + "unsinn"}, nil)
	antworten <- eineAntwort([]string{
		kennungsmarke + string(ich.Kennung()), namensmarke + "Annas Mac",
	}, nil)
	close(antworten)

	ergebnis := sammle(antworten)

	if ergebnis.Verworfen != 2 {
		t.Errorf("verworfen wurden %d", ergebnis.Verworfen)
	}

	if len(ergebnis.Funde) != 1 || ergebnis.Funde[0].Name != "Annas Mac" {
		t.Fatalf("gefunden wurde %+v", ergebnis.Funde)
	}

	if ergebnis.Funde[0].Adresse() != "127.0.0.1:41234" {
		t.Errorf("die Adresse ist %q", ergebnis.Funde[0].Adresse())
	}
}

// **Zweimal dieselbe Kennung ist einmal.**
//
// **Weil mDNS über IPv4 und IPv6 antwortet** und beide Wege zu demselben
// Rechner führen. Wer beide zeigte, ließe jemanden zwischen zwei Zeilen
// wählen, die dasselbe meinen.
func TestZweimalDieselbeKennungIstEinmal(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	felder := []string{kennungsmarke + string(ich.Kennung())}

	antworten := make(chan *mdns.ServiceEntry, 2)
	antworten <- eineAntwort(felder, nil)
	antworten <- eineAntwort(felder, func(a *mdns.ServiceEntry) {
		a.AddrV4, a.AddrV6 = nil, net.ParseIP("::1")
	})
	close(antworten)

	ergebnis := sammle(antworten)

	if len(ergebnis.Funde) != 1 {
		t.Fatalf("gefunden wurden %d", len(ergebnis.Funde))
	}

	// **Die erste gewinnt, und das ist IPv4** — sie steht in der Antwort
	// zuerst, und eine IPv4 erreicht heute mehr Wege als eine IPv6.
	if ergebnis.Funde[0].Wirt != "127.0.0.1" {
		t.Errorf("der Wirt ist %q", ergebnis.Funde[0].Wirt)
	}
}

// **Ohne IPv4 gilt die IPv6, und ohne beide der Name.**
func TestOhneIPv4GiltDieIPv6UndDannDerName(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		anders   func(*mdns.ServiceEntry)
		erwartet string
	}{
		"nur IPv6": {func(a *mdns.ServiceEntry) {
			a.AddrV4, a.AddrV6 = nil, net.ParseIP("2001:db8::1")
		}, "2001:db8::1"},
		"nur der Name": {func(a *mdns.ServiceEntry) {
			a.AddrV4, a.AddrV6 = nil, nil
		}, "probe.local"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if wirt := wirtAus(eineAntwort(nil, fall.anders)); wirt != fall.erwartet {
				t.Fatalf("der Wirt ist %q", wirt)
			}
		})
	}
}

// **Eine IPv6 wird richtig geklammert.**
//
// **`net.JoinHostPort` tut es**, und ein selbst zusammengesetztes
// `wirt + ":" + port` täte es nicht — dann stünde `2001:db8::1:41234` da, und
// das ist nach [domain.PruefeAdresse] keine Adresse.
func TestEineIPv6WirdGeklammert(t *testing.T) {
	t.Parallel()

	fund := Fund{Wirt: "2001:db8::1", Port: 41234}

	if fund.Adresse() != "[2001:db8::1]:41234" {
		t.Fatalf("die Adresse ist %q", fund.Adresse())
	}
}

// **Ohne Schnittstelle kündigt sich dieser Rechner unter allen an.**
func TestOhneSchnittstelleGiltJedeAdresse(t *testing.T) {
	t.Parallel()

	adressen, fehler := adressenVon(nil)
	if fehler != nil {
		t.Fatalf("adressenVon: %v", fehler)
	}

	if len(adressen) == 0 {
		t.Fatal("dieser Rechner hat keine Adresse")
	}
}

// **Und eine Schnittstelle, die es nicht gibt, ist ein Befund.**
//
// **Ein Index, den kein Gerät trägt** — `Addrs()` fragt danach beim
// Betriebssystem, und die Antwort ist ein Fehler und keine leere Liste.
func TestEineSchnittstelleDieEsNichtGibtIstEinBefund(t *testing.T) {
	t.Parallel()

	erfunden := &net.Interface{Index: 9999, Name: "gibtesnicht0"}

	if _, fehler := adressenVon(erfunden); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Eine Anmeldung über eine erfundene Schnittstelle kommt nicht zustande.**
func TestEineAnmeldungUeberEineErfundeneSchnittstelleScheitert(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if _, fehler := Melde(Angebot{
		Kennung: ich.Kennung(), Port: 41234,
		Schnittstelle: &net.Interface{Index: 9999, Name: "gibtesnicht0"},
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Und eine Suche darüber ebenso.**
func TestEineSucheUeberEineErfundeneSchnittstelleScheitert(t *testing.T) {
	t.Parallel()

	if _, fehler := Suche(t.Context(), Frage{
		Frist:         100 * time.Millisecond,
		Schnittstelle: &net.Interface{Index: 9999, Name: "gibtesnicht0"},
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Eine Schnittstelle lässt sich beim Namen finden — oder eben nicht.**
//
// **Sie steht hier und nicht in `net`**, weil `internal/app` `net` nicht
// kennen darf (§2) und den Namen trotzdem auflösen können muss.
func TestEineSchnittstelleLaesstSichBeimNamenFinden(t *testing.T) {
	t.Parallel()

	schnitt, fehler := SchnittstelleNamens("lo0")
	if fehler != nil {
		t.Skipf("dieser Rechner hat kein lo0: %v", fehler)
	}

	if schnitt.Name != "lo0" {
		t.Fatalf("gefunden wurde %q", schnitt.Name)
	}

	if _, fehler := SchnittstelleNamens("gibtesnicht0"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Ein Fund trägt beides: die Adresse und den Namen** —
// #1633.
//
// # Warum beides und nicht eines
//
// **Zum Wählen ist die Adresse schneller, zum Merken ist der Name
// haltbarer.** Eine IP von heute ist nach dem nächsten DHCP-Wechsel falsch —
// und der Kontakt zeigte dann auf einen Rechner, der inzwischen jemand
// anderem gehört. Genau davor warnt `PeerkontaktAdresse` in ihrer eigenen
// Dokumentation.
func TestEinFundTraegtAdresseUndNamen(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	fund, gilt := deute(eineAntwort(
		[]string{kennungsmarke + string(ich.Kennung())}, nil))
	if !gilt {
		t.Fatal("der Fund kam nicht durch")
	}

	if fund.Wirt != "127.0.0.1" {
		t.Errorf("der Wirt ist %q", fund.Wirt)
	}

	// **Ohne den Punkt am Ende** — `probe.local.` ist die Drahtform, und
	// `net.JoinHostPort` macht daraus sonst einen Namen, den niemand tippt.
	if fund.Wirtname != "probe.local" {
		t.Errorf("der Name ist %q", fund.Wirtname)
	}

	if fund.Adresse() != "127.0.0.1:41234" {
		t.Errorf("die Adresse ist %q", fund.Adresse())
	}

	if fund.Merkadresse() != "probe.local:41234" {
		t.Errorf("die Merkadresse ist %q", fund.Merkadresse())
	}
}

// **Und ein Fund ohne Namen behält seine Adresse.**
//
// **Sonst verlöre der Fall ohne Namen seinen Weg** — das Issue verlangt es
// ausdrücklich: gemerkt wird der Name, **wo es einen gibt**.
func TestEinFundOhneNamenBehaeltSeineAdresse(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	fund, gilt := deute(eineAntwort(
		[]string{kennungsmarke + string(ich.Kennung())},
		func(a *mdns.ServiceEntry) { a.Host = "" }))
	if !gilt {
		t.Fatal("der Fund kam nicht durch")
	}

	if fund.Wirtname != "" {
		t.Errorf("da steht ein Name: %q", fund.Wirtname)
	}

	if fund.Merkadresse() != fund.Adresse() {
		t.Errorf("die Merkadresse ist %q statt %q",
			fund.Merkadresse(), fund.Adresse())
	}
}

// **Ein Fund, der nur einen Namen hat, wählt über ihn.**
//
// **Das ist der Weg, den `wirtAus` schon kannte** — ohne Adresse bleibt der
// Name, und dann sind Adresse und Merkadresse dieselbe.
func TestEinFundOhneAdresseWaehltUeberDenNamen(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	fund, gilt := deute(eineAntwort(
		[]string{kennungsmarke + string(ich.Kennung())},
		func(a *mdns.ServiceEntry) { a.AddrV4, a.AddrV6 = nil, nil }))
	if !gilt {
		t.Fatal("der Fund kam nicht durch")
	}

	if fund.Wirt != "probe.local" || fund.Merkadresse() != fund.Adresse() {
		t.Errorf("Wirt %q, Merkadresse %q, Adresse %q",
			fund.Wirt, fund.Merkadresse(), fund.Adresse())
	}
}
