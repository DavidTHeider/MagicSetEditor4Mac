package leitung_test

// **Rotation mit sicheren Zuständen** — NET-03.AK-02
// (#1707):
// „Abgelaufene Zertifikate, Rotation und misslungener Reload haben
// definierte sichere Zustände."
//
// Die drei Zustände, und jeder hat hier seinen Lauf:
//
//  1. ein neues Paar, das die Prüfung besteht, gilt ab dem nächsten
//     Handschlag;
//  2. eines, das sie nicht besteht, gilt nicht — das alte bleibt, und der
//     Befund wird gemeldet;
//  3. ein abgelaufenes wird nie vorgezeigt — der Handschlag scheitert.

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// eineUhr ist eine Uhr, die nur auf Zuruf weitergeht.
type eineUhr struct {
	schloss sync.Mutex
	jetzt   time.Time
}

func (u *eineUhr) lies() time.Time {
	u.schloss.Lock()
	defer u.schloss.Unlock()

	return u.jetzt
}

func (u *eineUhr) weiter(dauer time.Duration) {
	u.schloss.Lock()
	defer u.schloss.Unlock()

	u.jetzt = u.jetzt.Add(dauer)
}

// gemeldet sammelt die Befunde des Werks.
type gemeldet struct {
	schloss sync.Mutex
	fehler  []error
}

func (g *gemeldet) melde(fehler error) {
	g.schloss.Lock()
	defer g.schloss.Unlock()

	g.fehler = append(g.fehler, fehler)
}

func (g *gemeldet) zahl() int {
	g.schloss.Lock()
	defer g.schloss.Unlock()

	return len(g.fehler)
}

func einWerk(t *testing.T, pki einePKI) (*leitung.Zeugniswerk, *eineUhr, *gemeldet) {
	t.Helper()

	uhr := &eineUhr{jetzt: time.Now()}
	befunde := &gemeldet{}

	werk, fehler := leitung.NeuesZeugniswerk(pki.dateien, nil, uhr.lies, befunde.melde)
	if fehler != nil {
		t.Fatalf("NeuesZeugniswerk: %v", fehler)
	}

	return werk, uhr, befunde
}

func vorgezeigt(t *testing.T, werk *leitung.Zeugniswerk) (*x509.Certificate, error) {
	t.Helper()

	zeugnis, fehler := werk.Einstellung().GetCertificate(&tls.ClientHelloInfo{})
	if fehler != nil {
		return nil, fehler
	}

	return zeugnis.Leaf, nil
}

// neuesPaar legt ein frisches Blatt derselben Stelle ab — mit anderer Zeit,
// damit der Blick auf die Dateien es sieht.
func neuesPaar(t *testing.T, pki einePKI, bis time.Time) *x509.Certificate {
	t.Helper()

	blatt, schluessel := einBlatt(t, blattwunsch{
		namen: []string{"dienst.beispiel.test"}, von: time.Now().Add(-time.Hour),
		bis: bis, aussteller: pki.zwischen,
	})

	schreibe(t, pki.dateien.Zertifikat, alsPEM(blatt), 0o644)
	schreibe(t, pki.dateien.Schluessel, schluesselAlsPEM(t, schluessel), 0o600)

	spaeter := time.Now().Add(time.Minute)
	for _, pfad := range []string{pki.dateien.Zertifikat, pki.dateien.Schluessel} {
		if fehler := os.Chtimes(pfad, spaeter, spaeter); fehler != nil {
			t.Fatal(fehler)
		}
	}

	return blatt
}

func TestEinNeuesPaarGiltAbDemNaechstenBlick(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	werk, uhr, befunde := einWerk(t, pki)

	neu := neuesPaar(t, pki, time.Now().Add(6*time.Hour))

	// **Vor der Frist bleibt das alte** — ein `stat` je Handschlag wäre
	// unter Last eine Platte voller Anfragen.
	uhr.weiter(time.Second)

	if blatt, fehler := vorgezeigt(t, werk); fehler != nil || !blatt.Equal(pki.blatt) {
		t.Error("vor der Frist wurde schon getauscht")
	}

	uhr.weiter(leitung.Nachsehfrist)

	if blatt, fehler := vorgezeigt(t, werk); fehler != nil || !blatt.Equal(neu) {
		t.Error("nach der Frist gilt noch das alte")
	}

	if werk.Befund().GueltigBis != neu.NotAfter || befunde.zahl() != 0 {
		t.Errorf("Befund %+v, gemeldet %d", werk.Befund(), befunde.zahl())
	}
}

func TestEinMisslungenesNachladenBehaeltDasAlte(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	werk, uhr, befunde := einWerk(t, pki)

	// **Ein Zertifikat ohne passenden Schlüssel** — wie mitten in einem
	// Austausch, der die zweite Datei noch nicht geschrieben hat.
	blatt, _ := einBlatt(t, blattwunsch{
		namen: []string{"dienst.beispiel.test"}, von: time.Now().Add(-time.Hour),
		bis: time.Now().Add(time.Hour), aussteller: pki.zwischen,
	})
	schreibe(t, pki.dateien.Zertifikat, alsPEM(blatt), 0o644)

	spaeter := time.Now().Add(time.Minute)
	if fehler := os.Chtimes(pki.dateien.Zertifikat, spaeter, spaeter); fehler != nil {
		t.Fatal(fehler)
	}

	uhr.weiter(leitung.Nachsehfrist)

	if gezeigt, fehler := vorgezeigt(t, werk); fehler != nil || !gezeigt.Equal(pki.blatt) {
		t.Error("ein Paar, das nicht passt, wurde vorgezeigt")
	}

	if befunde.zahl() != 1 {
		t.Errorf("gemeldet: %d", befunde.zahl())
	}

	// **Und derselbe kaputte Stand wird nicht bei jedem Blick neu gemeldet.**
	uhr.weiter(leitung.Nachsehfrist)

	if _, fehler := vorgezeigt(t, werk); fehler != nil || befunde.zahl() != 1 {
		t.Errorf("beim zweiten Blick: %v, gemeldet %d", fehler, befunde.zahl())
	}
}

func TestVerschwundeneDateienBehaltenDasAlte(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	werk, uhr, befunde := einWerk(t, pki)

	if fehler := os.Remove(pki.dateien.Zertifikat); fehler != nil {
		t.Fatal(fehler)
	}

	uhr.weiter(leitung.Nachsehfrist)

	if gezeigt, fehler := vorgezeigt(t, werk); fehler != nil || !gezeigt.Equal(pki.blatt) {
		t.Errorf("ohne Datei: %v", fehler)
	}

	if befunde.zahl() != 1 {
		t.Errorf("gemeldet: %d", befunde.zahl())
	}
}

func TestEinAbgelaufenesZeugnisWirdNieVorgezeigt(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	werk, uhr, _ := einWerk(t, pki)

	uhr.weiter(time.Until(pki.blatt.NotAfter) + time.Minute)

	if _, fehler := vorgezeigt(t, werk); fehler == nil {
		t.Error("ein abgelaufenes Zeugnis wurde vorgezeigt")
	}
}

func TestEinUngueltigesZeugnisStartetKeinWerk(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	pki.dateien.Anker = ""

	if _, fehler := leitung.NeuesZeugniswerk(pki.dateien, nil, time.Now, func(error) {}); fehler == nil {
		t.Error("ein Werk entstand mit einer Kette ohne Anker")
	}

	pki = eineGueltigePKI(t)
	pki.dateien.Kette += ".fort"

	if _, fehler := leitung.NeuesZeugniswerk(pki.dateien, nil, time.Now, func(error) {}); fehler == nil {
		t.Error("ein Werk entstand ohne seine Kette")
	}
}

// **Ein echter Handschlag mit einem prüfenden Klienten** — ohne
// `InsecureSkipVerify`, gegen den Anker des Hauses und den Namen im Blatt.
func TestEinPruefenderKlientKommtDurch(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	werk, _, _ := einWerk(t, pki)

	lauscher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = lauscher.Close() }) //nolint:errcheck // nur ein Lauf

	go func() {
		verbindung, annahmefehler := lauscher.Accept()
		if annahmefehler != nil {
			return
		}

		server := tls.Server(verbindung, werk.Einstellung())
		_ = server.HandshakeContext(context.Background()) //nolint:errcheck // gemessen wird am Klienten
		_ = server.Close()                                //nolint:errcheck // nur ein Lauf
	}()

	anker := x509.NewCertPool()
	anker.AddCert(pki.wurzel.zertifikat)

	klient, fehler := (&tls.Dialer{Config: &tls.Config{
		RootCAs: anker, ServerName: "dienst.beispiel.test", MinVersion: tls.VersionTLS13,
	}}).DialContext(t.Context(), "tcp", lauscher.Addr().String())
	if fehler != nil {
		t.Fatalf("der Handschlag scheiterte: %v", fehler)
	}

	_ = klient.Close() //nolint:errcheck // nur ein Lauf
}
