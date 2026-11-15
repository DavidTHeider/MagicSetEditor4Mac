package leitung_test

// **Das Zeugnis eines gewöhnlichen HTTPS-Dienstes** — Werkstück N6 von
// Phase 5.4
// (#1572).
//
// # Wonach hier gefragt wird
//
// Danach, dass es **ohne Zeugnispflicht für die Gegenseite** trägt — das ist
// der Unterschied zu `serverwerk`, und er ist der Grund, warum es diese
// Funktion überhaupt gibt: auf der anderen Seite sitzt ein Skript und kein
// Peer. Und danach, dass das, was vorgezeigt wird, **die Peeridentität** ist:
// wer sich verbindet, vergleicht ihren Fingerabdruck, und sonst gibt es
// nichts zu vergleichen
// ([0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)).

import (
	"crypto/ed25519"
	"crypto/tls"
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Ein Gegenüber ohne eigenes Zeugnis kommt durch — und sieht die Identität.**
func TestDasDienstzeugnisTraegtDiePeeridentitaet(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Identität erzeugen: %v", fehler)
	}

	werk, fehler := leitung.Serverzeugnis(ich)
	if fehler != nil {
		t.Fatalf("Serverzeugnis: %v", fehler)
	}

	if werk.MinVersion != tls.VersionTLS13 {
		t.Errorf("die Untergrenze ist %#x", werk.MinVersion)
	}

	// **Ohne Zeugnispflicht** — das ist die Zusage im Kopf der Funktion, und
	// eine Gegenprobe steckt darin: mit `RequireAnyClientCert` käme der
	// Handschlag darunter nicht zustande.
	if werk.ClientAuth != tls.NoClientCert {
		t.Errorf("die Gegenseite muss ein Zeugnis zeigen: %v", werk.ClientAuth)
	}

	lauscher, fehler := tls.Listen("tcp", "127.0.0.1:0", werk)
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	defer func() { _ = lauscher.Close() }() //nolint:errcheck // der Lauf endet ohnehin

	nimmAn(t, lauscher)

	waehler := &tls.Dialer{Config: &tls.Config{
		InsecureSkipVerify: true, //#nosec G402 -- geprüft wird der Abdruck
		MinVersion:         tls.VersionTLS13,
	}}

	verbindung, fehler := waehler.DialContext(
		t.Context(), "tcp", lauscher.Addr().String())
	if fehler != nil {
		t.Fatalf("den Handschlag führen: %v", fehler)
	}

	defer func() { _ = verbindung.Close() }() //nolint:errcheck // dito

	gesichert, taugt := verbindung.(*tls.Conn)
	if !taugt {
		t.Fatalf("die Verbindung ist %T und kein TLS", verbindung)
	}

	zeugnisse := gesichert.ConnectionState().PeerCertificates
	if len(zeugnisse) != 1 {
		t.Fatalf("es kamen %d Zeugnisse", len(zeugnisse))
	}

	oeffentlich, taugt := zeugnisse[0].PublicKey.(ed25519.PublicKey)
	if !taugt {
		t.Fatalf("das Zeugnis trägt %T", zeugnisse[0].PublicKey)
	}

	// **Und es ist wirklich diese Identität** — sonst zeigte der Dienst
	// einen Fingerabdruck, den das Fenster nicht kennt.
	if !oeffentlich.Equal(ich.Oeffentlich()) {
		t.Error("das Zeugnis trägt einen fremden Schlüssel")
	}

	kennung, fehler := peer.KennungFuer(oeffentlich)
	if fehler != nil {
		t.Fatalf("die Kennung rechnen: %v", fehler)
	}

	if peer.Fingerabdruck(kennung) != ich.Fingerabdruck() {
		t.Errorf("der Abdruck ist %q statt %q",
			peer.Fingerabdruck(kennung), ich.Fingerabdruck())
	}
}

// **Ein Gegenüber mit TLS 1.2 kommt nicht durch.**
//
// **Die Gegenprobe zur Untergrenze**: ohne sie stünde nur die Zahl im Feld,
// und nicht, dass sie greift.
func TestUnterTLS13KommtNiemandDurch(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Identität erzeugen: %v", fehler)
	}

	werk, fehler := leitung.Serverzeugnis(ich)
	if fehler != nil {
		t.Fatalf("Serverzeugnis: %v", fehler)
	}

	lauscher, fehler := tls.Listen("tcp", "127.0.0.1:0", werk)
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	defer func() { _ = lauscher.Close() }() //nolint:errcheck // der Lauf endet ohnehin

	nimmAn(t, lauscher)

	waehler := &tls.Dialer{Config: &tls.Config{
		InsecureSkipVerify: true, //#nosec G402 -- der Handschlag soll scheitern
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS12,
	}}

	verbindung, fehler := waehler.DialContext(
		t.Context(), "tcp", lauscher.Addr().String())
	if fehler == nil {
		_ = verbindung.Close() //nolint:errcheck // er hätte nicht zustande kommen dürfen

		t.Fatal("TLS 1.2 kam durch")
	}
}

// **Und ohne Zufall entsteht keines** — der Fehlerweg steht in
// `zeugnisinnen_intern_test.go`, weil er die Quelle des Pakets tauscht und
// deshalb nicht neben Läufen stehen kann, die parallel laufen.

// nimmAn nimmt eine Verbindung an und **führt den Handschlag zu Ende**.
//
// **Und das ist der Punkt.** `tls.Listen` gibt eine Verbindung heraus, bevor
// irgendetwas ausgehandelt ist; wer sie sofort schließt, bricht dem Gegenüber
// den Handschlag ab — der erste Anlauf endete mit „connection reset by peer"
// und maß nichts.
func nimmAn(t *testing.T, lauscher net.Listener) {
	t.Helper()

	fertig := make(chan struct{})

	t.Cleanup(func() { <-fertig })

	go func() {
		defer close(fertig)

		herein, angenommen := lauscher.Accept()
		if angenommen != nil {
			return
		}

		defer func() { _ = herein.Close() }() //nolint:errcheck // der Lauf endet ohnehin

		gesichert, taugt := herein.(*tls.Conn)
		if !taugt {
			return
		}

		// **Der Fehler wird nicht gemeldet**: in einem der zwei Läufe soll
		// der Handschlag scheitern, und dort ist das die Messung.
		_ = gesichert.HandshakeContext(t.Context()) //nolint:errcheck // siehe darüber

		// **Erst wenn der Lauf fertig ist, wird geschlossen** — sonst
		// verschwindet die Gegenseite mitten in der Prüfung.
		<-t.Context().Done()
	}()
}
