package leitung

// **Unter welcher Adresse jemand anderes ankommt** — der Befund aus der
// Sichtprüfung vom 2026-09-14.
//
// # Was hier gemessen wird
//
// **Dass die Bindeadresse nicht die Antwort ist.** Ein Lauscher meldet, wo er
// gebunden hat — `[::]:41234` für „jede Schnittstelle", `127.0.0.1:41234` für
// die Rückschleife. Beides ist wahr und beides nützt einem Gegenüber nichts:
// das erste ist keine Adresse, das zweite zeigt auf den eigenen Rechner.
//
// Gemessen am 2026-09-14 auf diesem Rechner:
//
//	:41234          -> [::]:41234
//	127.0.0.1:41234 -> 127.0.0.1:41234
//
// Was ein Gegenüber braucht, stand daneben und wurde nie gezeigt:
// `192.168.178.51:41234`.

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// probeadressen ist eine Schnittstellenliste, wie dieser Rechner sie hat.
//
// **Aus einer echten Messung und nicht erfunden**: `net.InterfaceAddrs()` gab
// am 2026-09-14 genau diese Gestalten — Rückschleife, Link-Local mit und ohne
// Zone, eine globale IPv6 und die eine, um die es geht.
func probeadressen() ([]net.Addr, error) {
	var liste []net.Addr

	for _, text := range []string{
		"127.0.0.1/8",
		"::1/128",
		"fe80::1/64",
		"fe80::104d:594b:7275:fbf4/64",
		"2001:db8:470f:cd00:1ce7:a492:8132:2b75/64",
		"192.168.178.51/24",
	} {
		adresse, netz, fehler := net.ParseCIDR(text)
		if fehler != nil {
			return nil, fehler
		}

		// **Die Adresse und nicht das Netz.** `ParseCIDR` gibt beides;
		// `net.InterfaceAddrs` liefert ein `*net.IPNet`, dessen `IP` die
		// **Adresse der Schnittstelle** trägt und nicht die Netzadresse.
		netz.IP = adresse

		liste = append(liste, netz)
	}

	return liste, nil
}

func TestEineGebundeneSchnittstelleIstDieAntwort(t *testing.T) {
	mitAdressen(t, probeadressen)

	erreichbar, fehler := ErreichbarUnter(context.Background(), "192.168.178.51:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if len(erreichbar.Adressen) != 1 || erreichbar.Adressen[0] != "192.168.178.51:41234" {
		t.Errorf("erwartet war die gebundene Adresse, kam: %v", erreichbar.Adressen)
	}
}

func TestDieRueckschleifeErreichtNiemandVonAussen(t *testing.T) {
	mitAdressen(t, probeadressen)

	erreichbar, fehler := ErreichbarUnter(context.Background(), "127.0.0.1:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	// **Eine leere Liste und keine erfundene Adresse.** Wer auf der
	// Rückschleife horcht, ist von außen nicht erreichbar — das ist die
	// Auskunft, und sie ist besser als eine Adresse, unter der niemand
	// ankommt.
	if len(erreichbar.Adressen) != 0 {
		t.Errorf("die Rückschleife soll nichts hergeben, kam: %v", erreichbar.Adressen)
	}
}

func TestJedeSchnittstelleGibtDieBrauchbarenHer(t *testing.T) {
	mitAdressen(t, probeadressen)

	erreichbar, fehler := ErreichbarUnter(context.Background(), "[::]:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	erwartet := []string{
		"192.168.178.51:41234",
		"[2001:db8:470f:cd00:1ce7:a492:8132:2b75]:41234",
	}

	if len(erreichbar.Adressen) != len(erwartet) {
		t.Fatalf("erwartet waren %v, kamen %v", erwartet, erreichbar.Adressen)
	}

	for _, wollte := range erwartet {
		if !enthaelt(erreichbar.Adressen, wollte) {
			t.Errorf("%s fehlt in %v", wollte, erreichbar.Adressen)
		}
	}
}

// **Und `0.0.0.0` ist dasselbe wie `::`** — beide heißen „jede
// Schnittstelle", und wer nur eines von beiden erkennte, hätte die Hälfte.
func TestAuchDieVierteUnbestimmteAdresseZaehlt(t *testing.T) {
	mitAdressen(t, probeadressen)

	erreichbar, fehler := ErreichbarUnter(context.Background(), "0.0.0.0:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if !enthaelt(erreichbar.Adressen, "192.168.178.51:41234") {
		t.Errorf("die brauchbare Adresse fehlt: %v", erreichbar.Adressen)
	}
}

// **Ein leerer Wirt ist auch „jede Schnittstelle".** `:41234` kommt so aus
// dem Formular; erst der Lauscher macht `[::]:41234` daraus.
func TestEinLeererWirtIstJedeSchnittstelle(t *testing.T) {
	mitAdressen(t, probeadressen)

	erreichbar, fehler := ErreichbarUnter(context.Background(), ":41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if !enthaelt(erreichbar.Adressen, "192.168.178.51:41234") {
		t.Errorf("die brauchbare Adresse fehlt: %v", erreichbar.Adressen)
	}
}

func TestEineKrummeAdresseIstEinBefund(t *testing.T) {
	mitAdressen(t, probeadressen)

	_, fehler := ErreichbarUnter(context.Background(), "kein Port")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet war ein Befund, kam: %v", fehler)
	}
}

func TestEineUnlesbareSchnittstellenlisteIstEinBefund(t *testing.T) {
	mitAdressen(t, func() ([]net.Addr, error) {
		return nil, errors.New("keine Schnittstellen")
	})

	_, fehler := ErreichbarUnter(context.Background(), "[::]:41234")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet war ein Befund, kam: %v", fehler)
	}
}

// **Und was keine IP-Adresse ist, wird übergangen statt gedeutet.**
// `net.InterfaceAddrs` gibt `net.Addr`, und darin steckt nicht immer ein
// `*net.IPNet` — ein Unix-Socket etwa trägt gar keine.
func TestWasKeineIPIstWirdUebergangen(t *testing.T) {
	mitAdressen(t, func() ([]net.Addr, error) {
		echte, fehler := probeadressen()
		if fehler != nil {
			return nil, fehler
		}

		return append(echte, &net.UnixAddr{Name: "/tmp/nichts"}), nil
	})

	erreichbar, fehler := ErreichbarUnter(context.Background(), "[::]:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if len(erreichbar.Adressen) != 2 {
		t.Errorf("erwartet waren zwei, kamen %v", erreichbar.Adressen)
	}
}

func enthaelt(liste []string, gesucht string) bool {
	for _, eintrag := range liste {
		if eintrag == gesucht {
			return true
		}
	}

	return false
}

// mitAdressen tauscht die Naht für die Dauer eines Laufs.
//
// **Hier und nicht im Produktivcode**: eine exportierte Tauschfunktion
// bräuchte `testing` in einer `.go`-Datei ohne `_test` — und damit zöge
// dieses Paket Testflaggen in jedes Programm, das es einbindet. Dieselbe
// Gestalt wie bei `zufallsquelle` in `zeugnisinnen_intern_test.go`.
func mitAdressen(t *testing.T, statt func() ([]net.Addr, error)) {
	t.Helper()

	vorher := schnittstellenadressen
	schnittstellenadressen = statt

	t.Cleanup(func() { schnittstellenadressen = vorher })

	// **Und der Rechnername dazu.** Ohne ihn fragte jeder dieser Läufe den
	// echten Auflöser dieses Netzes — und wäre anderswo rot (§3). Wer nur
	// Adressen misst, schaltet die Namen ab.
	ohneNamen(t)
}

// ohneNamen lässt die Namensprüfung leer ausgehen.
func ohneNamen(t *testing.T) {
	t.Helper()

	vorherName, vorherLoesen := rechnername, aufloesen
	rechnername = func() (string, error) { return "", nil }
	aufloesen = func(context.Context, string) ([]string, error) {
		t.Error("es wurde aufgelöst, obwohl es keinen Namen gibt")

		return nil, nil
	}

	t.Cleanup(func() { rechnername, aufloesen = vorherName, vorherLoesen })
}

// **Die brauchbarste Adresse steht oben** — der Befund aus der Sichtprüfung
// vom 2026-09-14.
//
// # Was gesehen wurde
//
// Sieben Zeilen unter „Adresse", und `192.168.178.51:41234` — die eine, die
// ein Mensch eintippt — stand an **vierter** Stelle. Sechs der sieben waren
// IPv6, fünf davon temporäre Adressen aus den Privacy Extensions, die
// täglich wechseln.
//
// **Gekürzt wird trotzdem nicht**: was dasteht, ist wahr, und eine Liste, die
// etwas verschweigt, wäre die teurere Hälfte. Geordnet wird.
func TestDieBrauchbarsteAdresseStehtOben(t *testing.T) {
	mitAdressen(t, func() ([]net.Addr, error) {
		return adressenAus(
			// **Absichtlich in der Reihenfolge, die das Betriebssystem gab.**
			"2001:db8:470f:cd00:1ce7:a492:8132:2b75/64",
			"fd12:3456:789a:0:b3:71d7:b68f:dd53/64",
			"192.168.178.51/24",
			"2001:db8:470f:cd00:4075:789a:8f49:ff26/64",
		)
	})

	erreichbar, fehler := ErreichbarUnter(context.Background(), "[::]:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	erwartet := []string{
		"192.168.178.51:41234",
		"[fd12:3456:789a:0:b3:71d7:b68f:dd53]:41234",
		"[2001:db8:470f:cd00:1ce7:a492:8132:2b75]:41234",
		"[2001:db8:470f:cd00:4075:789a:8f49:ff26]:41234",
	}

	if len(erreichbar.Adressen) != len(erwartet) {
		t.Fatalf("erwartet waren %v, kamen %v", erwartet, erreichbar.Adressen)
	}

	for i, wollte := range erwartet {
		if erreichbar.Adressen[i] != wollte {
			t.Errorf("an Stelle %d steht %q, erwartet war %q",
				i+1, erreichbar.Adressen[i], wollte)
		}
	}
}

// adressenAus baut eine Schnittstellenliste aus CIDR-Angaben.
func adressenAus(angaben ...string) ([]net.Addr, error) {
	var liste []net.Addr

	for _, text := range angaben {
		adresse, netz, fehler := net.ParseCIDR(text)
		if fehler != nil {
			return nil, fehler
		}

		netz.IP = adresse
		liste = append(liste, netz)
	}

	return liste, nil
}
