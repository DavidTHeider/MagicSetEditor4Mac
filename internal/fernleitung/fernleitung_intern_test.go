package fernleitung

// **Der Weg über ein fremdes Netz — ohne fremdes Netz** — Werkstück N3 von
// Phase 5.4 (#1456).
//
// # Was hier geprüft wird und was nicht
//
// **Geprüft wird alles diesseits von tailcat**: dass die Einladung ihre
// Fassung trägt und vor dem Wählen geprüft wird, dass ein hereinkommender
// Strom denselben TLS-Weg nimmt wie im eigenen Netz, dass der Horcher auf
// Aufruf entsteht und beim Niederlegen niemanden hängen lässt.
//
// **Nicht geprüft wird tailcat selbst.** Ein Lauf, der wirklich wählte,
// bräuchte ein DERP-Relais im Internet und einen zweiten Rechner. Das steht
// als Handprüfung in `docs/HAERTUNG.md`, und #1456 verlangt es ausdrücklich:
// „Produktivfreigabe bleibt offen, bis eine reale Direkt- und Relaisprobe
// zwischen zwei Rechnern vorliegt."

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// probeserver steht für tailcat, ohne eines zu sein.
type probeserver struct {
	adresse  string
	klagt    error
	annehmen func(net.Conn)
	wege     []Fernweg

	schloss sync.Mutex
	zu      bool
}

func (p *probeserver) Start() error { return p.klagt }

func (p *probeserver) Einladungsadresse() string { return p.adresse }

func (p *probeserver) Wege() []Fernweg { return p.wege }

func (p *probeserver) Close() error {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	p.zu = true

	return nil
}

func (p *probeserver) geschlossen() bool {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	return p.zu
}

// mitServer setzt die Naht für die Dauer eines Laufs.
func mitServer(t *testing.T, gebaut *probeserver) {
	t.Helper()

	vorher := serverBauen
	serverBauen = func(annehmen func(net.Conn)) fernserver {
		gebaut.annehmen = annehmen

		return gebaut
	}

	t.Cleanup(func() { serverBauen = vorher })
}

func eineIdentitaet(t *testing.T) peer.Identitaet {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return ich
}

// **Die Einladung trägt die Adresse des Servers — mit Fassung davor.**
func TestDieEinladungTraegtDieAdresseMitFassung(t *testing.T) {
	gebaut := &probeserver{adresse: "tc1abc"}
	mitServer(t, gebaut)

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	zurueck, fehler := domain.ZerlegeEinladung(horcher.Einladung())
	if fehler != nil {
		t.Fatalf("zerlegen: %v", fehler)
	}

	if zurueck != "tc1abc" {
		t.Errorf("die Adresse kam als %q zurück", zurueck)
	}
}

// **Ein Server, der nicht startet, wird gemeldet.**
func TestEinServerDerNichtStartetWirdGemeldet(t *testing.T) {
	mitServer(t, &probeserver{klagt: errors.New("kein DERP")})

	if _, fehler := Horche(eineIdentitaet(t)); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("erwartet war ein Befund, kam: %v", fehler)
	}
}

// **Ein hereinkommender Strom nimmt denselben TLS-Weg wie im eigenen Netz.**
//
// Das ist die Zusage, für die es dieses Paket gibt: über dem Tunnel liegt
// nichts Eigenes.
func TestEinHereinkommenderStromNimmtDenselbenWeg(t *testing.T) {
	dort := eineIdentitaet(t)
	hier := eineIdentitaet(t)

	gebaut := &probeserver{adresse: "tc1abc"}
	mitServer(t, gebaut)

	horcher, fehler := Horche(dort)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	hiesig, dortig := net.Pipe()
	gebaut.annehmen(dortig)

	fertig := make(chan error, 1)

	go func() {
		genommen, nahmfehler := horcher.Nimm(t.Context())
		if nahmfehler != nil {
			fertig <- nahmfehler

			return
		}

		nachricht, lesefehler := genommen.Empfange(t.Context())
		if lesefehler != nil {
			fertig <- lesefehler

			return
		}

		if string(nachricht) != "Mausbraten" {
			fertig <- errors.New("falsche Nachricht")

			return
		}

		fertig <- nil
	}()

	gewaehlt, fehler := leitung.UeberStrom(
		t.Context(), hiesig, hier, dort.Kennung())
	if fehler != nil {
		t.Fatalf("UeberStrom: %v", fehler)
	}

	if fehler := gewaehlt.Sende(t.Context(), []byte("Mausbraten")); fehler != nil {
		t.Fatalf("senden: %v", fehler)
	}

	if fehler := <-fertig; fehler != nil {
		t.Fatalf("die horchende Seite: %v", fehler)
	}
}

// **Eine volle Schlange hält den Tunnel nicht auf.**
//
// **Und der überzählige Strom wird geschlossen** — ein Gegenüber, das ewig
// wartet, wäre die schlechtere Antwort.
func TestEineVolleSchlangeHaeltDenTunnelNichtAuf(t *testing.T) {
	gebaut := &probeserver{adresse: "tc1abc"}
	mitServer(t, gebaut)

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	for range Schlangenlaenge {
		_, dortig := net.Pipe()
		gebaut.annehmen(dortig)
	}

	hiesig, ueberzaehlig := net.Pipe()
	gebaut.annehmen(ueberzaehlig)

	// **Der überzählige ist zu** — ein Lesen darauf endet sofort statt zu
	// warten.
	if _, fehler := hiesig.Read(make([]byte, 1)); fehler == nil {
		t.Error("der überzählige Strom steht noch offen")
	}
}

// **Ein niedergelegter Horcher lässt niemanden hängen.**
func TestEinNiedergelegterHorcherLaesstNiemandenHaengen(t *testing.T) {
	gebaut := &probeserver{adresse: "tc1abc"}
	mitServer(t, gebaut)

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	fertig := make(chan error, 1)
	go func() {
		_, fehler := horcher.Nimm(context.Background())
		fertig <- fehler
	}()

	if fehler := horcher.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	if fehler := <-fertig; !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("erwartet war ein Ende, kam: %v", fehler)
	}

	if !gebaut.geschlossen() {
		t.Error("der tailcat-Server läuft noch")
	}

	// **Zweimal niederlegen ist erlaubt** — sonst bräuchte jeder Aufrufer
	// einen Merker.
	if fehler := horcher.Schliesse(); fehler != nil {
		t.Errorf("das zweite Niederlegen klagte: %v", fehler)
	}
}

// **Und ein Warten mit Frist endet mit ihr.**
func TestEinWartenMitFristEndetMitIhr(t *testing.T) {
	mitServer(t, &probeserver{adresse: "tc1abc"})

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	ctx, abbrechen := context.WithCancel(context.Background())
	abbrechen()

	if _, fehler := horcher.Nimm(ctx); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("erwartet war ein Ende, kam: %v", fehler)
	}
}

// **Ein Strom, der nach dem Niederlegen ankommt, wird geschlossen.**
//
// **Der Zweig, den nur eine Reihenfolge erreicht**: tailcat ruft aus seiner
// eigenen Goroutine, und zwischen „der Horcher ist zu" und „da kommt noch
// einer" liegt keine Absprache. Wer ihn ausließe, hätte einen offenen Strom,
// den niemand mehr abholt.
func TestEinStromNachDemNiederlegenWirdGeschlossen(t *testing.T) {
	gebaut := &probeserver{adresse: "tc1abc"}
	mitServer(t, gebaut)

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	// **Die Schlange voll machen**, damit der Weg über `<-h.zu` geht und
	// nicht über den Vorgabezweig.
	for range Schlangenlaenge {
		_, dortig := net.Pipe()
		gebaut.annehmen(dortig)
	}

	if fehler := horcher.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	hiesig, spaet := net.Pipe()
	gebaut.annehmen(spaet)

	if _, fehler := hiesig.Read(make([]byte, 1)); fehler == nil {
		t.Error("der späte Strom steht noch offen")
	}
}

// **Ein Server, der beim Niederlegen klagt, wird gemeldet.**
func TestEinKlagenderServerWirdBeimNiederlegenGemeldet(t *testing.T) {
	gebaut := &klagenderServer{probeserver: probeserver{adresse: "tc1abc"}}

	vorher := serverBauen
	serverBauen = func(annehmen func(net.Conn)) fernserver {
		gebaut.annehmen = annehmen

		return gebaut
	}

	t.Cleanup(func() { serverBauen = vorher })

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	if fehler := horcher.Schliesse(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet war ein Befund, kam: %v", fehler)
	}
}

// klagenderServer scheitert beim Niederlegen.
type klagenderServer struct{ probeserver }

func (k *klagenderServer) Close() error {
	return errors.New("der Tunnel lässt sich nicht schließen")
}

// **Die Adresse eines Fernhorchers ist seine Einladung.**
//
// **Der Lauf gehört dazu, weil die Schnittstelle daran hängt**:
// `service.Hereinkommend` fragt nach einer Adresse, und über ein fremdes Netz
// lautet die Antwort nicht „Wirt und Port".
func TestDieAdresseEinesFernhorchersIstSeineEinladung(t *testing.T) {
	mitServer(t, &probeserver{adresse: "tc1abc"})

	horcher, fehler := Horche(eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	if horcher.Adresse() != string(horcher.Einladung()) {
		t.Errorf("Adresse %q, Einladung %q",
			horcher.Adresse(), horcher.Einladung())
	}

	if horcher.Adresse() != "mse4mac:1:tc1abc" {
		t.Errorf("die Adresse ist %q", horcher.Adresse())
	}
}

// **Der Horcher sagt, wie die Verbindungen gehen** — NET-02.AK-03
// (#1706):
// „Direkter Weg, Relay und begründeter Ausfall sind nachvollziehbar".
func TestDerHorcherNenntDieWege(t *testing.T) {
	server := &probeserver{adresse: "tailcat:probe", wege: []Fernweg{
		{Direkt: "203.0.113.7:41641"},
		{Relais: "fra"},
	}}

	vorher := serverBauen
	serverBauen = func(annehmen func(net.Conn)) fernserver {
		server.annehmen = annehmen

		return server
	}

	t.Cleanup(func() { serverBauen = vorher })

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	horcher, fehler := Horche(ich)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Schliesse() }) //nolint:errcheck // nur ein Lauf

	wege := horcher.Wege()
	if len(wege) != 2 || wege[0].Direkt != "203.0.113.7:41641" || wege[1].Relais != "fra" {
		t.Errorf("Wege: %+v", wege)
	}
}
