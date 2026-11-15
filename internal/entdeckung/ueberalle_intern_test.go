package entdeckung

// **Gesucht und angekündigt wird über jede taugliche Schnittstelle** —
// NET-02.NAT-01
// (#1706).
//
// # Warum nicht eine
//
// **Weil „die vom System zugewiesene" ein Glücksspiel ist.** `mdns` ruft
// `setInterface` nur, wenn eine genannt ist; sonst wählt das Betriebssystem.
// Auf einem Rechner mit acht Schnittstellen — gemessen am 2026-09-21 — muss
// das nicht das WLAN sein, und dann geht die Frage ins falsche Netz.
//
// # Was hier geprüft wird
//
// **Die Verteilung und das Zusammenführen**, ohne Netz: wie viele Rufe es
// gibt, worüber sie gehen, und was passiert, wenn einer davon scheitert.

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/hashicorp/mdns"
)

// zweiSchnittstellen hängt zwei taugliche an die Naht.
func zweiSchnittstellen(t *testing.T) {
	t.Helper()

	gefunden := []net.Interface{
		{Index: 1, Name: "en0", Flags: net.FlagUp | net.FlagMulticast},
		{Index: 2, Name: "en5", Flags: net.FlagUp | net.FlagMulticast},
	}

	vorher := tauglicheLesen
	tauglicheLesen = func() ([]net.Interface, error) { return gefunden, nil }

	t.Cleanup(func() { tauglicheLesen = vorher })

	// **Und ihre Adressen kommen auch aus der Naht.** `net.Interface.Addrs()`
	// fragt über den **Index** das Betriebssystem — eine erfundene
	// Schnittstelle mit `Index: 1` traf auf diesem Rechner `lo0` und bekam
	// dessen Adressen. Ein Lauf, der versehentlich echte Schnittstellen
	// liest, prüft die Maschine und nicht den Code (§3).
	vorherigeAdressen := adresslisteVon
	adresslisteVon = func(*net.Interface) ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.IPv4(192, 168, 1, 2)}}, nil
	}

	t.Cleanup(func() { adresslisteVon = vorherigeAdressen })
}

func TestEineSucheGehtUeberJedeTauglicheSchnittstelle(t *testing.T) {
	zweiSchnittstellen(t)

	rufe := mitRufen(t, func(_ int, _ *mdns.QueryParam) error { return nil })

	if _, fehler := Suche(context.Background(), Frage{}); fehler != nil {
		t.Fatalf("die Suche scheiterte: %v", fehler)
	}

	if len(*rufe) != 2 {
		t.Fatalf("es gab %d Rufe, erwartet 2 — einen je Schnittstelle",
			len(*rufe))
	}

	namen := map[string]bool{}
	for _, ruf := range *rufe {
		if ruf.Interface == nil {
			t.Fatal("ein Ruf nannte keine Schnittstelle")
		}

		namen[ruf.Interface.Name] = true
	}

	for _, name := range []string{"en0", "en5"} {
		if !namen[name] {
			t.Errorf("über %q wurde nicht gefragt", name)
		}
	}
}

// **Eine Schnittstelle, die scheitert, nimmt die anderen nicht mit.**
//
// Ein Rechner hat ein VPN, das gerade keine Route hat — und findet deshalb
// niemanden im WLAN. Genau der Fall, für den es diese Arbeit gibt.
func TestEineGescheiterteSchnittstelleNimmtDieAnderenNichtMit(t *testing.T) {
	zweiSchnittstellen(t)

	mitRufen(t, func(_ int, param *mdns.QueryParam) error {
		if param.Interface != nil && param.Interface.Name == "en5" {
			return unerreichbar("udp4")
		}

		return nil
	})

	if _, fehler := Suche(context.Background(), Frage{}); fehler != nil {
		t.Errorf("die Suche scheiterte an einer Schnittstelle: %v", fehler)
	}
}

// **Scheitern alle, scheitert die Suche** — und sagt, woran.
//
// **Der Steuerlauf** (§10): ohne ihn wäre der Lauf darüber auch dann grün,
// wenn ein Fehler **nie** durchkäme — und eine Suche, die immer gelingt und
// nichts findet, ist die schlechteste aller Antworten.
func TestScheiternAlleSchnittstellenScheitertDieSuche(t *testing.T) {
	zweiSchnittstellen(t)

	mitRufen(t, func(_ int, _ *mdns.QueryParam) error {
		return errors.New("etwas ganz anderes")
	})

	if _, fehler := Suche(context.Background(), Frage{}); fehler == nil {
		t.Error("die Suche gelang, obwohl jede Schnittstelle scheiterte")
	}
}

// **Ohne taugliche Schnittstelle bleibt es beim Betriebssystem.**
//
// Das ist, was vorher immer galt — und auf einem Rechner ohne Netz ist es
// die einzige Antwort, die nicht lügt.
func TestOhneTauglicheSchnittstelleFragtSieWieBisher(t *testing.T) {
	vorher := tauglicheLesen
	tauglicheLesen = func() ([]net.Interface, error) { return nil, nil }

	t.Cleanup(func() { tauglicheLesen = vorher })

	rufe := mitRufen(t, func(_ int, _ *mdns.QueryParam) error { return nil })

	if _, fehler := Suche(context.Background(), Frage{}); fehler != nil {
		t.Fatalf("die Suche scheiterte: %v", fehler)
	}

	if len(*rufe) != 1 {
		t.Fatalf("es gab %d Rufe, erwartet 1", len(*rufe))
	}

	if (*rufe)[0].Interface != nil {
		t.Error("der Ruf nannte eine Schnittstelle, obwohl keine taugt")
	}
}

// **Eine genannte Schnittstelle schlägt alles.**
//
// `MSE_ENTDECKUNGSSCHNITTSTELLE` ist für Läufe da — und für den Menschen,
// der weiß, was er tut. Wer sie setzt, meint **diese** und keine andere.
func TestEineGenannteSchnittstelleSchlaegtDieWahl(t *testing.T) {
	zweiSchnittstellen(t)

	rufe := mitRufen(t, func(_ int, _ *mdns.QueryParam) error { return nil })

	gewollt := &net.Interface{Index: 9, Name: "lo0"}

	if _, fehler := Suche(
		context.Background(), Frage{Schnittstelle: gewollt},
	); fehler != nil {
		t.Fatalf("die Suche scheiterte: %v", fehler)
	}

	if len(*rufe) != 1 {
		t.Fatalf("es gab %d Rufe, erwartet 1", len(*rufe))
	}

	if (*rufe)[0].Interface != gewollt {
		t.Error("der Ruf ging nicht über die genannte Schnittstelle")
	}
}

// mitServern hängt eine Antwortfolge an die Servernaht und sammelt die Rufe.
func mitServern(
	t *testing.T, antworte func(*mdns.Config) (niederlegbar, error),
) *[]mdns.Config {
	t.Helper()

	rufe := &[]mdns.Config{}
	vorher := serverBauen

	serverBauen = func(config *mdns.Config) (*mdns.Server, error) {
		*rufe = append(*rufe, *config)

		_, fehler := antworte(config)

		return nil, fehler
	}

	t.Cleanup(func() { serverBauen = vorher })

	return rufe
}

// einAngebot ist eine Ankündigung, die durch die Prüfungen kommt.
//
// **Mit einer echten Peerkennung**: [Melde] prüft sie über den
// Fingerabdruck, und eine erfundene käme gar nicht bis zur Schnittstelle.
func einAngebot(t *testing.T) Angebot {
	t.Helper()

	return Angebot{Kennung: eineKennung(t), Name: "Nachbar", Port: 4711}
}

func TestEineAnkuendigungGehtUeberJedeTauglicheSchnittstelle(t *testing.T) {
	zweiSchnittstellen(t)

	rufe := mitServern(t, func(*mdns.Config) (niederlegbar, error) {
		return nil, nil
	})

	if _, fehler := Melde(einAngebot(t)); fehler != nil {
		t.Fatalf("die Ankündigung scheiterte: %v", fehler)
	}

	if len(*rufe) != 2 {
		t.Fatalf("es gab %d Horchposten, erwartet 2 — einen je Schnittstelle",
			len(*rufe))
	}

	namen := map[string]bool{}
	for _, ruf := range *rufe {
		if ruf.Iface == nil {
			t.Fatal("ein Horchposten nannte keine Schnittstelle")
		}

		namen[ruf.Iface.Name] = true
	}

	for _, name := range []string{"en0", "en5"} {
		if !namen[name] {
			t.Errorf("auf %q wurde nicht gehorcht", name)
		}
	}
}

// **Wer eine Frage vom WLAN nicht hört, beantwortet sie auch nicht.**
//
// `net.ListenMulticastUDP("udp4", nil, …)` horcht nur auf der vom System
// zugewiesenen Schnittstelle — deshalb braucht jede einen eigenen Posten.
func TestEinGescheiterterHorchpostenNimmtDieAnderenNichtMit(t *testing.T) {
	zweiSchnittstellen(t)

	mitServern(t, func(config *mdns.Config) (niederlegbar, error) {
		if config.Iface != nil && config.Iface.Name == "en5" {
			return nil, errors.New("kein Listener")
		}

		return nil, nil
	})

	if _, fehler := Melde(einAngebot(t)); fehler != nil {
		t.Errorf("die Ankündigung scheiterte an einer Schnittstelle: %v",
			fehler)
	}
}

// **Scheitern alle, scheitert die Ankündigung** — der Steuerlauf (§10).
func TestScheiternAlleHorchpostenScheitertDieAnkuendigung(t *testing.T) {
	zweiSchnittstellen(t)

	mitServern(t, func(*mdns.Config) (niederlegbar, error) {
		return nil, errors.New("kein Listener")
	})

	if _, fehler := Melde(einAngebot(t)); fehler == nil {
		t.Error("die Ankündigung gelang, obwohl jeder Posten scheiterte")
	}
}

// ohneWahl hängt eine klagende oder leere Wahl an die Naht.
func ohneWahl(t *testing.T, fehler error) {
	t.Helper()

	vorher := tauglicheLesen
	tauglicheLesen = func() ([]net.Interface, error) { return nil, fehler }

	t.Cleanup(func() { tauglicheLesen = vorher })
}

// **Klagt die Wahl, klagt die Suche.**
//
// **Und sie rät nicht.** Wer ohne Liste der Schnittstellen einfach das
// Betriebssystem fragte, machte aus einem Fehler eine Vermutung — und die
// Suche fände vielleicht nichts, ohne dass jemand erführe, warum.
func TestKlagtDieWahlKlagtDieSuche(t *testing.T) {
	ohneWahl(t, errors.New("kein Zugriff"))

	if _, fehler := Suche(context.Background(), Frage{}); fehler == nil {
		t.Error("die Suche gelang, obwohl die Wahl klagte")
	}
}

// **Klagt die Wahl, klagt die Ankündigung** — aus demselben Grund.
func TestKlagtDieWahlKlagtDieAnkuendigung(t *testing.T) {
	ohneWahl(t, errors.New("kein Zugriff"))

	if _, fehler := Melde(einAngebot(t)); fehler == nil {
		t.Error("die Ankündigung gelang, obwohl die Wahl klagte")
	}
}

// **Ohne taugliche Schnittstelle kündigt sie an wie bisher.**
//
// Das ist, was vorher immer galt — und auf einem Rechner ohne Netz die
// einzige Antwort, die nicht lügt.
func TestOhneTauglicheSchnittstelleKuendigtSieWieBisherAn(t *testing.T) {
	ohneWahl(t, nil)

	vorherigeAdressen := adresslisteVon
	adresslisteVon = func(*net.Interface) ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.IPv4(192, 168, 1, 2)}}, nil
	}

	t.Cleanup(func() { adresslisteVon = vorherigeAdressen })

	rufe := mitServern(t, func(*mdns.Config) (niederlegbar, error) {
		return nil, nil
	})

	if _, fehler := Melde(einAngebot(t)); fehler != nil {
		t.Fatalf("die Ankündigung scheiterte: %v", fehler)
	}

	if len(*rufe) != 1 {
		t.Fatalf("es gab %d Horchposten, erwartet 1", len(*rufe))
	}

	if (*rufe)[0].Iface != nil {
		t.Error("der Posten nannte eine Schnittstelle, obwohl keine taugt")
	}
}
