package leitung_test

// **Eine kleine PKI für die Läufe** — NET-03
// (#1707).
//
// Eine Wurzel, eine Zwischenstelle und ein Blatt, alle ECDSA P-256, wie ein
// Haus sie ausstellen würde. **Selbst gebaut und nicht eingecheckt**: ein
// eingechecktes Zertifikat liefe ab, und der Lauf würde an einem Datum rot.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// stelle ist ein Aussteller samt Schlüssel.
type stelle struct {
	zertifikat *x509.Certificate
	schluessel *ecdsa.PrivateKey
}

// blattwunsch beschreibt ein Blatt.
type blattwunsch struct {
	namen      []string
	adressen   []net.IP
	von, bis   time.Time
	aussteller stelle
}

func einSchluessel(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	schluessel, fehler := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return schluessel
}

// seriennummer zählt über parallele Läufe hinweg — deshalb atomar.
var seriennummer atomic.Int64

func ausstellen(
	t *testing.T, vorlage *x509.Certificate, aussteller stelle,
	oeffentlich *ecdsa.PublicKey,
) *x509.Certificate {
	t.Helper()

	vorlage.SerialNumber = big.NewInt(seriennummer.Add(1))

	elternteil, unterschreiber := aussteller.zertifikat, aussteller.schluessel
	if elternteil == nil {
		elternteil = vorlage
	}

	roh, fehler := x509.CreateCertificate(rand.Reader, vorlage, elternteil, oeffentlich, unterschreiber)
	if fehler != nil {
		t.Fatal(fehler)
	}

	zertifikat, fehler := x509.ParseCertificate(roh)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return zertifikat
}

// eineStelle baut eine Wurzel (ohne Aussteller) oder eine Zwischenstelle.
func eineStelle(t *testing.T, name string, ueber *stelle) stelle {
	t.Helper()

	schluessel := einSchluessel(t)
	vorlage := &x509.Certificate{
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}

	aussteller := stelle{schluessel: schluessel}
	if ueber != nil {
		aussteller = *ueber
	}

	return stelle{zertifikat: ausstellen(t, vorlage, aussteller, &schluessel.PublicKey), schluessel: schluessel}
}

// einBlatt stellt ein Serverzeugnis aus und gibt es samt Schlüssel.
func einBlatt(t *testing.T, wunsch blattwunsch) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	schluessel := einSchluessel(t)
	vorlage := &x509.Certificate{
		Subject:     pkix.Name{CommonName: "mse-rest"},
		DNSNames:    wunsch.namen,
		IPAddresses: wunsch.adressen,
		NotBefore:   wunsch.von,
		NotAfter:    wunsch.bis,
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	return ausstellen(t, vorlage, wunsch.aussteller, &schluessel.PublicKey), schluessel
}

func alsPEM(zertifikate ...*x509.Certificate) []byte {
	aus := make([]byte, 0, 1024*len(zertifikate))
	for _, zertifikat := range zertifikate {
		aus = append(aus, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: zertifikat.Raw})...)
	}

	return aus
}

func schluesselAlsPEM(t *testing.T, schluessel *ecdsa.PrivateKey) []byte {
	t.Helper()

	roh, fehler := x509.MarshalPKCS8PrivateKey(schluessel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: roh})
}

func schreibe(t *testing.T, pfad string, inhalt []byte, rechte os.FileMode) {
	t.Helper()

	if fehler := os.WriteFile(pfad, inhalt, rechte); fehler != nil {
		t.Fatal(fehler)
	}

	// **WriteFile setzt die Rechte nur beim Anlegen** — ein Überschreiben
	// behielte die alten.
	if fehler := os.Chmod(pfad, rechte); fehler != nil {
		t.Fatal(fehler)
	}
}

// einePKI ist eine Wurzel, eine Zwischenstelle und ein gültiges Blatt,
// fertig abgelegt: Blatt, Schlüssel, Kette und Anker.
type einePKI struct {
	wurzel, zwischen stelle
	dateien          leitung.Zeugnisdateien
	blatt            *x509.Certificate
}

func eineGueltigePKI(t *testing.T) einePKI {
	t.Helper()

	wurzel := eineStelle(t, "Haus-Wurzel", nil)
	zwischen := eineStelle(t, "Haus-Zwischenstelle", &wurzel)

	blatt, schluessel := einBlatt(t, blattwunsch{
		namen:      []string{"dienst.beispiel.test"},
		adressen:   []net.IP{net.ParseIP("192.0.2.10")},
		von:        time.Now().Add(-time.Hour),
		bis:        time.Now().Add(12 * time.Hour),
		aussteller: zwischen,
	})

	ordner := t.TempDir()
	dateien := leitung.Zeugnisdateien{
		Zertifikat: filepath.Join(ordner, "dienst.pem"),
		Schluessel: filepath.Join(ordner, "dienst-schluessel.pem"),
		Kette:      filepath.Join(ordner, "kette.pem"),
		Anker:      filepath.Join(ordner, "anker.pem"),
	}

	schreibe(t, dateien.Zertifikat, alsPEM(blatt), 0o644)
	schreibe(t, dateien.Schluessel, schluesselAlsPEM(t, schluessel), 0o600)
	schreibe(t, dateien.Kette, alsPEM(zwischen.zertifikat), 0o644)
	schreibe(t, dateien.Anker, alsPEM(wurzel.zertifikat), 0o644)

	return einePKI{wurzel: wurzel, zwischen: zwischen, dateien: dateien, blatt: blatt}
}
