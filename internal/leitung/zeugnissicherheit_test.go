package leitung_test

// **Was ein ausgegebenes Zeugnis zusagt** — Befunde Z4 und Z5 des
// Zweitblicks über Phase 6.7 (NET-Teil).
//
//   - **Z4**: ein Zeugnis, das an TLS ging, ändert sich nicht mehr — Go
//     1.27.1 verlangt es (`crypto/tls/common.go`, `GetCertificate`), und ein
//     langsamer Handschlag liest Zertifikat und Schlüssel getrennt.
//   - **Z5**: eine Wiederaufnahme überspringt `GetCertificate`
//     (`handshake_server_tls13.go`) und damit Nachladen und Ablaufsperre —
//     deshalb gibt es keine.

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

func TestEinVorgezeigtesZeugnisBleibtWasEsWar(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	werk, uhr, _ := einWerk(t, pki)

	vorher, fehler := werk.Einstellung().GetCertificate(&tls.ClientHelloInfo{})
	if fehler != nil {
		t.Fatal(fehler)
	}

	neu := neuesPaar(t, pki, time.Now().Add(6*time.Hour))
	uhr.weiter(leitung.Nachsehfrist)

	nachher, fehler := werk.Einstellung().GetCertificate(&tls.ClientHelloInfo{})
	if fehler != nil || !nachher.Leaf.Equal(neu) {
		t.Fatalf("nach dem Tausch gilt nicht das neue: %v", fehler)
	}

	if !vorher.Leaf.Equal(pki.blatt) {
		t.Error("das schon ausgegebene Zeugnis wurde nachträglich verändert")
	}
}

func TestEinDienstNimmtKeineSitzungWiederAuf(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	werk, _, _ := einWerk(t, pki)

	lauscher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = lauscher.Close() }) //nolint:errcheck // nur ein Lauf

	// **Eine Einstellung für alle Verbindungen**, wie ein Dienst sie hält —
	// sonst hätte jede ihre eigenen Ticketschlüssel, und keine nähme ein
	// fremdes Ticket an.
	dienst := werk.Einstellung()

	go func() {
		for {
			verbindung, annahmefehler := lauscher.Accept()
			if annahmefehler != nil {
				return
			}

			server := tls.Server(verbindung, dienst)
			if server.HandshakeContext(context.Background()) == nil {
				_, _ = server.Write([]byte("x")) //nolint:errcheck // gemessen wird am Klienten
			}

			_ = server.Close() //nolint:errcheck // nur ein Lauf
		}
	}()

	anker := x509.NewCertPool()
	anker.AddCert(pki.wurzel.zertifikat)

	einstellung := &tls.Config{
		RootCAs: anker, ServerName: "dienst.beispiel.test", MinVersion: tls.VersionTLS13,
		ClientSessionCache: tls.NewLRUClientSessionCache(4),
	}

	for runde := range 2 {
		klient, waehlfehler := (&tls.Dialer{Config: einstellung}).DialContext(
			t.Context(), "tcp", lauscher.Addr().String())
		if waehlfehler != nil {
			t.Fatalf("Runde %d: %v", runde, waehlfehler)
		}

		// **Gelesen wird, damit das Ticket ankommt** — unter TLS 1.3 kommt es
		// nach dem Handschlag.
		_, _ = klient.Read(make([]byte, 1)) //nolint:errcheck // nur für das Ticket

		if gesichert, istTLS := klient.(*tls.Conn); !istTLS || gesichert.ConnectionState().DidResume {
			t.Errorf("Runde %d wurde wiederaufgenommen — ohne Nachladen und Ablaufsperre", runde)
		}

		_ = klient.Close() //nolint:errcheck // nur ein Lauf
	}
}
