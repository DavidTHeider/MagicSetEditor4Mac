package leitung_test

// **Ein Ersatzzeugnis trägt jede Adresse, die das erste trug** — Befund 2 des
// Zweitblicks über Phase G
// (#2272).
//
// Beim Start prüfte der Dienst das Zeugnis gegen alle gewählten Adressen,
// das Werk aber bekam nur die erste. Ein Austausch, der eine weitere nicht
// mehr nannte, galt als gelungen — und an ihr wies jeder prüfende Klient den
// Dienst danach ab.

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"os"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

func TestEinErsatzOhneDieZweiteAdresseGiltNicht(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	uhr := &eineUhr{jetzt: time.Now()}
	befunde := &gemeldet{}

	// **Die erste Bindung prüft nichts, die zweite ihre Adresse** — so lag es
	// beim Dienst auf allen Schnittstellen und einer bestimmten daneben.
	werk, fehler := leitung.NeuesZeugniswerk(pki.dateien,
		[]string{"0.0.0.0", "192.0.2.10"}, uhr.lies, befunde.melde)
	if fehler != nil {
		t.Fatalf("NeuesZeugniswerk: %v", fehler)
	}

	// **Das neue Paar nennt 192.0.2.10 nicht** (`neuesPaar`: nur den Namen).
	neuesPaar(t, pki, time.Now().Add(6*time.Hour))
	uhr.weiter(leitung.Nachsehfrist)

	if gezeigt, zeigfehler := vorgezeigt(t, werk); zeigfehler != nil || !gezeigt.Equal(pki.blatt) {
		t.Error("ein Ersatz ohne die zweite Adresse verdrängte das alte Zeugnis")
	}

	if befunde.zahl() != 1 {
		t.Errorf("gemeldet: %d — erwartet der Befund über die zweite Adresse", befunde.zahl())
	}
}

// **Geprüft wird das Paar, das veröffentlicht wird** — neuer Befund des
// Nachlaufs über Phase G. Das Werk las die Dateien je Adresse neu: wurde das
// Zertifikat zwischen zwei Prüfungen getauscht, bestätigte die zweite ein
// anderes als das, das danach galt.
//
// **Ohne `t.Parallel`**: der Griff um das Lesen gilt für das ganze Paket.
func TestEinTauschZwischenZweiPruefungenGiltNicht(t *testing.T) {
	pki := eineGueltigePKI(t)
	uhr := &eineUhr{jetzt: time.Now()}
	befunde := &gemeldet{}

	werk, fehler := leitung.NeuesZeugniswerk(pki.dateien,
		[]string{"0.0.0.0", "192.0.2.10"}, uhr.lies, befunde.melde)
	if fehler != nil {
		t.Fatalf("NeuesZeugniswerk: %v", fehler)
	}

	// **Zwei Blätter auf demselben Schlüssel**: das erste ohne die Adresse,
	// das zweite mit ihr.
	schluessel := einSchluessel(t)
	blatt := func(adressen []net.IP) *x509.Certificate {
		return ausstellen(t, &x509.Certificate{
			Subject:     pkix.Name{CommonName: "mse-rest"},
			DNSNames:    []string{"dienst.beispiel.test"},
			IPAddresses: adressen,
			NotBefore:   time.Now().Add(-time.Hour),
			NotAfter:    time.Now().Add(6 * time.Hour),
			KeyUsage:    x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}, pki.zwischen, &schluessel.PublicKey)
	}
	ohne, mit := blatt(nil), blatt([]net.IP{net.ParseIP("192.0.2.10")})

	schreibe(t, pki.dateien.Zertifikat, alsPEM(ohne), 0o644)
	schreibe(t, pki.dateien.Schluessel, schluesselAlsPEM(t, schluessel), 0o600)

	spaeter := time.Now().Add(time.Minute)
	for _, pfad := range []string{pki.dateien.Zertifikat, pki.dateien.Schluessel} {
		if err := os.Chtimes(pfad, spaeter, spaeter); err != nil {
			t.Fatal(err)
		}
	}

	// **Nach dem ersten Lesen des Zertifikats liegt das andere da.**
	getauscht := false
	leitung.LesenTauschen(t, func(pfad string, lies func(string) ([]byte, error)) ([]byte, error) {
		roh, lesefehler := lies(pfad)
		if pfad == pki.dateien.Zertifikat && !getauscht {
			getauscht = true
			schreibe(t, pki.dateien.Zertifikat, alsPEM(mit), 0o644)
		}

		return roh, lesefehler
	})

	uhr.weiter(leitung.Nachsehfrist)

	if gezeigt, zeigfehler := vorgezeigt(t, werk); zeigfehler != nil || !gezeigt.Equal(pki.blatt) {
		t.Error("ein Zeugnis ohne die zweite Adresse galt, weil ein anderes geprüft wurde")
	}

	if !getauscht {
		t.Fatal("der Griff um das Lesen wurde nie gerufen — der Lauf prüft nichts")
	}
}
