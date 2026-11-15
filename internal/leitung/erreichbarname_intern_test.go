package leitung

// **Der Name, der einen Leasewechsel überlebt** — der Nachtrag zum Befund
// vom 2026-09-14.
//
// # Was hier gemessen wird
//
// **Dass ein Name geprüft wird und nicht geraten.** Gemessen am selben Tag
// auf diesem Rechner:
//
//	os.Hostname()                               = "MacBook-Air-von-Ada.fritz.box"
//	LookupHost("MacBook-Air-von-Ada.fritz.box") = [192.168.178.51 2003:… fdc0:…]
//	LookupHost("…fritz.box.local")              = []  (i/o timeout)
//
// Der Router vergibt einen auflösbaren Namen; ein blind angehängtes `.local`
// ergibt Unsinn und kostet drei Sekunden.

import (
	"context"
	"errors"
	"net"
	"testing"
)

// mitNamen tauscht Rechnername und Auflöser für die Dauer eines Laufs.
func mitNamen(
	t *testing.T,
	name string, fehler error,
	auflösung map[string][]string,
) {
	t.Helper()

	vorherName, vorherLoesen := rechnername, aufloesen

	rechnername = func() (string, error) { return name, fehler }
	aufloesen = func(_ context.Context, gefragt string) ([]string, error) {
		treffer, steht := auflösung[gefragt]
		if !steht {
			return nil, errors.New("i/o timeout")
		}

		return treffer, nil
	}

	t.Cleanup(func() { rechnername, aufloesen = vorherName, vorherLoesen })
}

func TestEinNameDerHierherZeigtWirdGenannt(t *testing.T) {
	mitAdressenOhneNamensabschaltung(t, probeadressen)
	mitNamen(t, "MacBook-Air-von-Ada.fritz.box", nil, map[string][]string{
		"MacBook-Air-von-Ada.fritz.box": {"192.168.178.51"},
	})

	erreichbar, fehler := ErreichbarUnter(context.Background(), "[::]:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if len(erreichbar.Namen) != 1 ||
		erreichbar.Namen[0] != "MacBook-Air-von-Ada.fritz.box:41234" {
		t.Errorf("erwartet war der geprüfte Name, kam: %v", erreichbar.Namen)
	}
}

// **Ein Name, der woandershin zeigt, zählt nicht.**
//
// **Und das ist keine Formsache**: er schickte jemanden zu einem Fremden.
// Ein DHCP-Server, der einen alten Eintrag hält, ist genau dieser Fall.
func TestEinNameDerWoandershinZeigtWirdNichtGenannt(t *testing.T) {
	mitAdressenOhneNamensabschaltung(t, probeadressen)
	mitNamen(t, "fremder.fritz.box", nil, map[string][]string{
		"fremder.fritz.box": {"192.168.178.99"},
	})

	erreichbar, fehler := ErreichbarUnter(context.Background(), "[::]:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if len(erreichbar.Namen) != 0 {
		t.Errorf("erwartet war nichts, kam: %v", erreichbar.Namen)
	}

	// **Und die Adressen bleiben.** Ein Name, der nicht trägt, kostet den
	// Namen und nicht die Auskunft.
	if len(erreichbar.Adressen) == 0 {
		t.Error("die Adressen sind mit dem Namen verschwunden")
	}
}

// **`.local` kommt nur an einen unqualifizierten Namen.**
func TestAnEinenUnqualifiziertenNamenKommtLocal(t *testing.T) {
	mitAdressenOhneNamensabschaltung(t, probeadressen)
	mitNamen(t, "MacBook-Air-von-Ada", nil, map[string][]string{
		"MacBook-Air-von-Ada.local": {"192.168.178.51"},
	})

	erreichbar, fehler := ErreichbarUnter(context.Background(), "[::]:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if len(erreichbar.Namen) != 1 ||
		erreichbar.Namen[0] != "MacBook-Air-von-Ada.local:41234" {
		t.Errorf("erwartet war der .local-Name, kam: %v", erreichbar.Namen)
	}
}

// **Und an einen qualifizierten nicht** — das ist die gemessene Falle.
func TestAnEinenQualifiziertenNamenKommtKeinLocal(t *testing.T) {
	mitAdressenOhneNamensabschaltung(t, probeadressen)

	var gefragt []string

	vorherName, vorherLoesen := rechnername, aufloesen
	rechnername = func() (string, error) {
		return "MacBook-Air-von-Ada.fritz.box", nil
	}
	aufloesen = func(_ context.Context, name string) ([]string, error) {
		gefragt = append(gefragt, name)

		return nil, errors.New("i/o timeout")
	}

	t.Cleanup(func() { rechnername, aufloesen = vorherName, vorherLoesen })

	if _, fehler := ErreichbarUnter(
		context.Background(), "[::]:41234"); fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if len(gefragt) != 1 || gefragt[0] != "MacBook-Air-von-Ada.fritz.box" {
		t.Errorf("gefragt wurde nach %v", gefragt)
	}
}

// **Ein Rechner ohne Namen kostet nur den Namen.**
func TestOhneRechnernamenBleibenDieAdressen(t *testing.T) {
	mitAdressenOhneNamensabschaltung(t, probeadressen)
	mitNamen(t, "", errors.New("kein Name"), nil)

	erreichbar, fehler := ErreichbarUnter(context.Background(), "[::]:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if len(erreichbar.Namen) != 0 {
		t.Errorf("erwartet war nichts, kam: %v", erreichbar.Namen)
	}

	if len(erreichbar.Adressen) == 0 {
		t.Error("die Adressen fehlen")
	}
}

// **Ein gebundener Name geht hinaus, wie er hereinkam.**
//
// Wer einen Rechnernamen gebunden hat, hat ihn gebunden — die Liste dieses
// Rechners zu ziehen wäre eine Antwort auf eine andere Frage.
func TestEinGebundenerNameGehtSoHinausWieErHereinkam(t *testing.T) {
	mitAdressen(t, probeadressen)

	erreichbar, fehler := ErreichbarUnter(
		context.Background(), "mein-mac.fritz.box:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if len(erreichbar.Namen) != 1 ||
		erreichbar.Namen[0] != "mein-mac.fritz.box:41234" {
		t.Errorf("erwartet war der gebundene Name, kam: %v", erreichbar.Namen)
	}
}

// **Und was nicht auflösbar ist, ist auch nicht deutbar.**
//
// `LookupHost` gibt Zeichenketten; dass darin eine IP-Adresse steht, ist die
// Regel und keine Zusage.
func TestWasKeineAdresseIstZaehltNichtAlsTreffer(t *testing.T) {
	mitAdressenOhneNamensabschaltung(t, probeadressen)
	mitNamen(t, "krumm.fritz.box", nil, map[string][]string{
		"krumm.fritz.box": {"das ist keine Adresse"},
	})

	erreichbar, fehler := ErreichbarUnter(context.Background(), "[::]:41234")
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	if len(erreichbar.Namen) != 0 {
		t.Errorf("erwartet war nichts, kam: %v", erreichbar.Namen)
	}
}

// **Und `Leer` sagt, dass von außen niemand ankommt.**
func TestLeerSagtWennNiemandAnkommt(t *testing.T) {
	if !(Erreichbarkeit{}).Leer() {
		t.Error("nichts ist nicht leer")
	}

	if (Erreichbarkeit{Namen: []string{"a:1"}}).Leer() {
		t.Error("ein Name ist leer")
	}

	if (Erreichbarkeit{Adressen: []string{"1.2.3.4:1"}}).Leer() {
		t.Error("eine Adresse ist leer")
	}
}

// mitAdressenOhneNamensabschaltung setzt nur die Adressnaht.
func mitAdressenOhneNamensabschaltung(
	t *testing.T, statt func() ([]net.Addr, error),
) {
	t.Helper()

	vorher := schnittstellenadressen
	schnittstellenadressen = statt

	t.Cleanup(func() { schnittstellenadressen = vorher })
}
