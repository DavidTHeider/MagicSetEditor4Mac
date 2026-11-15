package app

// **Welches Zeugnis ein Dienst vorzeigt** — NET-03.AK-01 und AK-03
// (#1707).
//
// Drei Wege und ein Nein:
//
//  1. ohne Angaben das Zeugnis auf der Peeridentität, wie seit N6;
//  2. mit Zertifikat und Schlüssel in den Einstellungen das eigene;
//  3. eine Flagge schlägt die Einstellung;
//  4. **eine halbe Angabe startet nichts** — sie fiele sonst still auf die
//     Peeridentität zurück, und wer eine PKI eingetragen hat, bekäme ein
//     Zeugnis, das seine Klienten abweisen.

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineHausPKI legt eine Wurzel und ein Blatt für `name` ab und gibt die
// vier Pfade — die Kette ist leer, das Blatt stammt direkt von der Wurzel.
func eineHausPKI(t *testing.T, name string) leitung.Zeugnisdateien {
	t.Helper()

	ordner := t.TempDir()
	dateien := leitung.Zeugnisdateien{
		Zertifikat: filepath.Join(ordner, "dienst.pem"),
		Schluessel: filepath.Join(ordner, "schluessel.pem"),
		Anker:      filepath.Join(ordner, "anker.pem"),
	}

	wurzelschluessel := muss[*ecdsa.PrivateKey](t)(ecdsa.GenerateKey(elliptic.P256(), rand.Reader))
	wurzelvorlage := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: name},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}

	wurzelroh, fehler := x509.CreateCertificate(rand.Reader, wurzelvorlage, wurzelvorlage,
		&wurzelschluessel.PublicKey, wurzelschluessel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	wurzel := muss[*x509.Certificate](t)(x509.ParseCertificate(wurzelroh))

	blattschluessel := muss[*ecdsa.PrivateKey](t)(ecdsa.GenerateKey(elliptic.P256(), rand.Reader))
	blattroh, fehler := x509.CreateCertificate(rand.Reader, &x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "mse-rest"},
		DNSNames: []string{"dienst.haus.test"}, NotBefore: time.Now().Add(-time.Hour),
		NotAfter: time.Now().Add(time.Hour), ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}, wurzel, &blattschluessel.PublicKey, wurzelschluessel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	schluesselroh := muss[[]byte](t)(x509.MarshalPKCS8PrivateKey(blattschluessel))

	for pfad, block := range map[string]*pem.Block{
		dateien.Zertifikat: {Type: "CERTIFICATE", Bytes: blattroh},
		dateien.Schluessel: {Type: "PRIVATE KEY", Bytes: schluesselroh},
		dateien.Anker:      {Type: "CERTIFICATE", Bytes: wurzelroh},
	} {
		if fehler := os.WriteFile(pfad, pem.EncodeToMemory(block), 0o600); fehler != nil {
			t.Fatal(fehler)
		}
	}

	return dateien
}

// muss bricht den Lauf ab, wenn ein Schritt der Vorrichtung scheitert.
func muss[W any](t *testing.T) func(W, error) W {
	t.Helper()

	return func(wert W, fehler error) W {
		t.Helper()

		if fehler != nil {
			t.Fatal(fehler)
		}

		return wert
	}
}

// einDienst ist eine Dienstfläche mit Identität und eingesetzter Naht.
func einDienst(t *testing.T) (anwendung *App, setze func(schluessel, wert string)) {
	t.Helper()

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)

	anwendung = DienstflaecheUeber(db).MitPeerbund(&peerprobebund{})
	if antwort := anwendung.PeeridentitaetSicherstellen(anwendung.imProfil(ctx)); antwort.Fehler != nil {
		t.Fatalf("Identität: %+v", antwort.Fehler)
	}

	setze = func(schluessel, wert string) {
		t.Helper()

		if antwort := anwendung.EinstellungSetzen(ctx, schluessel, wert); antwort.Fehler != nil {
			t.Fatalf("%s setzen: %+v", schluessel, antwort.Fehler)
		}
	}

	return anwendung, setze
}

func stumm(error) {}

func TestOhneAngabenZeigtDerDienstDieIdentitaet(t *testing.T) {
	t.Parallel()

	anwendung, _ := einDienst(t)
	ctx := riegelkontext(t)

	einstellung, angabe, fehler := anwendung.Dienstzeugnis(ctx, nil, leitung.Zeugnisdateien{}, stumm)
	if fehler != nil {
		t.Fatalf("Dienstzeugnis: %v", fehler)
	}

	ich := muss[peer.Identitaet](t)(anwendung.Ausweis(ctx))
	if angabe.Fingerabdruck != ich.Fingerabdruck() || len(einstellung.Certificates) != 1 {
		t.Errorf("die Angabe ist %+v", angabe)
	}

	if !angabe.GueltigBis.IsZero() || angabe.Aussteller != "" {
		t.Errorf("das Zeugnis der Identität nennt eine Stelle: %+v", angabe)
	}
}

func TestDieEinstellungenBringenDasEigeneZeugnis(t *testing.T) {
	t.Parallel()

	anwendung, setze := einDienst(t)
	pki := eineHausPKI(t, "Haus-Wurzel")

	setze(domain.DienstTLSZertifikat, pki.Zertifikat)
	setze(domain.DienstTLSSchluessel, pki.Schluessel)
	setze(domain.DienstTLSAnker, pki.Anker)

	einstellung, angabe, fehler := anwendung.Dienstzeugnis(
		riegelkontext(t), nil, leitung.Zeugnisdateien{}, stumm)
	if fehler != nil {
		t.Fatalf("Dienstzeugnis: %v", fehler)
	}

	if angabe.Aussteller != "Haus-Wurzel" || angabe.GueltigBis.IsZero() ||
		len(angabe.Namen) != 1 || einstellung.GetCertificate == nil {
		t.Errorf("die Angabe ist %+v", angabe)
	}
}

// **„Nur lokal" lässt das eigene Zeugnis stehen** — Folgebefund F2 des
// Lastlaufs QA-08 (#2004).
//
// Die Richtlinie schaltet die fremden Dienste ab, nach denen dieses Programm
// fragt; die vier Pfade sind keiner davon. Bis zum 2026-10-02 wurden sie
// trotzdem zu „aus", und ein Dienst jenseits der Schleife wollte ein
// Zertifikat namens „aus" laden und startete nicht.
func TestNurLokalLaesstDasEigeneZeugnisStehen(t *testing.T) {
	t.Parallel()

	anwendung, setze := einDienst(t)
	pki := eineHausPKI(t, "Haus-Wurzel")

	setze(domain.DienstTLSZertifikat, pki.Zertifikat)
	setze(domain.DienstTLSSchluessel, pki.Schluessel)
	setze(domain.DienstTLSAnker, pki.Anker)
	setze(domain.RichtlinieNurLokal, domain.NurLokalAn)

	_, angabe, fehler := anwendung.Dienstzeugnis(
		riegelkontext(t), nil, leitung.Zeugnisdateien{}, stumm)
	if fehler != nil {
		t.Fatalf("unter »nur lokal« startet der Dienst nicht: %v", fehler)
	}

	if angabe.Aussteller != "Haus-Wurzel" {
		t.Errorf("es gilt das Zeugnis von %q", angabe.Aussteller)
	}
}

// **Eine Flagge schlägt die Einstellung**, Feld für Feld.
func TestEineFlaggeSchlaegtDieEinstellung(t *testing.T) {
	t.Parallel()

	anwendung, setze := einDienst(t)
	eingestellt := eineHausPKI(t, "Eingestellt")
	geflaggt := eineHausPKI(t, "Geflaggt")

	setze(domain.DienstTLSZertifikat, eingestellt.Zertifikat)
	setze(domain.DienstTLSSchluessel, eingestellt.Schluessel)
	setze(domain.DienstTLSAnker, eingestellt.Anker)

	_, angabe, fehler := anwendung.Dienstzeugnis(riegelkontext(t), nil, geflaggt, stumm)
	if fehler != nil {
		t.Fatalf("Dienstzeugnis: %v", fehler)
	}

	if angabe.Aussteller != "Geflaggt" {
		t.Errorf("es gilt %q", angabe.Aussteller)
	}
}

func TestEineHalbeAngabeStartetNichts(t *testing.T) {
	t.Parallel()

	anwendung, setze := einDienst(t)
	pki := eineHausPKI(t, "Haus-Wurzel")

	setze(domain.DienstTLSAnker, pki.Anker)

	_, _, fehler := anwendung.Dienstzeugnis(riegelkontext(t), nil, leitung.Zeugnisdateien{}, stumm)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("mit nur einem Anker: %v", fehler)
	}
}

// **Ein eigenes Zeugnis, das die Prüfung nicht besteht**, startet nichts.
func TestEinUngueltigesEigenesZeugnisStartetNichts(t *testing.T) {
	t.Parallel()

	anwendung, _ := einDienst(t)
	pki := eineHausPKI(t, "Haus-Wurzel")
	pki.Anker = ""

	if _, _, fehler := anwendung.Dienstzeugnis(riegelkontext(t), nil, pki, stumm); fehler == nil {
		t.Error("eine Kette ohne Anker kam durch")
	}
}

// **Ohne Identität und ohne eigene PKI** gibt es kein Zeugnis — und damit
// keinen Port jenseits der Schleife.
func TestOhneIdentitaetGibtEsKeinZeugnis(t *testing.T) {
	t.Parallel()

	anwendung := DienstflaecheUeber(bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	if _, _, fehler := anwendung.Dienstzeugnis(
		riegelkontext(t), nil, leitung.Zeugnisdateien{}, stumm); fehler == nil {
		t.Error("ohne Identität entstand ein Zeugnis")
	}
}

// klagendeEinstellungen lassen sich nicht lesen.
type klagendeEinstellungen struct{ netzeinstellungsprobe }

func (klagendeEinstellungen) Stand(context.Context) (map[string]service.Wahl, error) {
	return nil, errors.New("der Bestand klemmt")
}

// **Was sich nicht lesen lässt, startet keinen Dienst** — weder die
// Einstellungen noch ein Zeugnis, das sich nicht bauen lässt.
func TestWasNichtGehtStartetKeinenDienst(t *testing.T) {
	t.Parallel()

	klemmt := DienstflaecheUeber(bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})
	klemmt.MitEinstellungen(klagendeEinstellungen{})

	if _, _, fehler := klemmt.Dienstzeugnis(
		riegelkontext(t), nil, leitung.Zeugnisdateien{}, stumm); fehler == nil {
		t.Error("ohne lesbare Einstellungen entstand ein Zeugnis")
	}

	anwendung, _ := einDienst(t)
	anwendung.peerlage.serverzeugnis = func(peer.Identitaet) (*tls.Config, error) {
		return nil, errors.New("kein Zufall")
	}

	if _, _, fehler := anwendung.Dienstzeugnis(
		riegelkontext(t), nil, leitung.Zeugnisdateien{}, stumm); fehler == nil {
		t.Error("ein Zeugnis, das sich nicht bauen ließ, kam durch")
	}
}

// **Die vier Flaggen kommen an**, und ohne sie bleibt alles leer.
func TestDieZeugnisflaggenKommenAn(t *testing.T) {
	t.Parallel()

	fahne := flag.NewFlagSet("probe", flag.ContinueOnError)
	lesen := Zeugnisflaggen(fahne)

	if fehler := fahne.Parse([]string{
		"--zeugnis", "/z.pem", "--schluessel", "/s.pem", "--kette", "/k.pem", "--anker", "/a.pem",
	}); fehler != nil {
		t.Fatal(fehler)
	}

	if gelesen := lesen(); gelesen != (leitung.Zeugnisdateien{
		Zertifikat: "/z.pem", Schluessel: "/s.pem", Kette: "/k.pem", Anker: "/a.pem",
	}) {
		t.Errorf("gelesen: %+v", gelesen)
	}
}
