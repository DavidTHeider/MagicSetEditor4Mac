package fernleitung

// **Die wählende Seite** — und vor allem: was **vor** dem Netz entschieden
// wird.

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// probeklient steht für einen tailcat-Klienten.
type probeklient struct {
	strom net.Conn
	klagt error
	zu    bool
}

func (p *probeklient) DialTCPPort(
	_ context.Context, _ uint16,
) (net.Conn, error) {
	if p.klagt != nil {
		return nil, p.klagt
	}

	return p.strom, nil
}

func (p *probeklient) Close() error {
	p.zu = true

	return nil
}

func mitKlient(t *testing.T, gebaut *probeklient) *string {
	t.Helper()

	var gefragt string

	vorher := klientBauen
	klientBauen = func(adresse string) fernklient {
		gefragt = adresse

		return gebaut
	}

	t.Cleanup(func() { klientBauen = vorher })

	return &gefragt
}

// **Eine Einladung aus einer neueren Fassung geht gar nicht erst ins Netz.**
//
// Das ist die Antwort auf die Fassungsfrage aus #1456: der Bruch wird **vor**
// dem Wählen erkannt und nicht als Zeitüberschreitung.
func TestEineNeuereEinladungGehtNichtInsNetz(t *testing.T) {
	gebaut := &probeklient{klagt: errors.New("hätte nicht gewählt werden dürfen")}
	gefragt := mitKlient(t, gebaut)

	_, fehler := Waehle(
		t.Context(), "mse4mac:99:tc1abc", eineIdentitaet(t), "irgendwer")
	if !errors.Is(fehler, domain.ErrEinladungZuNeu) {
		t.Errorf("erwartet war ErrEinladungZuNeu, kam: %v", fehler)
	}

	if *gefragt != "" {
		t.Errorf("es wurde trotzdem gewählt: %q", *gefragt)
	}
}

// **Und eine krumme ebenso wenig.**
func TestEineKrummeEinladungGehtNichtInsNetz(t *testing.T) {
	gefragt := mitKlient(t, &probeklient{})

	if _, fehler := Waehle(
		t.Context(), "irgendwas", eineIdentitaet(t), "irgendwer"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("erwartet war ein Befund, kam: %v", fehler)
	}

	if *gefragt != "" {
		t.Errorf("es wurde trotzdem gewählt: %q", *gefragt)
	}
}

// **Die Adresse aus der Einladung erreicht den Klienten.**
func TestDieAdresseAusDerEinladungErreichtDenKlienten(t *testing.T) {
	gebaut := &probeklient{klagt: errors.New("kein Weg")}
	gefragt := mitKlient(t, gebaut)

	if _, fehler := Waehle(t.Context(), domain.NeueEinladung("tc1xyz"),
		eineIdentitaet(t), "irgendwer"); !errors.Is(
		fehler, domain.ErrPeerNichtErreichbar) || !errors.Is(fehler, domain.ErrNotFound) {
		// **Dieselbe Klasse wie im eigenen Netz** (#2491): ein Gegenüber,
		// das nicht horcht, ist kein gelöschter Eintrag.
		t.Errorf("erwartet war ErrPeerNichtErreichbar, kam: %v", fehler)
	}

	if *gefragt != "tc1xyz" {
		t.Errorf("gewählt wurde %q", *gefragt)
	}

	// **Und der Klient wird niedergelegt** — ein Tunnel, den niemand
	// schließt, hält einen DERP-Platz besetzt.
	if !gebaut.zu {
		t.Error("der Klient steht noch offen")
	}
}

// **Ein Strom, über dem der Handschlag scheitert, schließt den Klienten mit.**
func TestEinGescheiterterHandschlagSchliesstDenKlienten(t *testing.T) {
	hiesig, dortig := net.Pipe()
	t.Cleanup(func() { _ = dortig.Close() }) //nolint:errcheck // dito

	gebaut := &probeklient{strom: hiesig}
	mitKlient(t, gebaut)

	// **Die Gegenseite schweigt** — der Handschlag läuft in seine Frist.
	go func() {
		_, _ = dortig.Read(make([]byte, 1)) //nolint:errcheck // sie schweigt
	}()

	if _, fehler := Waehle(t.Context(), domain.NeueEinladung("tc1xyz"),
		eineIdentitaet(t), "irgendwer"); fehler == nil {
		t.Fatal("der Handschlag kam durch")
	}

	if !gebaut.zu {
		t.Error("der Klient steht noch offen")
	}
}

// **Und der ganze Weg, von der Einladung bis zur Nachricht.**
//
// **Über die Naht und nicht über das Internet**: was tailcat hier liefert, ist
// ein `net.Pipe`. Die Zusage, die damit gehalten wird, ist die des Pakets —
// über dem Tunnel liegt nichts Eigenes.
func TestDerGanzeWegVonDerEinladungBisZurNachricht(t *testing.T) {
	hier := eineIdentitaet(t)
	dort := eineIdentitaet(t)

	hiesig, dortig := net.Pipe()

	mitKlient(t, &probeklient{strom: hiesig})

	fertig := make(chan error, 1)

	go func() {
		genommen, fehler := leitung.NimmStrom(context.Background(), dortig, dort)
		if fehler != nil {
			fertig <- fehler

			return
		}

		fertig <- genommen.Sende(context.Background(), []byte("Mausbraten"))
	}()

	gewaehlt, fehler := Waehle(
		t.Context(), domain.NeueEinladung("tc1xyz"), hier, dort.Kennung())
	if fehler != nil {
		t.Fatalf("Waehle: %v", fehler)
	}

	// **Erst lesen, dann auf die andere Seite warten.**
	//
	// `net.Pipe` ist ungepuffert: die horchende Seite steht im `Sende`,
	// solange niemand liest. Wer hier zuerst `<-fertig` nähme, verklemmte
	// beide — und genau das ist mir beim ersten Anlauf passiert.
	nachricht, fehler := gewaehlt.Empfange(t.Context())
	if fehler != nil {
		t.Fatalf("empfangen: %v", fehler)
	}

	if fehler := <-fertig; fehler != nil {
		t.Fatalf("die horchende Seite: %v", fehler)
	}

	if string(nachricht) != "Mausbraten" {
		t.Errorf("empfangen wurde %q", nachricht)
	}

	// **Und der Kanal bindet an das Schlüsselmaterial** — nicht an eine
	// Adresse, die es über einen Tunnel gar nicht gibt.
	if gewaehlt.Kanal() == "" {
		t.Error("die Leitung hat keinen Kanal")
	}
}
